// Create or migrate the strains database.
package main

import (
	"github.com/jinzhu/gorm"
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
		log.WithError(err).Fatalf("unable to read seed file '%s'", cli.SeedFile)
	}
	defer seedFile.Close()
	strains, err := tms.ParseStrains(seedFile)
	if err != nil {
		log.WithError(err).Fatalf("unable to parse seed file '%s'", cli.SeedFile)
	}

	log.Infof("populating database with strains from seed file %s", cli.SeedFile)
	for _, strain := range strains {
		if err := writeStrain(strain, dbSrv.DB); err != nil {
			log.WithError(err).Errorf("population failed for strain ID %d", strain.ID)
		}
	}
}

func writeStrain(strain tms.StrainRepr, db *gorm.DB) error {
	tx := db.Begin()
	var flavors []tms.Flavor
	for _, flavor := range strain.Flavors {
		f := tms.Flavor{Name: flavor}
		tx.Model(&f).Where(&f).FirstOrCreate(&f)
		flavors = append(flavors, f)
	}

	var effects []tms.Effect
	for _, effect := range strain.Effects.Positive {
		e := tms.Effect{Name: effect, Category: "positive"}
		tx.Model(&e).Where(&e).FirstOrCreate(&e)
		effects = append(effects, e)
	}
	for _, effect := range strain.Effects.Negative {
		e := tms.Effect{Name: effect, Category: "negative"}
		tx.Model(&e).Where(&e).FirstOrCreate(&e)
		effects = append(effects, e)
	}
	for _, effect := range strain.Effects.Medical {
		e := tms.Effect{Name: effect, Category: "medical"}
		tx.Model(&e).Where(&e).FirstOrCreate(&e)
		effects = append(effects, e)
	}

	var s tms.Strain
	s.ReferenceID = strain.ID
	tx.Model(&s).Where(&s).FirstOrCreate(&s)

	s.Name = strain.Name
	s.Race = strain.Race
	s.Flavors = flavors
	s.Effects = effects
	// FIXME: add s.SaveInDB

	if err := tx.Model(&s).Save(&s).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Errorf("unable to save record for strain with ID %d", strain.ID)
	}

	log.Debugf("updating record for strain %s with ID %d", strain.Name, strain.ID)
	return tx.Commit().Error
}
