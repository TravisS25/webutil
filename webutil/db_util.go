package webutil

//go:generate mockgen -source=db_util.go -destination=../webutilmock/db_util_mock.go -package=webutilmock
//go:generate mockgen -source=db_util.go -destination=db_util_mock_test.go -package=webutil

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var _ DBInterface = (*sqlx.DB)(nil)

type runQuery struct {
	Query string
	Args  []interface{}
}

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
	DBConnStr = "%s://%s:%s@%s:%d/%s?ssl=%v&sslmode=%s&sslrootcert=%s&sslkey=%s&sslcert=%s"
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

	// ErrInvalidDBType is error returned when trying to pass an invalid
	// database type string to function
	ErrInvalidDBType = errors.New("webutil: invalid database type")
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

// DBInterfaceRecover implements setting DBInterface
// to struct that implements SetDBInterface
// This is generally used in apis to recover from
// database failure
type DBInterfaceRecover interface {
	SetDBInterface(DBInterface)
}

// EntityRecover implements setting Entity
// to struct that implements SetEntity
// This is generally used in form validators to
// recover from database failure
type EntityRecover interface {
	SetEntity(Entity)
}

// TxBeginner is for ability to create database transaction
type TxBeginner interface {
	Begin() (tx *sql.Tx, err error)
}

// QuerierTx is used for basic querying but also
// need transaction
type QuerierTx interface {
	TxBeginner
	QuerierExec
}

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
	TxBeginner
}

//////////////////////////////////////////////////////////////////
//-------------------------- TYPES ----------------------------
//////////////////////////////////////////////////////////////////

// RecoverDB is func that is passed to functions to try
// to recover from db failure
// This implementation can be used for any db but is made in
// mind for distributed databases ie. CockroachDB
type RecoverDB func(err error) (*sqlx.DB, error)

// RetryDB implementation should query database that has
// recovered from a failure and return whether you get
// an error or not
type RetryDB func(DBInterface) error

//////////////////////////////////////////////////////////////////
//---------------------- CONFIG STRUCTS ------------------------
//////////////////////////////////////////////////////////////////

// Count is used to retrieve from count queries
type Count struct {
	Total int `json:"total"`
}

//////////////////////////////////////////////////////////////////
//------------------------ FUNCTIONS ---------------------------
//////////////////////////////////////////////////////////////////

// NewDB is function that returns *sqlx.DB with given DB config
// If db connection fails, returns error
func NewDB(dbConfig DatabaseSetting, dbType string) (*sqlx.DB, error) {
	dbStr := fmt.Sprintf(
		DBConnStr,
		dbType,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
		dbConfig.SSL,
		dbConfig.SSLMode,
		dbConfig.SSLRootCert,
		dbConfig.SSLKey,
		dbConfig.SSLCert,
	)

	db, err := sqlx.Open(dbType, dbStr)

	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// NewDBWithList is function that returns *sqlx.DB with given slice DB config
// If no db connection can be established with given list, ErrNoConnection is returned
func NewDBWithList(dbConfigList []DatabaseSetting, dbType string) (*sqlx.DB, error) {
	if len(dbConfigList) == 0 {
		return nil, ErrEmptyConfigList
	}

	for _, v := range dbConfigList {
		newDB, err := NewDB(v, dbType)

		if err == nil {
			return newDB, nil
		}
	}

	return nil, ErrNoConnection
}

// HasDBError takes passed error and determines what to write
// back to client depending on settings set in config
func HasDBError(w http.ResponseWriter, err error, config ServerErrorConfig) bool {
	defaultDBErrors(&config)
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
		w.WriteHeader(*config.ClientErrorResponse.HTTPStatus)
		w.Write(config.ClientErrorResponse.HTTPResponse)
		return true
	}

	return dbError(w, err, config)
}

// PopulateDatabaseTables populates "database_table" in a database which
// should reference the tables in the database
//
// The dbTables parameter should be a map where the key is name of database
// table and the value is the string representation of the table
func PopulateDatabaseTables(db DBInterface, dbType string, dbTables map[string]string) error {
	var err error
	var bindVar int
	var publicQuery, columnQuery string
	var args []interface{}
	var query string

	if len(dbTables) == 0 {
		return errors.New("can not have empty inclusion map")
	}

	dbQuery := `select name from public.database_table where name = ?;`
	dbTableInsertQuery :=
		`
	insert into database_table(name, display_name, column_name)
	values (?, ?, ?);
	`

	switch dbType {
	case Postgres:
		bindVar = sqlx.DOLLAR
		columnQuery =
			`
		SELECT 
			column_name
		FROM 
			information_schema.columns
		WHERE 
			table_schema = 'public'
		AND 
			table_name = ?
		`
		publicQuery =
			`
		select
			tablename
		from
			pg_tables
		where
			schemaname = 'public'
		`
	case MySQL:
		bindVar = sqlx.QUESTION
	default:
		return ErrInvalidDBType
	}

	inclusionRower, err := db.Query(publicQuery)

	if err != nil {
		return errors.Wrap(err, "")
	}

	invalidInclusionTables := make([]string, 0)
	runQueries := make([]runQuery, 0)

	for inclusionRower.Next() {
		var tableName, filler string
		err = inclusionRower.Scan(
			&tableName,
		)

		if err != nil {
			//tx.Rollback()
			return errors.Wrap(err, "")
		}

		query = dbQuery

		if query, args, err = InQueryRebind(bindVar, query, tableName); err != nil {
			return errors.Wrap(err, "")
		}

		row := db.QueryRow(query, args...)
		err = row.Scan(&filler)

		if err != nil {
			if err == sql.ErrNoRows {
				if val, ok := dbTables[tableName]; ok {
					var columnName string
					query := columnQuery

					if query, args, err = InQueryRebind(bindVar, query, tableName); err != nil {
						return errors.Wrap(err, "")
					}

					rows, err := db.Query(query, args...)

					if err != nil {
						return errors.Wrap(err, "")
					}

					containsName := false

					for rows.Next() {
						if err = rows.Scan(&columnName); err != nil {
							return errors.Wrap(err, "")
						}

						if val == columnName {
							containsName = true
						}
					}

					if !containsName {
						errStr := fmt.Sprintf("table %s does not contain column '%s'", tableName, val)
						return errors.New(errStr)
					}

					displayName := strings.Title(strings.Replace(tableName, "_", " ", -1))
					query = dbTableInsertQuery

					if query, args, err = InQueryRebind(
						bindVar,
						query,
						tableName,
						displayName,
						val,
					); err != nil {
						return errors.Wrap(err, "")
					}

					runQueries = append(runQueries, runQuery{Query: query, Args: args})

					// if _, err = tx.Exec(query, args...); err != nil {
					// 	return errors.Wrap(err, "")
					// }
				} else {
					invalidInclusionTables = append(invalidInclusionTables, tableName)
				}
			} else {
				//tx.Rollback()
				return errors.Wrap(err, "")
			}
		}
	}

	if len(invalidInclusionTables) > 0 {
		errStr := "Table(s): \n"

		for _, v := range invalidInclusionTables {
			errStr += "\t" + v + "\n"
		}

		errStr += "are not in dbTables\n"
		//tx.Rollback()
		return errors.Wrap(errors.New(errStr), "")
	}

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	for _, v := range runQueries {
		if _, err = tx.Exec(v.Query, v.Args...); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
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
			if db, err := config.RecoverDB(err); err == nil {
				if config.DBInterfaceRecover != nil {
					config.DBInterfaceRecover.SetDBInterface(db)
					if config.RetryDB != nil {
						if err = config.RetryDB(db); err == nil {
							return false
						}
					}
				}
			}
		}

		w.WriteHeader(*config.ServerErrorResponse.HTTPStatus)
		w.Write(config.ServerErrorResponse.HTTPResponse)
		return true
	}

	return false
}
