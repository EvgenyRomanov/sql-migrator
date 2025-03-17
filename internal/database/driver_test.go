package database

import (
	"io"
	"testing"
)

type testDriver struct {
	url       string
	tableName string
}

func (t *testDriver) Open(url string, tableName string) (Driver, error) {
	return &testDriver{
		url:       url,
		tableName: tableName,
	}, nil
}

func (t *testDriver) Close() error {
	return nil
}

func (t *testDriver) Lock() error {
	return nil
}

func (t *testDriver) Unlock() error {
	return nil
}

func (t *testDriver) Run(_ io.Reader) error {
	return nil
}

func (t *testDriver) SetVersion(_ int64) error {
	return nil
}

func (t *testDriver) DeleteVersion(_ int64) error {
	return nil
}

func (t *testDriver) Version() (_ int64, err error) {
	return 0, nil
}

func (t *testDriver) List() (_ []*ListInfo, err error) {
	return make([]*ListInfo, 0), nil
}

func (t *testDriver) PrepareTable() error {
	return nil
}

func TestOpen(t *testing.T) {
	Register("test", &testDriver{})

	useCases := []struct {
		url string
		err bool
	}{
		{
			"test://app:!ChangeMe!@pgsql:5432/app?serverVersion=14&charset=utf8",
			false,
		},
		{
			"postgresql://app:!ChangeMe!@pgsql:5432/app?serverVersion=14&charset=utf8",
			true,
		},
	}

	for _, useCase := range useCases {
		t.Run(useCase.url, func(t *testing.T) {
			driver, err := Open(useCase.url, "migrations")

			if err == nil && useCase.err {
				t.Fatal("should be error for wrong driver")
			}

			if err == nil && !useCase.err {
				if md, ok := driver.(*testDriver); !ok {
					t.Fatalf("expected *testDriver got %T", driver)
				} else if md.url != useCase.url {
					t.Fatalf("expected %q got %q", useCase.url, md.url)
				}
			}

			if !useCase.err && err != nil {
				t.Fatalf("did not expect %q", err)
			}
		})
	}
}
