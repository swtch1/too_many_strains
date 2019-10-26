// Create or migrate the strains database.
package main

import (
	"github.com/swtch1/too_many_strains/cmd/database-migration/cli"
	tms "github.com/swtch1/too_many_strains/pkg"
	"log"
	"os"
)

var (
	// dbVersion is the desired version of the database
	dbVersion = 1
	version   = "v1"
)

func main() {
	cli.Init("migrate", version)
	tms.InitLogger(os.Stderr, cli.LogLevel, "text", false)

	dbSrv := tms.NewDBServer(tms.DefaultDatabaseName, cli.DatabaseUsername, cli.DatabasePassword)
	if err := dbSrv.Migrate(); err != nil {
		log.Fatal(err)
	}
	if err := dbSrv.Open(); err != nil {
		log.Fatal(err)
	}
	defer dbSrv.Close()
}

// New does lots of cool stuff.
func New() string {
	return ""
}
