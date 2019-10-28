//+build integration

package tms

import (
	"github.com/stretchr/testify/assert"
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

func TestCreatingStrainInDB(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name string
	}{
		{"foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: we really need a number generator for these tests that hands out new numbers and does not reuse them
			var ref uint = Unique.Next()
			in := Strain{Name: tt.name, ReferenceID: ref}
			// plug in the db
			in.DB = TestDB
			// set the reference to verify in the new object
			in.ReferenceID = ref
			assert.Nil(in.CreateInDB())

			out := Strain{ReferenceID: ref}
			// populate new object
			TestDB.Model(&out).First(&out)
			assert.Equal(tt.name, out.Name)
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
