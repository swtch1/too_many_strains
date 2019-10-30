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
	t.Parallel()
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
	t.Parallel()
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

func TestGettingStrainFromFlavorsFromDBByID(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		name       string
		baseStrain Strain
		flavors    []Flavor
	}{
		{
			"one_flavor",
			Strain{
				Name: "foo",
				Race: "xyz",
			},
			[]Flavor{{Name: "yummy"}},
		},
		{
			"many_flavors",
			Strain{
				Name: "bar",
				Race: "xyz",
			},
			[]Flavor{{Name: "yummy"}, {Name: "tasty"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := Unique.Next()
			tt.baseStrain.DB = TestDB

			// push strain with flavors in to DB
			tt.baseStrain.ReferenceID = ref
			tt.baseStrain.Flavors = tt.flavors
			assert.Nil(tt.baseStrain.CreateInDB())

			out := Strain{DB: TestDB}
			flavors, err := out.FlavorsFromDBByRefID(ref)
			assert.Nil(err)
			for _, f := range tt.flavors {
				var match bool
				for _, fFromDB := range flavors {
					if f.Name == fFromDB.Name {
						match = true
					}
				}
				assert.True(match, "did not find expected flavor '&s' in returned flavors", f.Name)
			}
		})
	}
}

func TestGettingEffectsFromFlavorsFromDBByID(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		name       string
		baseStrain Strain
		effects    []Effect
	}{
		{
			"one_effect",
			Strain{
				Name: "foo",
				Race: "xyz",
			},
			[]Effect{{Name: "happy", Category: "positive"}},
		},
		{
			"many_effects",
			Strain{
				Name: "bar",
				Race: "xyz",
			},
			[]Effect{{Name: "not_so_happy", Category: "negative"}, {Name: "nausea", Category: "medical"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := Unique.Next()
			tt.baseStrain.DB = TestDB

			// push strain with effects in to DB
			tt.baseStrain.ReferenceID = ref
			tt.baseStrain.Effects = tt.effects
			assert.Nil(tt.baseStrain.CreateInDB())

			out := Strain{DB: TestDB}
			effectsFromDB, err := out.EffectsFromDBByRefID(ref)
			assert.Nil(err)
			for _, e := range tt.effects {
				var match bool
				for _, eFromDB := range effectsFromDB {
					if e.Name == eFromDB.Name && e.Category == eFromDB.Category {
						match = true
					}
				}
				assert.True(match, "did not find expected effect with name %s and category %s in returned effects", e.Name, e.Category)
			}
		})
	}
}

func TestGettingStrainFromDBByIDHasCorrectFlavors(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		baseStrain Strain
		flavors    []Flavor
		effects    []Effect
	}{
		{
			Strain{
				Name: "foo",
				Race: "xyz",
			},
			[]Flavor{{Name: "so_gud"}},
			[]Effect{{Name: "happy", Category: "positive"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.baseStrain.Name, func(t *testing.T) {
			ref := Unique.Next()
			tt.baseStrain.DB = TestDB

			tt.baseStrain.Flavors = tt.flavors
			tt.baseStrain.Effects = tt.effects

			// push strain with effects in to DB
			tt.baseStrain.ReferenceID = ref
			assert.Nil(tt.baseStrain.CreateInDB())

			out := Strain{DB: TestDB}
			err := out.FromDBByRefID(ref)
			assert.Nil(err)
			for _, f := range tt.flavors {
				var match bool
				for _, fFromDB := range out.Flavors {
					if f.Name == fFromDB.Name {
						match = true
					}
				}
				assert.True(match, "did not find expected flavor with name '&s' in returned flavors", f.Name)
			}
		})
	}
}

func TestGettingStrainFromDBByIDHasCorrectEffects(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		baseStrain Strain
		flavors    []Flavor
		effects    []Effect
	}{
		{
			Strain{
				Name: "foo",
				Race: "xyz",
			},
			[]Flavor{{Name: "so_gud"}},
			[]Effect{{Name: "happy", Category: "positive"}},
		},
		{
			Strain{
				Name: "bar",
				Race: "zzz",
			},
			[]Flavor{{Name: "yums"}},
			[]Effect{{Name: "sleepy", Category: "medical"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.baseStrain.Name, func(t *testing.T) {
			ref := Unique.Next()
			tt.baseStrain.DB = TestDB

			tt.baseStrain.Flavors = tt.flavors
			tt.baseStrain.Effects = tt.effects

			// push strain with effects in to DB
			tt.baseStrain.ReferenceID = ref
			assert.Nil(tt.baseStrain.CreateInDB())

			out := Strain{DB: TestDB}
			err := out.FromDBByRefID(ref)
			assert.Nil(err)
			for _, e := range tt.effects {
				var match bool
				for _, eFromDB := range out.Effects {
					if e.Name == eFromDB.Name && e.Category == eFromDB.Category {
						match = true
					}
				}
				assert.True(match, "did not find expected effect with name '&s' and category %s in returned effects", e.Name, e.Category)
			}
		})
	}
}

func TestReplacingStrainReprInDBDeletesSuperfluousFlavors(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		name string
		repr StrainRepr
	}{
		{
			"basic",
			StrainRepr{
				Name:    "foo",
				Race:    "",
				Flavors: []string{"bubblegunm"},
				Effects: struct {
					Positive []string `json:"positive"`
					Negative []string `json:"negative"`
					Medical  []string `json:"medical"`
				}{Positive: []string{"sogud"}, Negative: []string{"notgud"}},
				DB: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := Unique.Next()
			tt.repr.ID = ref
			tt.repr.DB = TestDB

			// add a new flavor
			newFlav := "new_flavor"
			tt.repr.Flavors = append(tt.repr.Flavors, newFlav)
			assert.Nil(tt.repr.ReplaceInDB())

			out := Strain{DB: TestDB}
			err := out.FromDBByRefID(ref)
			assert.Nil(err)

			var match bool
			for _, fFromDB := range out.Flavors {
				if newFlav == fFromDB.Name {
					match = true
				}
			}
			assert.True(match, "expected test flavor was not found")

			// now remove the flavor to ensure it's cleaned from the table
			var flavorsMinusNewFlav []string
			for _, f := range tt.repr.Flavors {
				if f != newFlav {
					flavorsMinusNewFlav = append(flavorsMinusNewFlav, f)
				}
			}
			tt.repr.Flavors = flavorsMinusNewFlav

			// perform replacement again which should remove test flavor
			assert.Nil(tt.repr.ReplaceInDB())

			// get results from DB again
			out = Strain{DB: TestDB}
			err = out.FromDBByRefID(ref)
			assert.Nil(err)

			// test for new flavor, which we do not expect to see
			match = false
			for _, fFromDB := range out.Flavors {
				if newFlav == fFromDB.Name {
					match = true
				}
			}
			assert.False(match, "expected not to find test flavor in DB")
		})
	}
}

func TestGettingStrainFromDBByIDWithNoRecordReturnsError(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	var refThatDoesNotExist uint = 9999999999
	out := Strain{DB: TestDB}
	err := out.FromDBByRefID(refThatDoesNotExist)
	assert.Equal(ErrNotExists, err)
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
