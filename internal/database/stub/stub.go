package stub

import (
	"github.com/EvgenyRomanov/sql-migrator/internal/database"
	"io"
)

type Stub struct {
	url       string
	tableName string
	isLocked  bool
	version   int64
	list      []*database.ListInfo
}

func init() {
	s := Stub{}
	database.Register("stub", &s)
}

func (p *Stub) Open(url string, tableName string) (database.Driver, error) {
	instance := &Stub{
		url:       url,
		tableName: tableName,
	}

	return instance, nil
}

func (p *Stub) Close() error {
	return nil
}

func (p *Stub) Lock() error {
	if p.isLocked {
		return database.ErrLocked
	}
	return nil
}

func (p *Stub) Unlock() error {
	return nil
}

func (p *Stub) Run(_ io.Reader) error {
	return nil
}

func (p *Stub) SetVersion(version int64) error {
	p.version = version

	return nil
}

func (p *Stub) DeleteVersion(_ int64) error {
	return nil
}

func (p *Stub) Version() (int64, error) {
	return p.version, nil
}

func (p *Stub) List() ([]*database.ListInfo, error) {
	return p.list, nil
}

func (p *Stub) PrepareTable() error {
	return nil
}
