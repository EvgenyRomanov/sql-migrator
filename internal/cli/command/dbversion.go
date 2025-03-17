package command

import (
	"errors"

	"github.com/EvgenyRomanov/sql-migrator/internal/logger"
	"github.com/EvgenyRomanov/sql-migrator/pkg/core"
)

type DBVersion struct {
	Migrator *core.Migrate
	Logger   *logger.Logger
}

func (c *DBVersion) Run(_ []string) error {
	_, err := c.Migrator.DBVersion()

	if errors.Is(err, core.ErrNoCurrentVersion) {
		c.Logger.Info("%s", err.Error())
		return nil
	}

	return err
}
