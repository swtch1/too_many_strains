package tms

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestParsingStrainsFromJsonGetsName(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		name        string
		strainsJSON string
		expNames    []string
	}{
		{"basic", strainsJSON, []string{"foo", "bar"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strains, err := ParseStrains(TestStrainsReader{})
			assert.Nil(err)

			// ensure each item we expect is in at least one of the strains parsed
			for _, name := range tt.expNames {
				var match bool
				for _, strain := range strains {
					if strain.Name == name {
						match = true
					}
				}
				assert.True(match, "failed to find expected name '%s' in strain", name)
			}
		})
	}
}

func TestParsingStrainsFromJsonGetsRace(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		name        string
		strainsJSON string
		expRaces    []string
	}{
		{"basic", strainsJSON, []string{"r1", "r2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strains, err := ParseStrains(TestStrainsReader{})
			assert.Nil(err)

			// ensure each item we expect is in at least one of the strains parsed
			for _, race := range tt.expRaces {
				var match bool
				for _, strain := range strains {
					if strain.Race == race {
						match = true
					}
				}
				assert.True(match, "failed to find expected race '%s' in strain", race)
			}
		})
	}
}

func TestParsingStrainsFromJsonGetsFlavors(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		name        string
		strainsJSON string
		expFlavors  []string
	}{
		{"basic", strainsJSON, []string{"f1", "f2", "f3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strains, err := ParseStrains(TestStrainsReader{})
			assert.Nil(err)

			// ensure each item we expect is in at least one of the strains parsed
			for _, flavor := range tt.expFlavors {
				var match bool
				for _, strain := range strains {
					for _, f := range strain.Flavors {
						if f == flavor {
							match = true
						}
					}
				}
				assert.True(match, "failed to find expected flavor '%s' in strain", flavor)
			}
		})
	}
}

func TestParsingStrainsFromJsonGetsPositiveEffects(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tests := []struct {
		name        string
		strainsJSON string
		expEffects  []string
	}{
		{"basic", strainsJSON, []string{"pos1", "pos2", "pos3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strains, err := ParseStrains(TestStrainsReader{})
			assert.Nil(err)

			// ensure each item we expect is in at least one of the strains parsed
			for _, effect := range tt.expEffects {
				var match bool
				for _, strain := range strains {
					for _, e := range strain.Effects.Positive {
						if e == effect {
							match = true
						}
					}
				}
				assert.True(match, "failed to find expected effect '%s' in strain", effect)
			}
		})
	}
}

// TestStrainsReader provides a simple reader for testing strain JSON.
type TestStrainsReader struct{}

func (s TestStrainsReader) Read(p []byte) (int, error) {
	r := bytes.NewReader([]byte(strainsJSON))
	n, err := r.Read(p)
	if err != nil {
		return 0, err
	}
	return n, io.EOF
}

var strainsJSON = `
[
    {
		"name": "foo",
		"id": 1,
		"race": "r1",
		"flavors": [
			"f1",
			"f2"
		],
		"effects": {
			"positive": [
				"pos1",
				"pos2"
			],
			"negative": [
				"neg1"
			],
			"medical": [
				"med1"
			]
		}
	},
	{
		"name": "bar",
		"id": 2,
		"race": "r2",
		"flavors": [
			"f1",
			"f3"
		],
		"effects": {
			"positive": [
				"pos3"
			],
			"negative": [
				"neg2"
			],
			"medical": [
				"med2"
			]
		}
	}
]
`
