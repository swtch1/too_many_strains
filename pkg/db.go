package tms

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultDatabaseName = "so_many_strains"
	DBConnectOptions    = "charset=utf8&parseTime=True&loc=Local"
)

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
	DBIteration uint
	Username    string
	Password    string

	// DB is the database connection through which all transactions are brokered.
	DB *gorm.DB
	// isOpen is true when the database connection is open.
	isOpen bool
}

func NewDBServer(name, username, password string) *DBServer {
	srv := &DBServer{
		Username: username,
		Password: password,
		Name:     name,
	}
	return srv
}

// Open establishes a connection to the database.
func (srv *DBServer) Open() error {
	if err := srv.validateConfig(); err != nil {
		return err
	}

	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?%s", srv.Username, srv.Password, srv.Name, DBConnectOptions))
	if err != nil {
		return errors.Wrapf(err, "unable to connect to database '%s", srv.Name)
	}
	srv.isOpen = true
	db.SingularTable(true)
	// handle and log errors as they are received
	db.LogMode(false)
	// don't timeout our connection
	db.DB().SetConnMaxLifetime(0)

	if err := db.DB().Ping(); err != nil {
		return errors.Wrapf(err, "unable to ping database '%s' after connection", srv.Name)
	}

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
	if srv.DBIteration == 0 {
		srv.DBIteration = 1
	}

	if err := srv.ensureDatabase(); err != nil {
		return errors.Wrap(err, "unable to create database")
	}
	if !srv.isOpen {
		if err := srv.Open(); err != nil {
			return errors.Wrapf(err, "unable to open database connection")
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
		return errors.Wrapf(err, "unable to update database to iteration %d", srv.DBIteration)
	}
	return nil
}

// updateSchemaVersion ensures the database version is set in the database.
func (srv *DBServer) updateSchemaVersion() error {
	var verFromDB DatabaseVer
	srv.DB.Last(&verFromDB)
	switch {
	case verFromDB.Iteration > srv.DBIteration:
		log.Debugf("version from database %d is greater than current iteration %d", verFromDB.Iteration, srv.DBIteration)
		return ErrDatabaseVersionNewer
	case verFromDB.Iteration == srv.DBIteration:
		log.Infof("database version %d is already the latest", verFromDB.Iteration)
		return nil
	default:
		log.Infof("database version %d is behind desired version %d, updating version tracking table", verFromDB.Iteration, srv.DBIteration)
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

	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@/?%s", srv.Username, srv.Password, DBConnectOptions))
	if err != nil {
		return errors.Wrapf(err, "unable to create database '%s'", srv.Name)
	}
	defer func() {
		err = db.Close()
	}()

	if err := db.DB().Ping(); err != nil {
		return errors.Wrapf(err, "unable to ping database '%s' after connection", srv.Name)
	}

	db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", srv.Name))
	return nil
}
