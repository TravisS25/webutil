package webutil

//go:generate mockgen -source=db_util.go -destination=../webutilmock/db_util_mock.go -package=webutilmock
//go:generate mockgen -source=db_util.go -destination=db_util_mock_test.go -package=webutil

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//////////////////////////////////////////////////////////////////
//------------------------ SSL MODES ---------------------------
//////////////////////////////////////////////////////////////////

const (
	// SSLDisable represents disable value for "sslmode" query parameter
	SSLDisable = "disable"

	// SSLRequire represents require value for "sslmode" query parameter
	SSLRequire = "require"

	// SSLVerifyCA represents verify-ca value for "sslmode" query parameter
	SSLVerifyCA = "verify-ca"

	// SSLVerifyFull represents verify-full value for "sslmode" query parameter
	SSLVerifyFull = "verify-full"
)

//////////////////////////////////////////////////////////////////
//---------------------- DATABASE TYPES ------------------------
//////////////////////////////////////////////////////////////////

const (
	// Postgres is used in NewDB function to initialize
	// postgres db connection
	Postgres = "postgres"

	// MySQL is used in NewDB function to initialize
	// mysql db connection
	MySQL = "mysql"
)

//////////////////////////////////////////////////////////////////
//------------------------ STRING CONSTS -----------------------
//////////////////////////////////////////////////////////////////

const (
	// DBConnStr is default connection string
	DBConnStr = "host=%s user=%s password=%s dbname=%s port=%d sslmode=%s"
)

//////////////////////////////////////////////////////////////////
//------------------------ ERROR TYPES -------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrEmptyConfigList is error returned when trying to recover
	// from database error and there is no backup configs set up
	ErrEmptyConfigList = errors.New("empty config list")

	// ErrNoConnection is error returned when there is no
	// connection to database available
	ErrNoConnection = errors.New("connection could not be established")
)

//////////////////////////////////////////////////////////////////
//------------------------ INTERFACES ---------------------------
//////////////////////////////////////////////////////////////////

// Executer implementation should exec against a db
type Executer interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

// Querier implementation is basic querying of a db
type Querier interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// QuerierExec is for querying and executing against db
type QuerierExec interface {
	Querier
	Executer
}

// Scanner will scan row returned from database
// type Scanner interface {
// 	Scan(dest ...interface{}) error
// }

// // Rower loops through rows returns from database with
// // abilty to scan each row
// type Rower interface {
// 	Scanner
// 	Next() bool
// 	Columns() ([]string, error)
// }

// // Tx is for transaction related queries
// type Tx interface {
// 	QuerierExec
// 	SqlxDB
// 	Commit() error
// 	Rollback() error
// }

// Tx is for ability to create database transaction
type Tx interface {
	Begin() (tx *sql.Tx, err error)
	Commit(tx *sql.Tx) error
}

// QuerierTx is used for basic querying but also
// need transaction
type QuerierTx interface {
	Tx
	QuerierExec
}

// // QuerierExec allows to query rows but also exec statement against database
// type QuerierExec interface {
// 	Querier
// 	//Exec(string, ...interface{}) (sql.Result, error)
// }

// SqlxDB uses the sqlx library methods Get and Select for ability to
// easily query results into structs
type SqlxDB interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

// Entity is mainly used for FormValidation
type Entity interface {
	QuerierExec
	SqlxDB
}

// DBInterface is the main interface that should be used in your
// request handler functions
type DBInterface interface {
	Entity
	Tx
	// /Recover
}

// Recover implementation is used to recover from db failure
type Recover interface {
	RecoverError(err error) (*DB, error)
}

// // RecoverQuerier is used to be able to do basic querying
// // and recover from db failure if neccessary
// type RecoverQuerier interface {
// 	Querier
// 	Recover
// }

//////////////////////////////////////////////////////////////////
//-------------------------- TYPES ----------------------------
//////////////////////////////////////////////////////////////////

// RecoverDB is func that is passed to functions to try
// to recover from db failure
// This implementation can be used for any db but is made in
// mind for distributed databases ie. CockroachDB
type RecoverDB func(err error) error

// RecoverDBQuery is func used to be able to recovery from db failure
// and to re-execute query without any down time
type RecoverDBQuery func(err error, retryFn func(db Entity, query string, args ...interface{}) error) error

// RecoverDBTransaction is func used to be able to recovery from db failure
// and to re-execute transaction without any down time
type RecoverDBTransaction func(err error, retryFn func(db QuerierTx, queries []RunQuery) error) error

//////////////////////////////////////////////////////////////////
//---------------------- CONFIG STRUCTS ------------------------
//////////////////////////////////////////////////////////////////

// Count is used to retrieve from count queries
type Count struct {
	Total int `json:"total"`
}

type RunQuery struct {
	Query string
	Args  []interface{}
}

//////////////////////////////////////////////////////////////////
//-------------------------- STRUCTS ---------------------------
//////////////////////////////////////////////////////////////////

// DB extends sqlx.DB with some extra functions
type DB struct {
	*sqlx.DB
	dbConfigList  []DatabaseSetting
	currentConfig DatabaseSetting
	dbType        string
}

// RecoverError will check if given err is not nil and if it is
// it will loop through dbConfigList, if any, and try to establish
// a new connection with a different database
//
// This function should be used if you have a distributed type database
// i.e. CockroachDB and don't want any interruptions if a node goes down
//
// This function does not check what type of err is passed, just checks
// if err is nil or not so it's up to user to use appropriately; however
// we do a quick ping check just to make sure db is truely down
//
// This function is NOT thread safe so one should create a mutex around
// this function when trying to recover from error
func (db *DB) RecoverError(err error) (*DB, error) {
	if err != nil {
		dbInfo := fmt.Sprintf(
			DBConnStr,
			db.currentConfig.Host,
			db.currentConfig.User,
			db.currentConfig.Password,
			db.currentConfig.DBName,
			db.currentConfig.Port,
			db.currentConfig.SSLMode,
		)

		_, err = db.Driver().Open(dbInfo)

		if err != nil {
			fmt.Printf("connection officially failed\n")
			if len(db.dbConfigList) == 0 {
				return nil, ErrEmptyConfigList
			}

			newDB, err := NewDBWithList(db.dbConfigList, db.dbType)

			if err != nil {
				return nil, ErrNoConnection
			}

			return newDB, err
		}

		return db, nil
	}
	return db, nil
}

//////////////////////////////////////////////////////////////////
//------------------------ FUNCTIONS ---------------------------
//////////////////////////////////////////////////////////////////

// NewDB is function that returns *DB with given DB config
// If db connection fails, returns error
func NewDB(dbConfig DatabaseSetting, dbType string) (*DB, error) {
	dbInfo := fmt.Sprintf(
		DBConnStr,
		dbConfig.Host,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.DBName,
		dbConfig.Port,
		dbConfig.SSLMode,
	)

	fmt.Printf("conn: %s\n", dbInfo)

	db, err := sqlx.Open(dbType, dbInfo)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{DB: db, dbType: dbType}, nil
}

// NewDBWithList is function that returns *DB with given slice DB config
// If no db connection can be established with given list, ErrNoConnection is returned
func NewDBWithList(dbConfigList []DatabaseSetting, dbType string) (*DB, error) {
	if len(dbConfigList) == 0 {
		return nil, ErrEmptyConfigList
	}

	for _, v := range dbConfigList {
		newDB, err := NewDB(v, dbType)

		if err == nil {
			newDB.dbConfigList = dbConfigList
			newDB.currentConfig = v
			return newDB, nil
		}
	}

	return nil, ErrNoConnection
}

// HasDBError takes passed error and determines what to write
// back to client depending on settings set in config
func HasDBError(w http.ResponseWriter, err error, config ServerErrorConfig) bool {
	SetHTTPResponseDefaults(
		&config.ServerErrorResponse,
		http.StatusInternalServerError,
		[]byte(ErrServer.Error()),
	)
	return dbError(w, err, config)
}

// HasNoRowsOrDBError takes passed error and determines what to write
// back to client depending on settings set in config
//
// If error is "sql.ErrNoRows", then another response is written
// to client based on config passed
func HasNoRowsOrDBError(w http.ResponseWriter, err error, config ServerErrorConfig) bool {
	defaultDBErrors(&config)

	if err == sql.ErrNoRows {
		http.Error(
			w,
			string(config.ClientErrorResponse.HTTPResponse),
			*config.ClientErrorResponse.HTTPStatus,
		)
		return true
	}

	return dbError(w, err, config)
}

// QueryCount is used for queries that consist of count in select statement
func QueryCount(db SqlxDB, query string, args ...interface{}) (*Count, error) {
	var dest Count
	err := db.Get(&dest, query, args...)
	return &dest, err
}

func defaultDBErrors(config *ServerErrorConfig) {
	SetHTTPResponseDefaults(&config.ClientErrorResponse, http.StatusNotFound, []byte("Not Found"))
	SetHTTPResponseDefaults(
		&config.ServerErrorResponse,
		http.StatusInternalServerError,
		[]byte(ErrServer.Error()),
	)
}

func dbError(w http.ResponseWriter, err error, config ServerErrorConfig) bool {
	if err != nil {
		if config.RecoverDB != nil {
			if err = config.RecoverDB(err); err != nil {
				http.Error(
					w,
					string(config.ServerErrorResponse.HTTPResponse),
					*config.ServerErrorResponse.HTTPStatus,
				)
				return true
			}
		} else {
			return true
		}
	}

	return false
}
