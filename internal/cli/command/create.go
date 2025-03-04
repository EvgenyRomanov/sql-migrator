package command

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EvgenyRomanov/sql-migrator/internal/cli/config"
	"github.com/EvgenyRomanov/sql-migrator/internal/logger"
)

var ErrMissingName = errors.New("no migration name was set")

type Create struct {
	Cfg    *config.MigratorConf
	Logger *logger.Logger
}

func (c *Create) Run(args []string) error {
	if len(args) == 0 {
		return ErrMissingName
	}

	return c.create(args[0])
}

func (c *Create) create(name string) error {
	// Define new version of migration file.
	version := time.Now().UTC().UnixMilli()

	// Define full filename.
	fullName := fmt.Sprintf("%v_%v", version, c.snakeCase(name))
	filename := fmt.Sprintf("%v.sql", fullName)

	// Define template.
	tmpl := sqlMigrationTemplate

	// Try to create path.
	err := os.MkdirAll(c.Cfg.Dir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create migration folder: %w", err)
	}

	// Target path.
	path := filepath.Join(c.Cfg.Dir, filename)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	// Try to create file.
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create migration file2: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, struct{}{}); err != nil {
		return fmt.Errorf("failed to execute tmpl: %w", err)
	}

	c.Logger.Info("Success create new migration %s", filename)
	return nil
}

func (c *Create) snakeCase(s string) string {
	var b strings.Builder

	diff := 'a' - 'A'
	l := len(s)

	for i, v := range s {
		// Replace all dots and other "danger" symbols.
		ss := string(v)

		if ss == "+" || ss == "-" || ss == "—" || ss == "." || ss == "/" || ss == "," || ss == "_" {
			b.WriteRune('_')
			continue
		}

		if v >= 'a' {
			b.WriteRune(v)
			continue
		}

		if (i != 0 || i == l-1) && ((i > 0 && rune(s[i-1]) >= 'a') ||
			(i < l-1 && rune(s[i+1]) >= 'a')) {
			b.WriteRune('_')
		}

		b.WriteRune(v + diff)
	}

	return b.String()
}

var sqlMigrationTemplate = template.Must(template.New("gomigrator.sql-migration").Parse(`-- +gomigrator Up
SELECT 'up SQL query';

-- +gomigrator Down
SELECT 'down SQL query';
`))
