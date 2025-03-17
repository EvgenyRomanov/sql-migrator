package command

import (
	"github.com/EvgenyRomanov/sql-migrator/internal/logger"
	"github.com/EvgenyRomanov/sql-migrator/pkg/core"
)

type Down struct {
	Migrator *core.Migrate
	Logger   *logger.Logger
}

func (c *Down) Run(_ []string) error {
	return c.Migrator.Down()
}
