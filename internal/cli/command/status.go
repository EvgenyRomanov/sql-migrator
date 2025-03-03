package command

import (
	"errors"
	"github.com/EvgenyRomanov/sql-migrator/pkg/core"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

var ErrGeneralError = errors.New("unable to show status table")

type Status struct {
	Migrator *core.Migrate
}

func (c *Status) Run(_ []string) error {
	migrations, err := c.Migrator.Status()
	if err != nil {
		return ErrGeneralError
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Version", "Name", "Applied At"})

	for i, migration := range migrations {
		t.AppendRows([]table.Row{
			{i + 1, migration.Version, migration.Source, migration.AppliedAt.Format("2006-01-02 15:04:05")},
		})
	}

	t.AppendSeparator()
	t.AppendFooter(table.Row{"", "Total", len(migrations)})
	t.Render()

	return nil
}
