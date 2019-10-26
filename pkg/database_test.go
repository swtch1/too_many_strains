package tms

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateDatabaseErrorsIfNameNotSet(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	db := DBServer{
		Username: "someusername",
		Password: "somepass",
	}
	assert.Equal(ErrDatabaseNameNotSet, db.ensureDatabase())
}

func TestCreateDatabaseErrorsIfUsernameNotSet(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	db := DBServer{
		Name:     "some_db",
		Password: "somepass",
	}
	assert.Equal(ErrDatabaseUsernameNotSet, db.ensureDatabase())
}
