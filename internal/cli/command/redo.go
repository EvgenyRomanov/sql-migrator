package command

import (
	"github.com/EvgenyRomanov/sql-migrator/internal/logger"
	"github.com/EvgenyRomanov/sql-migrator/pkg/core"
)

type Redo struct {
	Migrator *core.Migrate
	Logger   *logger.Logger
}

func (c *Redo) Run(_ []string) error {
	return c.Migrator.Redo()
}
