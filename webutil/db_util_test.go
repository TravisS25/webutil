package webutil

import (
	"bytes"
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var _ DBInterfaceRecover = (*testAPI)(nil)

var (
	dbMutex sync.Mutex
	//db      *DB

	errDB     = errors.New("db error")
	recoverDB = func(err error) (*DB, error) {
		return &DB{}, nil
	}
	failedRecoverDB = func(err error) (*DB, error) {
		return nil, err
	}
)

type channel struct {
	ready chan struct{}
}

func newChannel() *channel {
	return &channel{ready: make(chan struct{})}
}

func (m *channel) Stop() {
	close(m.ready)
}

func (m *channel) Get() <-chan struct{} {
	return m.ready
}

type testAPI struct {
	DB DBInterface
}

func (f *testAPI) SetDBInterface(db DBInterface) {
	f.DB = db
}

func (f *testAPI) Index(w http.ResponseWriter, r *http.Request) {
	var db *DB
	var err error
	//var recoverFn func(err error) (*DB, error)
	var removeFn func() error

	//initDB()
	if db, _, removeFn, err = dbRecoverSetup(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("couldn't past recover set up"))
		return
	}

	f.DB = db

	if err = removeFn(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	retry := func(innerDB DBInterface) error {
		_, err = innerDB.Query(testConf.DBResetConf.ValidateQuery)
		return err
	}

	if err = retry(f.DB); err == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("couldn't retry"))
		return
	}

	conf := ServerErrorConfig{
		RecoverConfig: RecoverConfig{
			RecoverDB: func(err error) (*DB, error) {
				if err != nil {
					dbMutex.Lock()
					defer dbMutex.Unlock()
					db, err = db.RecoverError()
					return db, err
				}

				return db, nil
			},
			RetryDB:            retry,
			DBInterfaceRecover: f,
		},
	}

	if HasDBError(w, err, conf) {
		return
	}

	if _, err = f.DB.Query(testConf.DBResetConf.ValidateQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("didn't recovery right"))
	}
}

// dbRecoverSetup sets up and returns new db with implemented RecoverDB function
func dbRecoverSetup() (*DB, func(err error) (*DB, error), func() error, error) {
	var db *DB
	var err error
	var chosenPort int
	var dockerName string
	var conf teardownConfig
	//var chosenPortInt int

	rand.Seed(time.Now().UnixNano())
	minPort := 3000
	maxPort := 30000
	portAttempts := 0

	for {
		chosenPort = rand.Intn(maxPort-minPort+1) + minPort
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(chosenPort))

		if err == nil {
			if err = ln.Close(); err != nil {
				return nil, nil, nil, errors.Wrap(err, "")
			}

			break
		}

		portAttempts++

		if portAttempts > testConf.DBResetConf.MaxPortAttempts {
			return nil, nil, nil, errors.New("can't find empty port")
		}
	}

	conf.ChosenPort = chosenPort
	portFormat := "%v"
	portArgs := []interface{}{chosenPort}
	dynamicArgs := []string{}

	if testConf.DBResetConf.DBStartCommand.PortConfig.DockerPort != "" {
		rand.Seed(time.Now().UnixNano())
		dockerName = strconv.Itoa(rand.Int())
		conf.DockerName = dockerName
		portFormat += ":%v"

		dynamicArgs = append(dynamicArgs, "--name", dockerName, "--hostname", dockerName)

		portArgs = append(
			portArgs,
			testConf.DBResetConf.DBStartCommand.PortConfig.DockerPort,
		)
	}

	portVal := fmt.Sprintf(portFormat, portArgs...)
	dynamicArgs = append(
		dynamicArgs,
		testConf.DBResetConf.DBStartCommand.PortConfig.FlagKey,
		portVal,
	)

	startArgs := []string{}
	hasDynamicArgs := false

	for _, v := range testConf.DBResetConf.DBStartCommand.Args {
		if v[0] == '-' && !hasDynamicArgs {
			startArgs = append(startArgs, v)

			for _, t := range dynamicArgs {
				startArgs = append(startArgs, t)
			}

			hasDynamicArgs = true
		} else {
			startArgs = append(startArgs, v)
		}
	}

	fmt.Printf("command: %s\n", testConf.DBResetConf.DBStartCommand.Command)
	fmt.Printf("args: %v\n", startArgs)

	cmd := exec.Command(
		testConf.DBResetConf.DBStartCommand.Command,
		startArgs...,
	)
	err = cmd.Run()

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "")
	}

	time.Sleep(time.Second * 2)

	baseConn := testConf.DBResetConf.BaseConnection
	baseConn.Port = chosenPort

	dbSettings := []DatabaseSetting{baseConn}
	dbSettings = append(dbSettings, testConf.DBResetConf.DBConnections...)
	db, err = NewDBWithList(dbSettings, Postgres)

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "")
	}

	return db, func(err error) (*DB, error) {
			if err != nil {
				db, err = db.RecoverError()
				return db, errors.Wrap(err, "")
			}

			return db, nil
		}, func() error {
			stopArgs := testConf.DBResetConf.DBRemoveCommand.Args

			if conf.DockerName != "" {
				stopArgs = append(stopArgs, conf.DockerName)
			} else {
				stopArgs = append(stopArgs, strconv.Itoa(conf.ChosenPort))
			}

			cmd := exec.Command(
				testConf.DBResetConf.DBRemoveCommand.Command,
				stopArgs...,
			)
			return cmd.Run()
		}, nil
}

func TestHasDBErrorUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	conf := ServerErrorConfig{}

	if HasDBError(rr, nil, conf) {
		t.Errorf("should not have db error\n")
	}

	if !HasDBError(rr, errDB, conf) {
		t.Errorf("should have err\n")
	}

	// Mocking recovering from db so should
	// return false
	conf.RecoverDB = recoverDB
	conf.DBInterfaceRecover = &testAPI{}
	conf.RetryDB = func(db DBInterface) error {
		return nil
	}

	if HasDBError(rr, errDB, conf) {
		buf := &bytes.Buffer{}
		buf.ReadFrom(rr.Result().Body)
		rr.Result().Body.Close()
		t.Errorf("should not have db error\n")
		t.Errorf("response: %s\n", buf.String())
	}

	// Mocking fail recovery from db so should
	// return true
	conf.RecoverDB = failedRecoverDB

	if !HasDBError(rr, errDB, conf) {
		t.Errorf("should have db error\n")
	}
}

func TestHasNoRowsOrDBErrorUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	conf := ServerErrorConfig{}

	if HasNoRowsOrDBError(rr, nil, conf) {
		t.Errorf("should not have db error\n")
	}

	if !HasNoRowsOrDBError(rr, sql.ErrNoRows, conf) {
		t.Errorf("should have db error\n")
	}
}

// func TestPopulateDatabaseTablesUnitTest(t *testing.T) {
// 	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

// 	if err != nil {
// 		t.Fatalf("fatal err: %s\n", err.Error())
// 	}

// 	rows := sqlmock.NewRows([]string{"tableName"}).
// 		AddRow("phone").
// 		AddRow("phone_status")
// 	mockDB.ExpectQuery("select").WillReturnRows()
// 	mockDB.ExpectBegin()
// }

func TestPopulateDatabaseTablesIntegrationTest(t *testing.T) {
	var err error
	var recoverFn func(err error) (*DB, error)
	var db *DB
	var removeFn func() error

	if db, recoverFn, removeFn, err = dbRecoverSetup(); err != nil {
		t.Fatalf("err: %s\n", errors.Cause(err).Error())
	}

	if err = removeFn(); err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	fooCreate :=
		`
	create table IF NOT EXISTS foo(
		id serial primary key,
		name text not null
	);
	`

	if db, err = recoverFn(errors.New("foo")); err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	// barCreate :=
	// 	`
	// create table IF NOT EXISTS bar(
	// 	id serial primary key,
	// 	name text not null
	// );
	// `

	// bazCreate :=
	// 	`
	// create table IF NOT EXISTS baz(
	// 	id serial primary key,
	// 	name text not null
	// );
	// `

	// databaseTableCreate :=
	// 	`
	// CREATE TABLE IF NOT EXISTS database_table(
	// 	id serial primary key,
	// 	name text not null unique,
	// 	display_name text not null unique,
	// 	column_name text not null
	// );
	// `

	// tableCreate :=
	// 	`
	// create table IF NOT EXISTS foo(
	// 	id serial primary key,
	// 	name text not null
	// );

	// create table IF NOT EXISTS bar(
	// 	id serial primary key,
	// 	name text not null
	// );

	// create table IF NOT EXISTS baz(
	// 	id serial primary key,
	// 	name text not null
	// );

	// CREATE TABLE IF NOT EXISTS database_table(
	// 	id serial primary key,
	// 	name text not null unique,
	// 	display_name text not null unique,
	// 	column_name text not null
	// );
	// `

	if _, err = db.Exec(fooCreate); err != nil {
		removeFn()
		t.Fatalf("err: %s\n", err.Error())
	}

	// if _, err = db.Exec(barCreate); err != nil {
	// 	removeFn()
	// 	t.Fatalf("err: %s\n", err.Error())
	// }

	// if _, err = db.Exec(bazCreate); err != nil {
	// 	removeFn()
	// 	t.Fatalf("err: %s\n", err.Error())
	// }

	// if _, err = db.Exec(databaseTableCreate); err != nil {
	// 	removeFn()
	// 	t.Fatalf("err: %s\n", err.Error())
	// }

	if err = PopulateDatabaseTables(
		db,
		Postgres,
		map[string]string{
			"foo": "name",
		},
		[]string{"baz"},
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		t.Errorf("err: %s\n", err.Error())
	}

	if err = removeFn(); err != nil {
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestHasDBErrorIntegrationTest(t *testing.T) {
	api := testAPI{}
	r := mux.NewRouter()
	r.HandleFunc("/test", api.Index)

	s := httptest.NewServer(r)
	c := s.Client()

	res, err := c.Get(s.URL + "/test")

	if err != nil {
		t.Errorf("%s\n", err.Error())
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("did not return status ok\n")
		t.Errorf("status: %s\n", err.Error())
	}
}

func TestRecoveryErrorIntegrationTest(t *testing.T) {
	var err error
	var recoverFn func(err error) (*DB, error)
	var db *DB
	var wg sync.WaitGroup
	var removeFn func() error

	if db, recoverFn, removeFn, err = dbRecoverSetup(); err != nil {
		t.Fatalf("err: %s\n", errors.Cause(err).Error())
	}

	oneShot := newChannel()
	r := mux.NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		var name string
		fmt.Printf("req from: %s\n", req.RemoteAddr)

		scanner := db.QueryRow(testConf.DBResetConf.ValidateQuery)
		err = scanner.Scan(&name)

		if err != nil {
			fmt.Printf("db is down from req: %s\n", req.RemoteAddr)

			if _, err = recoverFn(err); err == nil {
				fmt.Printf("able to recover err from req: %s\n", req.RemoteAddr)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("could not recover err from req: %s\n", req.RemoteAddr)
			}

		} else {
			fmt.Printf("no db error from reg: %s\n", req.RemoteAddr)
		}
	})

	numOfClients := 50
	s := httptest.NewServer(r)

	// Spin up threads of clients making requests to test api point
	for i := 0; i < numOfClients; i++ {
		wg.Add(1)
		c := s.Client()
		go func() {
			for {
				time.Sleep(time.Millisecond * 700)
				res, err := c.Get(s.URL + "/test")

				if err != nil {
					t.Errorf("%s\n", err.Error())
				}

				if res.StatusCode != http.StatusOK {
					t.Errorf("did not return staus ok\n")
				}

				select {
				case <-oneShot.Get():
					fmt.Printf("stop request from client\n")
					wg.Done()
					return
				default:
					fmt.Printf("default foo\n")
					continue
					// wg.Done()
					// break L
				}
			}
		}()
	}

	// Allow for the clients to make a couple of requests
	time.Sleep(time.Second * 2)

	// stopArgs := testConf.DBResetConf.DBRemoveCommand.Args

	// if conf.DockerName != "" {
	// 	stopArgs = append(stopArgs, conf.DockerName)
	// } else {
	// 	stopArgs = append(stopArgs, strconv.Itoa(conf.ChosenPort))
	// }

	// cmd := exec.Command(
	// 	testConf.DBResetConf.DBRemoveCommand.Command,
	// 	stopArgs...,
	// )
	// err = cmd.Run()

	err = removeFn()

	if err != nil {
		t.Errorf("Could not quit database: %s\n", err.Error())
		t.Errorf(
			"command: %s, args: %v\n",
			testConf.DBResetConf.DBRemoveCommand.Command,
			testConf.DBResetConf.DBRemoveCommand.Args,
		)
		t.Fatalf("err: %s", err.Error())
	}

	oneShot.Stop()
	wg.Wait()
}

// func TestHasDBErrorIntegrationTest(t *testing.T) {
// 	var err error
// 	//var recoverFn RecoverDB
// 	var rows *sql.Rows
// 	var db *DB
// 	var removeFn func() error

// 	if db, _, removeFn, err = dbRecoverSetup(); err != nil {
// 		t.Fatalf("err: %s\n", errors.Cause(err).Error())
// 	}

// 	rr := httptest.NewRecorder()
// 	conf := ServerErrorConfig{
// 		RecoverConfig: RecoverConfig{
// 			RecoverDB: func(err error) error {
// 				if err != nil {
// 					db, err = db.RecoverError()
// 					return err
// 				}

// 				return nil
// 			},
// 		},
// 	}
// 	err = removeFn()

// 	if err != nil {
// 		t.Fatalf("err: %s\n", err.Error())
// 	}

// 	fmt.Printf("config: %v", db.currentConfig)

// 	// defer func() {
// 	// 	cmd = exec.Command(
// 	// 		testConf.DBResetConf.DBRemoveCommand.Command,
// 	// 		testConf.DBResetConf.DBRemoveCommand.Args...,
// 	// 	)
// 	// 	cmd.Run()
// 	// }()

// 	validateQuery := func() error {
// 		rows, err = db.Query(testConf.DBResetConf.ValidateQuery)
// 		return err
// 	}

// 	err = validateQuery()

// 	if err == nil {
// 		t.Errorf("should have error\n")
// 	}

// 	conf.RetryDB = validateQuery

// 	// if HasDBError(rr, err, conf) {
// 	// 	t.Fatalf("could not recover")
// 	// }

// 	if rows == nil {
// 		t.Errorf("rows is nil\n")
// 	}

// 	results := make([]interface{}, 0)

// 	for rows.Next() {
// 		var result interface{}
// 		err = rows.Scan(&result)

// 		if err != nil {
// 			t.Fatalf("err: %s\n", err.Error())
// 		}

// 		results = append(results, result)
// 	}
// }

// func TestRecoverAndRetryIntegrationTest(t *testing.T) {
// 	var err error
// 	var db *DB
// 	var removeFn func() error

// 	if db, _, removeFn, err = dbRecoverSetup(); err != nil {
// 		t.Fatalf("err: %s\n", errors.Cause(err).Error())
// 	}

// 	if err = removeFn(); err != nil {
// 		t.Fatalf("err: %s\n", err.Error())
// 	}

// 	var dbI DBInterface

// 	dbI = db

// 	validateQuery := func(innerDBI DBInterface) error {
// 		_, err = innerDBI.Query(testConf.DBResetConf.ValidateQuery)
// 		return err
// 	}

// 	if err = validateQuery(dbI); err == nil {
// 		t.Fatalf("should have error\n")
// 	}

// 	recoverDB := func(err error) (*DB, error) {
// 		if err != nil {
// 			dbMutex.Lock()
// 			defer dbMutex.Unlock()
// 			return db.RecoverError()
// 		}

// 		return db, nil
// 	}

// 	retryDB := func(foo DBInterface) error {
// 		return validateQuery(foo)
// 	}

// 	conf := RecoverConfig{
// 		RecoverDB: recoverDB,
// 		RetryDB:   retryDB,
// 	}

// 	if dbI, err = RecoverAndRetry(err, conf); err != nil {
// 		t.Fatalf("err: %s\n", err.Error())
// 	}

// 	if err = validateQuery(dbI); err != nil {
// 		t.Fatalf("err: %s\n", err.Error())
// 	}
// }
