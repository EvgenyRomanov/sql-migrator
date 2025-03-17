package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/EvgenyRomanov/sql-migrator/internal/cli/command"
	"github.com/EvgenyRomanov/sql-migrator/internal/cli/config"
	"github.com/EvgenyRomanov/sql-migrator/internal/logger"
	"github.com/EvgenyRomanov/sql-migrator/pkg/core"
)

func initFlags() {
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			`Usage: gomigrator [OPTIONS] COMMAND [arg...]
  
  You can override varuables from config file by ENV, just use something like "${DB_DSN}"

  OPTIONS:
    -config         Path to configuration file (no default value)
    -dsn            DSN string to database
    -dir            Folder for migrations files ("./migrations" by default)
    -tableName      Name of migrations table ("migrations" by default)	
		
  COMMAND:
    create [name]   Create migration with 'name'
    up              Migrate the DB to the most recent version available
    down            Roll back the version by 1
    redo            Re-run the latest migration
    status          Print all migrations status
    dbversion       Print migrations status (last applied migration)
    help            Print usage
    version         Application version

  Examples:
    gomigrator -config="../configs/config-test.yml" create "create_user_table"
    DB_DSN="postgresql://app:test@pgsql:5432/app" gomigrator up
`)
	}
}

func printUsage() {
	flag.Usage()
	os.Exit(1)
}

func Main() {
	initFlags()

	// Get args.
	args := os.Args[1:]

	// Do not init anything if no arguments or just help.
	if len(args) == 0 || args[0] == "help" {
		printUsage()
	}

	// Init config.
	cfg := config.NewConfig()

	// Init logger.
	logger := logger.New(cfg.Logger.Level, os.Stdout)

	var cmd command.Command

	// Init migrate api
	migrator, err := core.NewMigrator(cfg.Migrator.DSN, cfg.Migrator.TableName, cfg.Migrator.Dir)
	if err != nil {
		logger.Error("[ERROR] Can't initialize migrator api! %s", err)
		return
	}

	// Add logger.
	migrator.Log = logger

	// Close migrator.
	defer migrator.Close()

	switch flag.Arg(0) {
	case "create":
		cmd = &command.Create{
			Cfg:    &cfg.Migrator,
			Logger: logger,
		}
	case "up":
		cmd = &command.Up{
			Migrator: migrator,
			Logger:   logger,
		}
	case "down":
		cmd = &command.Down{
			Migrator: migrator,
			Logger:   logger,
		}
	case "redo":
		cmd = &command.Redo{
			Migrator: migrator,
			Logger:   logger,
		}
	case "dbversion":
		cmd = &command.DBVersion{
			Migrator: migrator,
			Logger:   logger,
		}
	case "status":
		cmd = &command.Status{
			Migrator: migrator,
		}
	default:
		printUsage()
	}

	err = cmd.Run(args[2:])

	if errors.Is(err, core.ErrAlreadyUpToDate) || errors.Is(err, core.ErrNoAvailableMigrations) {
		logger.Info("%s", err.Error())
	} else if err != nil {
		logger.Error("Error executing CLI: %s\n", err.Error())
		logger.Info("Try 'gomigrator help' for more information.")
	}
}
