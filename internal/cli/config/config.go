package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Migrator MigratorConf `mapstructure:"migrator"`
	Logger   LoggerConf   `mapstructure:"logger"`
}

type MigratorConf struct {
	DSN       string `mapstructure:"dsn"`
	Dir       string `mapstructure:"dir"`
	TableName string `mapstructure:"table_name"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
}

var (
	configFile string
	dsn        string
	dir        string
	tableName  string
)

func initFlag() {
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.StringVar(&dsn, "dsn", "", "Database string connection")
	flag.StringVar(&dir, "dir", "./migrations", "Path to migration folder")
	flag.StringVar(&tableName, "tableName", "migrations", "Name of migrations table")

	flag.Parse()
}

func NewConfig() *Config {
	initFlag()

	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configFile)

	var config Config

	// Read from file.
	if configFile != "" {
		if err := v.ReadInConfig(); err != nil {
			fmt.Printf("[ERROR] Couldn't load config: %s\n", err)
			os.Exit(1)
		}

		// Support to set ENV variables in file.
		for _, k := range v.AllKeys() {
			val := v.GetString(k)
			v.Set(k, os.ExpandEnv(val))
		}

		if err := v.Unmarshal(&config); err != nil {
			fmt.Printf("[ERROR] Couldn't read config: %s\n", err)
		}
	} else {
		// Use values from flags.
		config.Migrator = MigratorConf{
			DSN:       dsn,
			Dir:       dir,
			TableName: tableName,
		}

		// Default config for logger.
		config.Logger = LoggerConf{
			Level: "INFO",
		}
	}

	if config.Migrator.DSN == "" {
		fmt.Printf("[ERROR] Wrong configuration of app. Cannot get a DSN setting!\n")
		os.Exit(1)
	}

	if config.Migrator.Dir == "" {
		fmt.Printf("[ERROR] Wrong configuration of app. Cannot get a Dir setting!\n")
		os.Exit(1)
	}

	if config.Migrator.TableName == "" {
		fmt.Printf("[ERROR] Wrong configuration of app. Cannot get a TableName setting!\n")
		os.Exit(1)
	}

	return &config
}
