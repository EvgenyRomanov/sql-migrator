package core

import "time"

type Migration struct {
	// Migration version.
	Version int64

	// Path to file.
	Source string

	// The time of migration application.
	AppliedAt time.Time

	// Statements to run up (used by SQL-migrations).
	UpSQL string

	// Statements to run down (used by SQL-migrations).
	DownSQL string
}

func New() *Migration {
	return &Migration{}
}
