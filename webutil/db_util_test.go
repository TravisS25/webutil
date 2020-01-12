package webutil

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	dbMutex sync.Mutex
	db      *DB

	errDB     = errors.New("db error")
	recoverDB = func(err error) error {
		return nil
	}
	failedRecoverDB = func(err error) error {
		return err
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

func RecoverFromError(err error) error {
	if err != nil {
		dbMutex.Lock()
		defer dbMutex.Unlock()
		db, err = db.RecoverError(err)
		return err
	}

	return nil
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
	conf.RetryDB = func() error {
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

func TestRecoveryErrorIntegrationTest(t *testing.T) {
	var err error

	var wg sync.WaitGroup

	oneShot := newChannel()
	r := mux.NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		var name string
		fmt.Printf("req from: %s\n", req.RemoteAddr)

		scanner := db.QueryRow(testConf.DBResetConfiguration.ValidateQuery)
		err = scanner.Scan(&name)

		if err != nil {
			fmt.Printf("db is down from req: %s\n", req.RemoteAddr)

			if err = RecoverFromError(err); err == nil {
				fmt.Printf("able to recover err from req: %s\n", req.RemoteAddr)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("could not recover err from req: %s\n", req.RemoteAddr)
			}

		} else {
			fmt.Printf("No db error from reg: %s\n", req.RemoteAddr)
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
					t.Errorf("Did not return staus ok\n")
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

	cmd := exec.Command(
		testConf.DBResetConfiguration.DbStopCommand.Command,
		testConf.DBResetConfiguration.DbStopCommand.Args...,
	)
	err = cmd.Run()

	if err != nil {
		t.Errorf("Could not quit database: %s\n", err.Error())
		t.Errorf(
			"command: %s, args: %v\n",
			testConf.DBResetConfiguration.DbStopCommand.Command,
			testConf.DBResetConfiguration.DbStopCommand.Args,
		)
		t.Fatalf("err: %s", err.Error())
	}

	oneShot.Stop()
	wg.Wait()

	cmd = exec.Command(
		testConf.DBResetConfiguration.DbStartCommand.Command,
		testConf.DBResetConfiguration.DbStartCommand.Args...,
	)
	err = cmd.Run()

	if err != nil {
		t.Fatalf("Could not bring database back up: %s\n", err.Error())
	}
}

func TestRecoverDBIntegrationTest(t *testing.T) {
	var err error
	var rows *sql.Rows

	rr := httptest.NewRecorder()
	conf := ServerErrorConfig{
		RecoverDB: RecoverFromError,
	}
	cmd := exec.Command(
		testConf.DBResetConfiguration.DbStopCommand.Command,
		testConf.DBResetConfiguration.DbStopCommand.Args...,
	)
	err = cmd.Run()

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	defer func() {
		cmd = exec.Command(
			testConf.DBResetConfiguration.DbStartCommand.Command,
			testConf.DBResetConfiguration.DbStartCommand.Args...,
		)
		cmd.Run()
	}()

	validateQuery := func() error {
		rows, err = db.Query(testConf.DBResetConfiguration.ValidateQuery)
		return err
	}

	err = validateQuery()

	if err == nil {
		t.Errorf("should have error\n")
	}

	conf.RetryDB = validateQuery

	if HasDBError(rr, err, conf) {
		t.Fatalf("could not recover")
	}

	if rows == nil {
		t.Errorf("rows is nil\n")
	}

	results := make([]interface{}, 0)

	for rows.Next() {
		var result interface{}
		err = rows.Scan(&result)

		if err != nil {
			t.Fatalf("err: %s\n", err.Error())
		}

		results = append(results, result)
	}
}
