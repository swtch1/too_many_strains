package tms

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var (
	// Integration signals that integration test should be run and can be
	// configured with `go test ./... -args -Integration=true`
	Integration      bool
	TestDatabaseName = "test_db"
	// TestDB which can be used by all tests concurrently.  The
	TestDB *gorm.DB
)

// do work to initialize the tests
func TestMain(m *testing.M) {
	// don't produce normal logs during tests
	log.SetOutput(nopWriter{})

	setupDB()
	code := m.Run()
	cleanupTables()
	os.Exit(code)
}

// setupDB just creates a test database we can integrate with.
func setupDB() {
	setup := func() {
		dbSrv := NewDBServer(TestDatabaseName, "root", "password")
		if err := dbSrv.Migrate(); err != nil {
			panic(errors.Wrap(err, "unable to migrate test db"))
		}
		if err := dbSrv.Open(); err != nil {
			panic(errors.Wrapf(err, "unable to open test db"))
		}
		TestDB = dbSrv.DB
	}

	defer func() {
		if r := recover(); r != nil {
			// we may panic on cleanupTables if the DB doesn't exist, but it's helpful
			// to run it pre-tests anyway just to ensure consistency.  just recover and
			// keep going.
			setup()
		}
	}()
	if Integration {
		cleanupTables()
		setup()
	}
}

func cleanupTables() {
	// TODO: dropping the database would be better here but initial attempts failed.  this
	// TODO: method is more brittle, but will do for now.
	dbSrv := NewDBServer(TestDatabaseName, "root", "password")
	if err := dbSrv.Open(); err != nil {
		fmt.Printf("%s: unable to open test db", err)
	}
	if Integration {
		tables := []string{
			"database_ver",
			"strain",
			"effect",
			"flavor",
			"strain_effects",
			"strain_flavors",
		}
		for _, tbl := range tables {
			if dbSrv.DB.HasTable(tbl) {
				dbSrv.DB.DropTable(tbl)
			}
		}
	}
}

type nopWriter struct{}

func (w nopWriter) Write(p []byte) (int, error) {
	return 0, nil
}
