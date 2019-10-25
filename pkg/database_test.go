package tms

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateDatabaseErrorsIfUsernameNotSet(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	db := Database{
		Name:     "some_db",
		Password: "somepass",
	}
	assert.Equal(ErrDatabaseUsernameNotSet, db.CreateDatabase())
}
