//+build integration

package tms

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

// Unique hands out a number that nobody else is using
var Unique uniqueNum

func init() {
	// set integration here so it is only set when integration tests are run.
	Integration = true
}

func TestMigrateDatabase(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		tableName string
	}{
		{"database_ver"},
		{"strain"},
		{"effect"},
		{"strain_effects"},
		{"flavor"},
		{"strain_flavors"},
	}

	for _, tt := range tests {
		t.Run(tt.tableName, func(t *testing.T) {
			// this is done in the test main setup, but that's too far from here to count on it always
			dbSrv := NewDBServer(TestDatabaseName, "root", "password")
			assert.Nil(dbSrv.Migrate())
			assert.Nil(dbSrv.Open())
			assert.True(dbSrv.DB.HasTable(tt.tableName))
		})
	}
}

func TestCreatingStrainInDBUpdatesName(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name string
	}{
		{"foo"},
		{"bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref uint = Unique.Next()
			in := Strain{Name: tt.name, ReferenceID: ref}
			// plug in the db
			in.DB = TestDB
			// set the reference to verify in the new object
			in.ReferenceID = ref
			assert.Nil(in.CreateInDB())

			out := Strain{ReferenceID: ref}
			// populate new object
			TestDB.Where(&out).First(&out)
			assert.Equal(tt.name, out.Name)
		})
	}
}

func TestCreatingStrainInDBUpdatesRace(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		race string
	}{
		{"indica"},
		{"sativa"},
	}

	for _, tt := range tests {
		t.Run(tt.race, func(t *testing.T) {
			var ref uint = Unique.Next()
			in := Strain{Name: tt.race, Race: tt.race, ReferenceID: ref}
			// plug in the db
			in.DB = TestDB
			// set the reference to verify in the new object
			in.ReferenceID = ref
			assert.Nil(in.CreateInDB())

			out := Strain{ReferenceID: ref}
			// populate new object
			TestDB.Where(&out).First(&out)
			assert.Equal(tt.race, out.Race)
		})
	}
}

func TestSavingStrainUpdatesName(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		name string
	}{
		{"foo"},
		{"bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref uint = Unique.Next()
			in := Strain{Name: "first", ReferenceID: ref}
			// plug in the db
			in.DB = TestDB
			// set the reference to verify in the new object
			in.ReferenceID = ref
			assert.Nil(in.CreateInDB())

			in.Name = tt.name
			in.SaveInDB()

			out := Strain{ReferenceID: ref}
			// populate new object
			TestDB.Where(&out).First(&out)
			assert.Equal(tt.name, out.Name)
		})
	}
}

func TestGettingStrainFromFlavorsFromDBByID(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		name       string
		strain     Strain
		expFlavors []string
	}{
		{
			"one_flavor",
			Strain{
				Name:    "foo",
				Race:    "xyz",
				Flavors: []Flavor{{Name: "yummy"}},
			},
			[]string{"yummy"},
		},
		{
			"many_flavors",
			Strain{
				Name:    "foo",
				Race:    "xyz",
				Flavors: []Flavor{{Name: "yummy"}, {Name: "tasty"}},
			},
			[]string{"yummy", "tasty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.strain.Name, func(t *testing.T) {
			ref := Unique.Next()
			tt.strain.DB = TestDB

			// push strain with flavors in to DB
			tt.strain.ReferenceID = ref
			assert.Nil(tt.strain.CreateInDB())

			out := Strain{ReferenceID: ref}
			out.DB = TestDB
			flavors, err := out.FlavorsFromDBByID()
			assert.Nil(err)
			for _, expected := range tt.expFlavors {
				var match bool
				for _, f := range flavors {
					if f.Name == expected {
						match = true
					}
				}
				assert.True(match, "did not find expected flavor %s in returned flavors", expected)
			}
		})
	}
}

type uniqueNum struct {
	number uint
	lock   sync.Mutex
}

func (n *uniqueNum) Next() uint {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.number++
	return n.number
}
