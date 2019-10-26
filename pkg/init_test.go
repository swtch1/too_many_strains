package tms

import (
	"flag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var (
	// Integration signals that integration test should be run and can be
	// configured with `go test ./... -args -Integration=true`
	Integration bool
)

func init() {
	flag.BoolVar(&Integration, "Integration", false, "set true to enable Integration testing")
}

// do work to initialize the tests
func TestMain(m *testing.M) {
	// don't produce normal logs during tests
	log.SetOutput(nopWriter{})

	setupDB()
	code := m.Run()
	cleanupDB()
	os.Exit(code)
}

// setupDB just creates a test database we can integrate with.
func setupDB() {
	if Integration {
		dbSrv := NewDBServer(TestDatabaseName, "root", "password")
		if err := dbSrv.Migrate(); err != nil {
			panic(errors.Wrap(err, ""))
		}
	}
}

func cleanupDB() {
	if Integration {

	}
}

type nopWriter struct{}

func (w nopWriter) Write(p []byte) (int, error) {
	return 0, nil
}
