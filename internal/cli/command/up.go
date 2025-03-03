package command

import (
	"github.com/EvgenyRomanov/sql-migrator/internal/logger"
	"github.com/EvgenyRomanov/sql-migrator/pkg/core"
)

type Up struct {
	Migrator *core.Migrate
	Logger   *logger.Logger
}

func (c *Up) Run(_ []string) error {
	return c.Migrator.Up()
}
