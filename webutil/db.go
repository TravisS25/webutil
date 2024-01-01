package webutil

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//var _ DBInterface = (*sqlx.DB)(nil)

// Bindvar types supported by Rebind, BindMap and BindStruct.
const (
	UNKNOWN = iota
	QUESTION
	DOLLAR
	NAMED
	AT
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

	Sqlite = "sqlite"
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
	ErrEmptyConfigList = errors.New("webutil: empty config list")

	// ErrNoConnection is error returned when there is no
	// connection to database available
	ErrNoConnection = errors.New("webutil: connection could not be established")

	// ErrInvalidDBType is error returned when trying to pass an invalid
	// database type string to function
	ErrInvalidDBType = errors.New("webutil: invalid database type")
)

//////////////////////////////////////////////////////////////////
//------------------------ INTERFACES ---------------------------
//////////////////////////////////////////////////////////////////

type ColScanner interface {
	Columns() ([]string, error)
	Scan(dest ...interface{}) error
	Err() error
}

// Executer implementation should exec against a db
type Executer interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

// Querier implementation is basic querying of a db
type Querier interface {
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	QueryRowxRebind(bindvar int, query string, args ...interface{}) (*sqlx.Row, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryxRebind(bindvar int, query string, args ...interface{}) (*sqlx.Rows, error)
}

// QuerierExec is for querying and executing against db
type QuerierExec interface {
	Querier
	Executer
}

// TxBeginner is for ability to create database transaction
type TxBeginner interface {
	Beginx() (tx *sqlx.Tx, err error)
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
	GetRebind(dest interface{}, bindType int, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	SelectRebind(dest interface{}, bindType int, query string, args ...interface{}) error
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

type Database interface {
	qrm.DB
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

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

func GenUUID() uuid.UUID {
	id, _ := uuid.NewV4()
	return id
}

func GenUUIDStr() string {
	id, _ := uuid.NewV4()
	return id.String()
}

// GetDSNConnStr returns dns connection strings based on settings passed
func GetDSNConnStr(dbCfg DatabaseSetting) string {
	return fmt.Sprintf(
		DBConnStr,
		dbCfg.DBType,
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.DBName,
		dbCfg.SSL,
		dbCfg.SSLMode,
		dbCfg.SSLRootCert,
		dbCfg.SSLKey,
		dbCfg.SSLCert,
	)
}

// NewDBWithDriver works just like NewDB but instead of using
// DatabaseSetting as parameter, we use dataSourceName
func NewDBWithDriver(driverName, dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Open(driverName, dataSourceName)

	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// NewDB is function that returns *sqlx.DB with given DB config
// If db connection fails, returns error
func NewSqlxDB(dbConfig DatabaseSetting, dbType string) (*sqlx.DB, error) {
	dbStr := GetDSNConnStr(dbConfig)
	return NewDBWithDriver(dbType, dbStr)
}

func NoRowsOrDBError(
	w http.ResponseWriter,
	err error,
	clientResp HTTPResponseConfig,
	serverResp HTTPResponseConfig,
) bool {
	if err != nil {
		SetHTTPResponseDefaults(&clientResp, http.StatusNotFound, []byte("Not Found"))
		SetHTTPResponseDefaults(&serverResp, http.StatusInternalServerError, []byte(serverErrTxt))

		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(*clientResp.HTTPStatus)
			w.Write(clientResp.HTTPResponse)
		} else {
			w.WriteHeader(*serverResp.HTTPStatus)
			w.Write(serverResp.HTTPResponse)
		}

		return true
	}

	return false
}

func NoRowsOrDBErrorL(
	w http.ResponseWriter,
	err error,
	logFunc func(err error),
	clientResp HTTPResponseConfig,
	serverResp HTTPResponseConfig,
) bool {
	if err != nil {
		SetHTTPResponseDefaults(&clientResp, http.StatusNotFound, []byte("Not Found"))
		SetHTTPResponseDefaults(&serverResp, http.StatusInternalServerError, []byte(serverErrTxt))

		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(*clientResp.HTTPStatus)
			w.Write(clientResp.HTTPResponse)
		} else {
			w.WriteHeader(*serverResp.HTTPStatus)
			w.Write(serverResp.HTTPResponse)
		}

		if logFunc != nil {
			logFunc(err)
		}

		return true
	}

	return false
}

func DBErrorL(
	w http.ResponseWriter,
	err error,
	logFunc func(err error),
	serverResp HTTPResponseConfig,
) bool {
	if err != nil {
		SetHTTPResponseDefaults(&serverResp, http.StatusInternalServerError, []byte(serverErrTxt))

		w.WriteHeader(*serverResp.HTTPStatus)
		w.Write(serverResp.HTTPResponse)

		if logFunc != nil {
			logFunc(err)
		}

		return true
	}

	return false
}

// QueryCount is used for queries that consist of count in select statement
func QueryCount(db SqlxDB, query string, args ...interface{}) (Count, error) {
	var dest Count
	err := db.Get(&dest, query, args...)
	return dest, err
}

func MapScanner(r ColScanner, dest map[string]interface{}) error {
	columns, values, err := scanColVals(r)

	if err != nil {
		return err
	}

	// getInnerMap takes in colMap and colWords and gets the inner most map and returns it
	getInnerMap := func(colMap map[string]interface{}, colWords []string) map[string]interface{} {
		if len(colWords) == 0 {
			return nil
		}

		var innerMap map[string]interface{}

		for i := 0; i < len(colWords); i++ {
			if i == 0 {
				innerMap = colMap[colWords[i]].(map[string]interface{})
			} else {
				innerMap = innerMap[colWords[i]].(map[string]interface{})
			}
		}

		return innerMap
	}

	for i := range columns {
		colWords := strings.Split(columns[i], ".")

		for idx := range colWords {
			if idx == len(colWords)-1 {
				innerMap := getInnerMap(dest, colWords[:idx])

				if innerMap == nil {
					dest[colWords[idx]] = *(values[i].(*interface{}))
				} else {
					innerMap[colWords[idx]] = *(values[i].(*interface{}))
				}
			} else {
				innerMap := getInnerMap(dest, colWords[:idx])

				if innerMap == nil {
					if _, ok := dest[colWords[idx]]; !ok {
						dest[colWords[idx]] = make(map[string]interface{})
					}
				} else {
					if _, ok := innerMap[colWords[idx]]; !ok {
						innerMap[colWords[idx]] = make(map[string]interface{})
					}
				}
			}
		}
	}

	return r.Err()
}

func scanColVals(r ColScanner) ([]string, []interface{}, error) {
	// ignore r.started, since we needn't use reflect for anything.
	columns, err := r.Columns()
	if err != nil {
		return nil, nil, err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	err = r.Scan(values...)
	return columns, values, err
}
