package core

import (
	"testing"

	_ "github.com/EvgenyRomanov/sql-migrator/internal/database/stub"
	"github.com/stretchr/testify/assert"
)

var testMigrator *Migrate

func init() {
	testMigrator, _ = NewMigrator(
		"stub://stub:stub@localhost:3254/gomigrator",
		"migrations",
		"../../test/migrations")
}

func TestFindAvailableMigrations(t *testing.T) {
	migrations, err := testMigrator.findAvailableMigrations()

	assert.NoError(t, err)

	assert.Len(t, migrations, 3)

	m := migrations[0]
	assert.Equal(t, int64(20250302201917), m.Version)

	m = migrations[1]
	assert.Equal(t, int64(20250302211917), m.Version)
}

func TestGetVersionFromFileName(t *testing.T) {
	version := testMigrator.getVersionFromFileName("1234567_qwerty_test_migration.sql")
	assert.NotEmpty(t, version)
	assert.Equal(t, version, int64(1234567))
}
