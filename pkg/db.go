package tms

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
	"github.com/swtch1/too_many_strains/cmd/database-migration/cli"
)

const DefaultDatabaseName = "so_many_strains"

var (
	ErrDatabaseNameNotSet     = errors.New("database name was not set")
	ErrDatabaseUsernameNotSet = errors.New("database username was not set")
	ErrDatabaseVersionNewer   = errors.New("the actual database version is newer than the desired migration version")
)

// DBServer is the database server where records will be stored and queried.
type DBServer struct {
	// Name of the logical database on the server.
	Name string
	// DBIteration tracks the current iteration of the database schema.  Migrate() must be called on the
	// database to ensure the current schema is up to date with DBIteration.
	DBIteration int16
	Username    string
	Password    string

	// DB is the database connection through which all transactions are brokered.
	DB *gorm.DB
	// isOpen is true when the database connection is open.
	isOpen bool
}

func NewDBServer(name, username, password string) *DBServer {
	srv := &DBServer{
		Username: cli.DatabaseUsername,
		Password: cli.DatabasePassword,
		Name:     DefaultDatabaseName,
	}
	return srv
}

// Open establishes a connection to the database.
func (srv *DBServer) Open() error {
	if err := srv.validateConfig(); err != nil {
		return err
	}

	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8", srv.Username, srv.Password, srv.Name))
	if err != nil {
		return errors.Wrapf(err, "unable to connect to database '%s", srv.Name)
	}
	srv.isOpen = true

	// handle and log errors as they are received
	db.LogMode(false)

	// don't timeout our connection
	db.DB().SetConnMaxLifetime(0)
	srv.DB = db
	return nil
}

// Close terminates the connection with the database.
func (srv *DBServer) Close() error {
	srv.isOpen = false
	return srv.DB.Close()
}

// Migrate will migrate the database to ensure the current version of the schema.
func (srv *DBServer) Migrate() error {
	migrationFailedMsg := fmt.Sprintf("database migration to version %d failed", srv.DBIteration)

	if srv.DBIteration == 0 {
		srv.DBIteration = 1
	}

	if err := srv.ensureDatabase(); err != nil {
		return errors.Wrap(err, migrationFailedMsg)
	}
	if !srv.isOpen {
		if err := srv.Open(); err != nil {
			return errors.Wrapf(err, migrationFailedMsg)
		}
	}

	srv.DB.SingularTable(true)
	srv.DB.AutoMigrate(
		&DatabaseVer{},
		&Strain{},
		&Flavor{},
		&Effect{},
	)
	if err := srv.updateSchemaVersion(); err != nil {
		return errors.Wrap(err, migrationFailedMsg)
	}
	return nil
}

// updateSchemaVersion ensures the database version is set in the database.
func (srv *DBServer) updateSchemaVersion() error {
	var verFromDB DatabaseVer
	if srv.DB.First(&verFromDB); verFromDB.Iteration > srv.DBIteration {
		return ErrDatabaseVersionNewer
	}

	ver := DatabaseVer{Iteration: srv.DBIteration}
	return srv.DB.Assign(ver).FirstOrCreate(&ver).Error
}

// validateConfig ensures that initial values are set so that we can short-circuit configurations
// that could cause more obscure errors down the line.
func (srv *DBServer) validateConfig() error {
	if srv.Name == "" {
		return ErrDatabaseNameNotSet
	}
	if srv.Username == "" {
		return ErrDatabaseUsernameNotSet
	}
	return nil
}

// ensureDatabase idempotently ensures the database exists on the server, and terminates the connection.
func (srv *DBServer) ensureDatabase() (err error) {
	if err := srv.validateConfig(); err != nil {
		return err
	}

	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@/?charset=utf8", srv.Username, srv.Password))
	if err != nil {
		return errors.Wrapf(err, "unable to create database '%s'", srv.Name)
	}
	defer func() {
		err = db.Close()
	}()

	db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", srv.Name))
	return nil
}
