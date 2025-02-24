package core

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/EvgenyRomanov/sql-migrator/internal/database"
	_ "github.com/EvgenyRomanov/sql-migrator/internal/database/postgres"
	"github.com/EvgenyRomanov/sql-migrator/internal/logger"
	"github.com/EvgenyRomanov/sql-migrator/internal/parser"
)

var (
	ErrNoCurrentVersion      = errors.New("no current version found. Please check your DB state")
	ErrNoAvailableMigrations = errors.New("no available migrations found")
	ErrAlreadyUpToDate       = errors.New("already up to date")
)

const DefaultTableName = "migrations"

type Migrate struct {
	Log       *logger.Logger
	driver    database.Driver
	tableName string
	dir       string
}

// Migrations slice.
type Migrations []*Migration

func NewMigrator(dsn string, tableName string, dir string) (*Migrate, error) {
	if tableName == "" {
		tableName = DefaultTableName
	}

	// Get driver.
	driver, err := database.Open(dsn, tableName)
	if err != nil {
		return nil, fmt.Errorf("can't get driver: %w", err)
	}

	migrate := &Migrate{
		driver:    driver,
		tableName: tableName,
		dir:       dir,
	}

	// Create table if it does not exist.
	err = migrate.prepareDatabase()

	if err != nil {
		return nil, fmt.Errorf("can't initialize table: %w", err)
	}

	return migrate, nil
}

func (m *Migrate) Up() error {
	if err := m.lock(); err != nil {
		return err
	}

	migrations, err := m.migrationsForRun(true, 0)
	if err != nil {
		return m.unlock(err)
	}

	for _, migration := range migrations {
		if err := m.driver.Run(strings.NewReader(migration.UpSQL)); err != nil {
			return m.unlock(fmt.Errorf("can't execute migration with version %d: %w", migration.Version, err))
		}

		// Set version if success.
		m.setVersion(migration.Version)
		m.printLog(fmt.Sprintf("Migration %d successfully applied!", migration.Version))
	}

	return m.unlock(nil)
}

func (m *Migrate) Down() error {
	if err := m.lock(); err != nil {
		return err
	}

	migrations, err := m.migrationsForRun(false, 1)
	if err != nil {
		return m.unlock(err)
	}

	for _, migration := range migrations {
		if err := m.driver.Run(strings.NewReader(migration.DownSQL)); err != nil {
			return m.unlock(fmt.Errorf("can't rollback migration with version %d: %w", migration.Version, err))
		}

		// Delete version if success.
		m.deleteVersion(migration.Version)
		m.printLog(fmt.Sprintf("Migration %d successfully rollback!", migration.Version))
	}

	return m.unlock(nil)
}

func (m *Migrate) Redo() error {
	if err := m.lock(); err != nil {
		return err
	}

	currentMigration, err := m.currentMigration()
	if errors.Is(err, ErrNoCurrentVersion) {
		m.printLog(err.Error())
		return m.unlock(nil)
	}

	// Rollback it first.
	if err := m.driver.Run(strings.NewReader(currentMigration.DownSQL)); err != nil {
		return m.unlock(fmt.Errorf("can't rollback migration with version %d: %w", currentMigration.Version, err))
	}
	m.deleteVersion(currentMigration.Version)
	m.printLog(fmt.Sprintf("Migration %d successfully rollback!", currentMigration.Version))

	// ...and then run to up
	if err := m.driver.Run(strings.NewReader(currentMigration.UpSQL)); err != nil {
		return m.unlock(fmt.Errorf("can't execute migration with version %d: %w", currentMigration.Version, err))
	}
	m.setVersion(currentMigration.Version)
	m.printLog(fmt.Sprintf("Migration %d successfully applied!", currentMigration.Version))

	return m.unlock(nil)
}

func (m *Migrate) DbVersion() (int64, error) {
	if err := m.lock(); err != nil {
		return -1, err
	}
	defer m.unlock(nil)

	currentMigration, err := m.currentMigration()
	if err != nil {
		return -1, ErrNoCurrentVersion
	}

	m.printLog(fmt.Sprintf("Current migration version: %d", currentMigration.Version))

	return currentMigration.Version, nil
}

// Close migrator API.
// Just close DB connection in our case.
func (m *Migrate) Close() error {
	return m.driver.Close()
}

func (m *Migrate) Status() (Migrations, error) {
	migrations := make([]*Migration, 0)

	if err := m.lock(); err != nil {
		return migrations, err
	}

	list, err := m.driver.List()
	if err != nil {
		return migrations, m.unlock(fmt.Errorf("can't get full list of applied migraions: %w", err))
	}

	// Get available migrations.
	availableMigrations, err := m.findAvailableMigrations()
	if err != nil {
		return migrations, m.unlock(err)
	}

	// Mapping.
	for _, appliedMigration := range list {
		migration, err := m.getMigrationByVersion(availableMigrations, appliedMigration.Version)
		if err == nil {
			// Add applied_at time.
			migration.AppliedAt = appliedMigration.AppliedAt
			migrations = append(migrations, migration)
		}
	}

	m.unlock(nil)

	return migrations, nil
}

// Prepare migrations slice for next Run
// up -- direction
// limit -- how many migrations should be executed (0 -- without limit).
func (m *Migrate) migrationsForRun(up bool, limit int) (Migrations, error) {
	// Get available migrations.
	availableMigrations, err := m.findAvailableMigrations()
	if err != nil {
		return make(Migrations, 0), err
	}

	// No available migrations, so skip all the next.
	if len(availableMigrations) == 0 {
		return make(Migrations, 0), ErrNoAvailableMigrations
	}

	// Get list of applied migrations.
	listAppliedMigrations, err := m.list()
	if err != nil {
		return make(Migrations, 0), err
	}

	var appliedVersions []int64
	for _, ap := range listAppliedMigrations {
		appliedVersions = append(appliedVersions, ap.Version)
	}

	// If we go down and don't have any applied migrations - do nothing.
	if len(appliedVersions) == 0 && !up {
		return make(Migrations, 0), ErrAlreadyUpToDate
	}

	// If we go up and don't have any applied migrations - run all then.
	if len(appliedVersions) == 0 && up {
		return availableMigrations, nil
	}

	var migrationsForRun Migrations

	// Calc the difference between them.
	if !up {
		// Sort desc.
		sort.Slice(availableMigrations, func(i, j int) bool {
			return availableMigrations[i].Version > availableMigrations[j].Version
		})

		// Filter them.
		for _, migration := range availableMigrations {
			if migration.Version <= appliedVersions[len(appliedVersions)-1] {
				migrationsForRun = append(migrationsForRun, migration)
			}
		}
	} else {
		// Filter them.
		for _, migration := range availableMigrations {
			// If it already applied - skip.
			if slices.Contains(appliedVersions, migration.Version) {
				continue
			}

			if migration.Version > appliedVersions[0] {
				migrationsForRun = append(migrationsForRun, migration)
			}
		}
	}

	if len(migrationsForRun) == 0 {
		return make(Migrations, 0), ErrAlreadyUpToDate
	}

	// Slice target slice.
	if limit > 0 {
		migrationsForRun = migrationsForRun[0:limit]
	}

	return migrationsForRun, nil
}

func (m *Migrate) currentMigration() (*Migration, error) {
	// Get available migrations.
	availableMigrations, err := m.findAvailableMigrations()
	if err != nil {
		return nil, err
	}

	// Get current migration version from DB.
	currentVersion, err := m.current()
	if err != nil {
		return nil, err
	}

	// Get available migration by version.
	migration, err := m.getMigrationByVersion(availableMigrations, currentVersion)
	if err != nil {
		return nil, ErrNoCurrentVersion
	}

	return migration, nil
}

func (m *Migrate) getMigrationByVersion(migrations Migrations, version int64) (*Migration, error) {
	for _, migrate := range migrations {
		if migrate.Version == version {
			return migrate, nil
		}
	}

	return nil, fmt.Errorf("no migration find by version %d", version)
}

func (m *Migrate) findAvailableMigrations() (Migrations, error) {
	migrations := make([]*Migration, 0)

	file, err := http.Dir(m.dir).Open(".")
	if err != nil {
		return nil, err
	}

	files, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}

	for _, info := range files {
		name := info.Name()
		if strings.HasSuffix(name, ".sql") {
			migration, err := m.parseSQLMigration(info)
			if err != nil {
				return nil, err
			}

			migrations = append(migrations, migration)
		}
	}

	// Insure then they are sorted by version correctly.
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// Parse SQL migration file.
func (m *Migrate) parseSQLMigration(info fs.FileInfo) (*Migration, error) {

	file, err := http.Dir(m.dir).Open(path.Join("./", info.Name()))
	if err != nil {
		return nil, fmt.Errorf("error while opening %s: %w", info.Name(), err)
	}
	defer func() { _ = file.Close() }()

	if err != nil {
		return nil, fmt.Errorf("error while opening %s: %w", info.Name(), err)
	}

	version := m.getVersionFromFileName(info.Name())

	migration := &Migration{
		Version: version,
		Source:  info.Name(),
	}

	parsed, err := parser.ParseMigration(file)
	if err != nil {
		return nil, fmt.Errorf("error while parsing file %s: %w", info.Name(), err)
	}

	// Set statements.
	migration.UpSQL = parsed.UpStatements
	migration.DownSQL = parsed.DownStatements

	return migration, nil
}

func (m *Migrate) getVersionFromFileName(filename string) int64 {
	version := strings.Split(filename, "_")[0]
	i, _ := strconv.ParseInt(version, 10, 64)

	return i
}

// Create migrations table if it doesn't exist.
func (m *Migrate) prepareDatabase() error {
	return m.driver.PrepareTable()
}

func (m *Migrate) setVersion(version int64) error {
	err := m.driver.SetVersion(version)
	if err != nil {
		return fmt.Errorf("can't set new migraion version: %w", err)
	}

	return nil
}

func (m *Migrate) deleteVersion(version int64) error {
	err := m.driver.DeleteVersion(version)
	if err != nil {
		return fmt.Errorf("can't delete migraion version: %w", err)
	}

	return nil
}

func (m *Migrate) list() ([]*database.ListInfo, error) {
	list, err := m.driver.List()
	if err != nil {
		return []*database.ListInfo{}, fmt.Errorf("can't get list of applied migraions: %w", err)
	}

	return list, nil
}

// Get current migration version from DB driver.
func (m *Migrate) current() (int64, error) {
	curVersion, err := m.driver.Version()
	if err != nil {
		return -1, fmt.Errorf("can't get current migration: %w", err)
	}

	return curVersion, nil
}

// Lock the driver.
func (m *Migrate) lock() error {
	return m.driver.Lock()
}

// Release lock and return err if exists.
func (m *Migrate) unlock(prevError error) error {
	if err := m.driver.Unlock(); err != nil {
		finalError := fmt.Errorf("can't unlock from database driver: %w", err)
		if prevError != nil {
			finalError = fmt.Errorf("%w. Additional err: %w", finalError, prevError)
		}

		return finalError
	}

	return prevError
}

func (m *Migrate) printLog(msg string) {
	if m.Log != nil {
		m.Log.Info(msg)
	}
}
