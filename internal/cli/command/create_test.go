package command

import (
	"fmt"
	"github.com/EvgenyRomanov/sql-migrator/internal/cli/config"
	"github.com/EvgenyRomanov/sql-migrator/internal/logger"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	// Disable logger for testing.
	logger := logger.New("DEBUG", io.Discard)

	tests := []struct {
		cfg              *config.MigratorConf
		filenames        []string
		expectedFileName string
		expectedErr      error
	}{
		{
			&config.MigratorConf{
				DSN: "",
				Dir: t.TempDir(),
			},
			[]string{
				"TestMigrationSQL",
				"Test.migration.SQL",
			},
			"test_migration_sql.sql",
			nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			cmd := &Create{
				Cfg:    tt.cfg,
				Logger: logger,
			}

			for _, f := range tt.filenames {
				time.Sleep(1 * time.Millisecond)
				err := cmd.create(f)
				if tt.expectedErr == nil && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				files, err := os.ReadDir(tt.cfg.Dir)
				if err != nil {
					t.Fatal(err)
				}

				// Check created files.
				for _, f := range files {
					if !strings.Contains(f.Name(), tt.expectedFileName) {
						t.Errorf("Error: Expected contains: %v, but received: %v", tt.expectedFileName, f.Name())
					}
				}
			}
		})
	}
}
