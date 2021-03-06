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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var _ ResetDB = (*testAPI)(nil)

var (
	dbMutex sync.Mutex

	errDB     = errors.New("db error")
	recoverDB = func(err error) (*sqlx.DB, error) {
		return &sqlx.DB{}, nil
	}
	failedRecoverDB = func(err error) (*sqlx.DB, error) {
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

func (f *testAPI) SetDB(db DBInterface) {
	f.DB = db
}

func (f *testAPI) Index(w http.ResponseWriter, r *http.Request) {
	var db *sqlx.DB
	var err error
	var removeFn func() error

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
		_, err = innerDB.Queryx(testConf.DBResetConf.ValidateQuery)
		return err
	}

	if err = retry(f.DB); err == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("couldn't retry"))
		return
	}

	conf := ServerErrorConfig{
		RecoverConfig: RecoverConfig{
			RecoverDB: func(err error) (*sqlx.DB, error) {
				if err != nil {
					dbMutex.Lock()
					defer dbMutex.Unlock()
					db, err = NewDBWithList(testConf.DBResetConf.DBConnections, Postgres)
					return db, err
				}

				return db, nil
			},
		},
	}

	if HasDBError(w, r, err, retry, conf) {
		return
	}

	if _, err = f.DB.Queryx(testConf.DBResetConf.ValidateQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("didn't recovery right"))
	}
}

// dbRecoverSetup sets up and returns new db with implemented RecoverDB function
func dbRecoverSetup() (*sqlx.DB, func(err error) (*sqlx.DB, error), func() error, error) {
	var db *sqlx.DB
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

	for _, v := range testConf.DBResetConf.DBStartCommand.CmdCommand.Args {
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

	fmt.Printf("command: %s\n", testConf.DBResetConf.DBStartCommand.CmdCommand.Command)
	fmt.Printf("args: %v\n", startArgs)

	cmd := exec.Command(
		testConf.DBResetConf.DBStartCommand.CmdCommand.Command,
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

	return db, func(err error) (*sqlx.DB, error) {
			if err != nil {
				db, err = NewDBWithList(dbSettings, Postgres)
				return db, errors.Wrap(err, "")
			}

			return db, nil
		}, func() error {
			stopArgs := testConf.DBResetConf.DBRemoveCommand.CmdCommand.Args

			if conf.DockerName != "" {
				stopArgs = append(stopArgs, conf.DockerName)
			} else {
				stopArgs = append(stopArgs, strconv.Itoa(conf.ChosenPort))
			}

			cmd := exec.Command(
				testConf.DBResetConf.DBRemoveCommand.CmdCommand.Command,
				stopArgs...,
			)
			return cmd.Run()
		}, nil
}

func TestIsDBError(t *testing.T) {
	conf := ServerErrorConfig{}
	req := httptest.NewRequest(http.MethodGet, "/url", nil)

	if IsDBError(req, nil, nil, conf) {
		t.Errorf("should not have db error\n")
	}
}

func TestHasDBErrorUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/url", nil)
	conf := ServerErrorConfig{}

	if HasDBError(rr, r, nil, nil, conf) {
		t.Errorf("should not have db error\n")
	}

	if !HasDBError(rr, r, errDB, nil, conf) {
		t.Errorf("should have err\n")
	}

	// Mocking recovering from db so should
	// return false
	conf.RecoverDB = recoverDB
	retryFn := func(db DBInterface) error {
		return nil
	}

	if HasDBError(rr, r, errDB, retryFn, conf) {
		buf := &bytes.Buffer{}
		buf.ReadFrom(rr.Result().Body)
		rr.Result().Body.Close()
		t.Errorf("should not have db error\n")
		t.Errorf("response: %s\n", buf.String())
	}

	// Mocking fail recovery from db so should
	// return true
	conf.RecoverDB = failedRecoverDB

	if !HasDBError(rr, r, errDB, retryFn, conf) {
		t.Errorf("should have db error\n")
	}

	retryFn = nil
	conf.RecoverDB = recoverDB
	// conf.RetryQuerier = func(db Querier) error {
	// 	return nil
	// }

	// if HasDBError(rr, errDB, conf) {
	// 	buf := &bytes.Buffer{}
	// 	buf.ReadFrom(rr.Result().Body)
	// 	rr.Result().Body.Close()
	// 	t.Errorf("should not have db error\n")
	// 	t.Errorf("response: %s\n", buf.String())
	// }
}

func TestServerErrorRecoverUnitTest(t *testing.T) {
	var valid bool
	var err error

	r := httptest.NewRequest(http.MethodGet, "/url", nil)
	cfg := ServerErrorConfig{}

	valid, err = ServerErrorRecover(r, nil, nil, nil, cfg)

	if !valid {
		t.Errorf("should be valid")
	}

	if err != nil {
		t.Errorf("should not have err; got %s\n", err.Error())
	}

	valid, err = ServerErrorRecover(
		r,
		sql.ErrNoRows,
		[]error{sql.ErrNoRows},
		nil,
		cfg,
	)

	if !valid {
		t.Errorf("should be valid")
	}

	if err == nil {
		t.Errorf("should have sql.ErrNoRows error")
	} else {
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("should have sql.ErrNoRows error; got %s\n", err.Error())
		}
	}

	cfg.RecoverDB = func(err error) (*sqlx.DB, error) {
		return &sqlx.DB{}, nil
	}

	testErr := fmt.Errorf("test error")

	valid, err = ServerErrorRecover(
		r,
		testErr,
		[]error{sql.ErrNoRows},
		func(db DBInterface) error {
			return sql.ErrNoRows
		},
		cfg,
	)

	if !valid {
		t.Errorf("should be valid")
	}

	if err == nil {
		t.Errorf("should have sql.ErrNoRows error")
	} else {
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("should have sql.ErrNoRows error; got %s\n", err.Error())
		}
	}

	cfg.Logger = func(r *http.Request, cfg LogConfig) {
		if !errors.Is(cfg.CauseErr, testErr) {
			t.Errorf("should have test err: got %s\n", cfg.CauseErr.Error())
		}
	}

	valid, err = ServerErrorRecover(
		r,
		testErr,
		[]error{sql.ErrNoRows},
		func(db DBInterface) error {
			return testErr
		},
		cfg,
	)

	if valid {
		t.Errorf("should not be valid")
	}

	if err == nil {
		t.Errorf("should have testErr")
	} else {
		if !errors.Is(err, testErr) {
			t.Errorf("should have testErr; got %s\n", err.Error())
		}
	}

	valid, err = ServerErrorRecover(
		r,
		testErr,
		[]error{sql.ErrNoRows},
		nil,
		cfg,
	)

	if valid {
		t.Errorf("should not be valid")
	}

	if err == nil {
		t.Errorf("should have testErr")
	} else {
		if !errors.Is(err, testErr) {
			t.Errorf("should have testErr; got %s\n", err.Error())
		}
	}

	cfg.RecoverDB = func(err error) (*sqlx.DB, error) {
		return nil, testErr
	}

	valid, err = ServerErrorRecover(
		r,
		testErr,
		[]error{sql.ErrNoRows},
		nil,
		cfg,
	)

	if valid {
		t.Errorf("should not be valid")
	}

	if err == nil {
		t.Errorf("should have testErr")
	} else {
		if !errors.Is(err, testErr) {
			t.Errorf("should have testErr; got %s\n", err.Error())
		}
	}

	cfg.RecoverDB = nil

	valid, err = ServerErrorRecover(
		r,
		testErr,
		[]error{sql.ErrNoRows},
		nil,
		cfg,
	)

	if valid {
		t.Errorf("should not be valid")
	}

	if err == nil {
		t.Errorf("should have testErr")
	} else {
		if !errors.Is(err, testErr) {
			t.Errorf("should have testErr; got %s\n", err.Error())
		}
	}
}

func TestHasNoRowsOrDBErrorUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/url", nil)
	conf := ServerErrorConfig{}

	if HasNoRowsOrDBError(rr, r, nil, nil, HTTPResponseConfig{}, conf) {
		t.Errorf("should not have db error\n")
	}

	if !HasNoRowsOrDBError(rr, r, sql.ErrNoRows, nil, HTTPResponseConfig{}, conf) {
		t.Errorf("should have db error\n")
	}
}

func TestPopulateDatabaseTablesIntegrationTest(t *testing.T) {
	var err error
	var recoverFn func(err error) (*sqlx.DB, error)
	var db *sqlx.DB
	var removeFn func() error

	if db, recoverFn, removeFn, err = dbRecoverSetup(); err != nil {
		t.Fatalf("err: %s\n", errors.Cause(err).Error())
	}

	if err = removeFn(); err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	if db, err = recoverFn(errors.New("foo")); err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	tableCreate :=
		`
	create table IF NOT EXISTS foo(
		id serial primary key,
		name text not null
	);

	create table IF NOT EXISTS bar(
		id serial primary key,
		name text not null
	);

	create table IF NOT EXISTS baz(
		id serial primary key,
		name text not null
	);

	create table IF NOT EXISTS database_table(
		id serial primary key,
		name text not null unique,
		display_name text not null unique,
		column_name text not null
	);
	`

	if _, err = db.Exec(tableCreate); err != nil {
		removeFn()
		t.Fatalf("err: %s\n", err.Error())
	}

	if err = PopulateDatabaseTables(
		db,
		Postgres,
		nil,
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != "can not have empty inclusion map" {
			t.Errorf("error should be '%s'", "can not have empty inclusion map")
		}
	}

	if err = PopulateDatabaseTables(
		db,
		Postgres,
		map[string]string{
			"foo": "name",
			"baz": "name",
		},
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if strings.Contains(err.Error(), "baz") {
			t.Errorf("error should contain bar table")
		}
	}

	if err = PopulateDatabaseTables(
		db,
		Postgres,
		map[string]string{
			"foo": "name",
			"baz": "nam",
		},
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err.Error() != "table baz does not contain column 'nam'" {
			t.Errorf("error should contain baz column being wrong\n")
			t.Errorf("err: %s\n", err.Error())
		}
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
	var recoverFn func(err error) (*sqlx.DB, error)
	var db *sqlx.DB
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
			testConf.DBResetConf.DBRemoveCommand.CmdCommand.Command,
			testConf.DBResetConf.DBRemoveCommand.CmdCommand.Args,
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
// 	var db *sqlx.DB
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
// 	var db *sqlx.DB
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

// 	recoverDB := func(err error) (*sqlx.DB, error) {
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
