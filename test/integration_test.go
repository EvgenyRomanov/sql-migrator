//go:build integration

package test

import (
	"fmt"
	"github.com/EvgenyRomanov/sql-migrator/internal/database"
	"github.com/EvgenyRomanov/sql-migrator/pkg/core"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
	"time"
)

type MigratorSuite struct {
	suite.Suite
	migrator *core.Migrate

	dsn    string
	driver database.Driver
}

const DefaultTableName = "migrations"

func (s *MigratorSuite) SetupSuite() {
	dsn := os.Getenv("DSN")
	dir := os.Getenv("DIR")

	s.dsn = dsn

	// Init migrator by data from env.
	migrator, err := core.NewMigrator(dsn, DefaultTableName, dir)
	s.Require().NoError(err)
	s.Require().NotNil(migrator)
	s.migrator = migrator

	// Init additional driver connection for checking.
	driver, err := database.Open(s.dsn, DefaultTableName)
	s.Require().NoError(err)
	s.Require().NotNil(driver)
	s.driver = driver
}

// close connection after finishing suite.
func (s *MigratorSuite) TearDownSuite() {
	defer s.migrator.Close()
}

// clear everything after each test.
func (s *MigratorSuite) TearDownTest() {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS test;TRUNCATE %s`, DefaultTableName)
	err := s.driver.Run(strings.NewReader(query))
	s.Require().NoError(err)
}

func (s *MigratorSuite) TestMigratorUp() {
	// Ensure that migrator exists.
	s.NotNil(s.T(), s.migrator)

	// Run all up.
	err := s.migrator.Up()
	s.NoError(err)
	s.checkAppliedListCount(3)

	// Run all up again.
	err = s.migrator.Up()
	s.ErrorIs(err, core.ErrAlreadyUpToDate)
	s.checkAppliedListCount(3)
}

func (s *MigratorSuite) TestMigratorDown() {
	// Ensure that migrator exists.
	s.NotNil(s.T(), s.migrator)

	// Run all up.
	err := s.migrator.Up()
	s.NoError(err)
	s.checkAppliedListCount(3)

	// Run one Down.
	err = s.migrator.Down()
	s.NoError(err)
	s.checkAppliedListCount(2)

	// One more Down.
	err = s.migrator.Down()
	s.NoError(err)
	s.checkAppliedListCount(1)

	// One more Down.
	err = s.migrator.Down()
	s.NoError(err)
	s.checkAppliedListCount(0)

	// Final Down.
	err = s.migrator.Down()
	s.ErrorIs(err, core.ErrAlreadyUpToDate)
	s.checkAppliedListCount(0)
}

func (s *MigratorSuite) TestMigratorRedo() {
	// Ensure that migrator exists.
	s.NotNil(s.T(), s.migrator)

	// Run all up.
	err := s.migrator.Up()
	s.NoError(err)
	s.checkAppliedListCount(3)

	// Get last migration.
	list, err := s.migrator.Status()
	s.NoError(err)
	lastMigration := list[len(list)-1]

	time.Sleep(1 * time.Second)

	// Run Redo.
	err = s.migrator.Redo()
	s.NoError(err)
	s.checkAppliedListCount(3)

	// Get last migration after Redo.
	list, err = s.migrator.Status()
	s.NoError(err)
	lastMigrationAfterRedo := list[len(list)-1]

	s.NotEqual(lastMigration.AppliedAt.Unix(), lastMigrationAfterRedo.AppliedAt.Unix())
	s.Greater(lastMigrationAfterRedo.AppliedAt.Unix(), lastMigration.AppliedAt.Unix())
}

func (s *MigratorSuite) TestMigratorDBVersion() {
	// Ensure that migrator exists.
	s.NotNil(s.T(), s.migrator)

	// Check empty version.
	version, err := s.migrator.DBVersion()
	s.ErrorIs(err, core.ErrNoCurrentVersion)
	s.Equal(version, int64(-1))

	// Run all up.
	err = s.migrator.Up()
	s.NoError(err)
	s.checkAppliedListCount(3)

	// Get last migration.
	list, err := s.migrator.Status()
	s.NoError(err)
	lastMigration := list[len(list)-1]

	// Check that are equals.
	version, err = s.migrator.DBVersion()
	s.NoError(err)
	s.Equal(lastMigration.Version, version)
}

// Check applied list.
func (s *MigratorSuite) checkAppliedListCount(expectedCount int) {
	list, err := s.migrator.Status()
	s.NoError(err)
	s.Equal(len(list), expectedCount)
}

func TestMigrator(t *testing.T) {
	suite.Run(t, new(MigratorSuite))
}
