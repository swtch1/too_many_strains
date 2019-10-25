package tms

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
)

const DefaultDatabaseName = "so_many_strains"

var (
	ErrDatabaseUsernameNotSet = errors.New("database username was not set")
)

type Database struct {
	Name     string
	Username string
	Password string
	db       *gorm.DB
}

// CreateDatabase idempotently creates the Name name, and closes the Name connection.
func (db *Database) CreateDatabase() (err error) {
	if db.Username == "" {
		return ErrDatabaseUsernameNotSet
	}

	tmpDb, err := gorm.Open("mysql", "root:password@/?charset=utf8")
	if err != nil {
		return errors.Wrapf(err, "unable to create database '%s'", db.Name)
	}
	defer func() {
		err = tmpDb.Close()
	}()

	tmpDb.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", "so_many_strains"))
	//tmpDb.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXIST %s;", db.Name))
	return nil
}

// Connect establishes a connection to the database.
func (db *Database) Connect() error {
	gdb, err := gorm.Open("mysql", fmt.Sprintf("root:password@/%s?charset=utf8", db.Name))
	if err != nil {
		return errors.Wrapf(err, "unable to connect to database '%s", db.Name)
	}
	db.db = gdb
	return nil
}

// Close terminates the connection with the database.
func (db *Database) Close() error {
	return db.db.Close()
}
