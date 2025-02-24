package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

type ParsedMigration struct {
	UpStatements   string
	DownStatements string
}

var prefix = "-- +gomigrator"

var ErrIncorrectTemplate = errors.New("incorrect sql-migration template")

func ParseMigration(r io.ReadSeeker) (*ParsedMigration, error) {
	p := &ParsedMigration{}

	_, err := r.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var direction string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, prefix+" Up") {
			direction = "up"
		}

		if strings.HasPrefix(line, prefix+" Down") {
			if direction != "" {
				buf.Reset()
			}

			direction = "down"
		}

		// If no direction found, terminate.
		if direction == "" {
			return nil, ErrIncorrectTemplate
		}

		if !strings.HasPrefix(line, "-- +") {
			if _, err := buf.WriteString(line + "\n"); err != nil {
				return nil, err
			}
		}

		if direction == "up" {
			p.UpStatements += buf.String()
		} else {
			p.DownStatements += buf.String()
		}
		buf.Reset()
	}

	return p, nil
}
