// Create or migrate the strains database.
package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/swtch1/too_many_strains/cmd/database-migration/cli"
	tms "github.com/swtch1/too_many_strains/pkg"
	"os"
)

var (
	// dbVersion is the desired version of the database
	dbVersion uint = 1
)

func main() {
	cli.Init("migrate", "v1")
	tms.InitLogger(os.Stderr, cli.LogLevel, "text", false)

	dbSrv := tms.NewDBServer(cli.DatabaseName, cli.DatabaseUsername, cli.DatabasePassword)
	dbSrv.DBIteration = dbVersion
	if err := dbSrv.Migrate(); err != nil {
		log.Fatal(err)
	}

	if err := dbSrv.Open(); err != nil {
		log.Fatal(err)
	}
	defer dbSrv.Close()

	log.Tracef("reading seed file %s", cli.SeedFile)
	seedFile, err := os.Open(cli.SeedFile)
	if err != nil {
		log.WithError(err).Fatalf("unable to read seed file %s", cli.SeedFile)
	}
	defer seedFile.Close()
	strainReprs, err := tms.ParseStrains(seedFile)
	if err != nil {
		log.WithError(err).Fatalf("unable to parse seed file %s", cli.SeedFile)
	}

	log.Infof("populating database with strains from seed file %s", cli.SeedFile)
	for _, repr := range strainReprs {
		repr.DB = dbSrv.DB
		if err := repr.ReplaceInDB(); err != nil {
			log.WithError(err).Errorf("population failed for strain ID %d", repr.ID)
		}
	}
}
