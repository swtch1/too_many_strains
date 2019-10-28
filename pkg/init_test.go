package tms

import (
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
	cleanupDB()
	os.Exit(code)
}

// setupDB just creates a test database we can integrate with.
func setupDB() {
	if Integration {
		dbSrv := NewDBServer(TestDatabaseName, "root", "password")
		if err := dbSrv.Migrate(); err != nil {
			panic(errors.Wrap(err, "unable to migrate test db"))
		}
		if err := dbSrv.Open(); err != nil {
			panic(errors.Wrapf(err, "unable to open test db"))
		}
		TestDB = dbSrv.DB
	}
}

func cleanupDB() {
	// TODO: dropping the database would be better here but initial attempts failed.  this
	// TODO: method is more brittle, but will do for now.
	if Integration {
		TestDB.DropTable("database_ver")
		TestDB.DropTable("strain")
		TestDB.DropTable("effect")
		TestDB.DropTable("flavor")
		TestDB.DropTable("strain_effects")
		TestDB.DropTable("strain_flavors")
	}
}

type nopWriter struct{}

func (w nopWriter) Write(p []byte) (int, error) {
	return 0, nil
}
