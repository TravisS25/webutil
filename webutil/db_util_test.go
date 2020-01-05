package webutil

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	dbMutex sync.Mutex

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

type dbCommand struct {
	Command string   `yaml:"command" mapstructure:"command"`
	Args    []string `yaml:"args" mapstructure:"args"`
}

type dbConfiguration struct {
	DbConnections  []DatabaseSetting `yaml:"db_connections" mapstructure:"db_connections"`
	DbStopCommand  dbCommand         `yaml:"db_stop_command" mapstructure:"db_stop_command"`
	DbStartCommand dbCommand         `yaml:"db_start_command" mapstructure:"db_start_command"`
	ValidateQuery  string            `yaml:"validate_query" mapstructure:"validate_query"`
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

	if HasDBError(rr, errDB, conf) {
		t.Errorf("should not have db error\n")
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
	var config dbConfiguration

	filePath := os.Getenv(WebUtilConfig)

	if filePath == "" {
		t.Fatalf("env var not set\n")
	}

	v := viper.New()
	v.SetConfigFile(filePath)

	if err = v.ReadInConfig(); err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}
	if err = v.Unmarshal(&config); err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	db, err := NewDBWithList(config.DbConnections, Postgres)

	if err != nil {
		t.Fatalf(err.Error())
	}

	var wg sync.WaitGroup

	RecoverFromError := func(err error) error {
		dbMutex.Lock()
		defer dbMutex.Unlock()
		db, err = db.RecoverError(err)
		return err
	}

	oneShot := newChannel()
	r := mux.NewRouter()
	r.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		var name string
		fmt.Printf("req from: %s\n", req.RemoteAddr)

		scanner := db.QueryRow(config.ValidateQuery)
		err = scanner.Scan(&name)

		if err != nil {
			fmt.Printf("db is down from req: %s\n", req.RemoteAddr)

			if err = RecoverFromError(err); err == nil {
				fmt.Printf("able to recover err from req: %s\n", req.RemoteAddr)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Printf("could not recover err from req: %s\n", req.RemoteAddr)
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

	cmd := exec.Command(config.DbStopCommand.Command, config.DbStopCommand.Args...)
	err = cmd.Start()

	if err != nil {
		t.Fatalf("Could not quit database\n")
	}

	// Allow for at least one client to connect to
	// db while down to try to recover and allow other
	// clients to connect to new db connection
	time.Sleep(time.Second * 5)

	oneShot.Stop()
	wg.Wait()

	cmd = exec.Command(config.DbStartCommand.Command, config.DbStartCommand.Args...)
	err = cmd.Start()

	if err != nil {
		t.Fatalf("Could not bring database back up\n")
	}
}

func TestBar(t *testing.T) {
	_, err := NewDB(DatabaseSetting{
		DBName:   "test1",
		User:     "test",
		Password: "password",
		Host:     "localhost",
		Port:     "26257",
		SSLMode:  SSLRequire,
	}, Postgres)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}
}
