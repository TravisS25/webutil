package webutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gomock "github.com/golang/mock/gomock"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"

	"github.com/gorilla/securecookie"
)

const (
	cookieName   = "user"
	statusErrTxt = "Status should be %d; got %d"
)

var (
	keyPairs = [][]byte{
		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
	}
)

func getMockSession(cookieName string) *sessions.Session {
	mockSession := &MockSessionStore{}
	return sessions.NewSession(mockSession, cookieName)
}

func TestDecodeCookieUnitTest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/url", nil)
	session := getMockSession(cookieName)
	encoded, err := securecookie.EncodeMulti(
		session.Name(),
		session.ID,
		securecookie.CodecsFromPairs(keyPairs...)...,
	)
	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	req.AddCookie(sessions.NewCookie(cookieName, encoded, &sessions.Options{}))
	_, err = DecodeCookie(req, cookieName, keyPairs[0], keyPairs[1])

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}
}

func TestSetSecureCookieUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	session := getMockSession(cookieName)

	if _, err := SetSecureCookie(rr, session, keyPairs...); err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}
}

func TestGetUserUnitTest(t *testing.T) {
	var user []byte

	userBytes := []byte(`{ id: 1, email: "test@email.com" }`)
	req := httptest.NewRequest(http.MethodGet, "/url", nil)

	if user = GetUser(req); user != nil {
		t.Fatalf("user should be nil\n")
	}

	ctx := context.WithValue(req.Context(), UserCtxKey, userBytes)
	req = req.WithContext(ctx)

	if user = GetUser(req); user == nil {
		t.Fatalf("should return user\n")
	}
}

func TestGetMiddlewareUserUnitTest(t *testing.T) {
	user := &MiddlewareUser{
		ID:    "1",
		Email: "email@email.com",
	}
	req := httptest.NewRequest(http.MethodGet, "/url", nil)

	if u := GetMiddlewareUser(req); u != nil {
		t.Fatalf("user should be nil\n")
	}

	ctx := context.WithValue(req.Context(), MiddlewareUserCtxKey, user)
	req = req.WithContext(ctx)

	if user = GetMiddlewareUser(req); user == nil {
		t.Fatalf("should return user\n")
	}
}

func TestSendPayloadUnitTest(t *testing.T) {
	var err error

	rr := httptest.NewRecorder()
	conf := HTTPResponseConfig{}
	payload := struct {
		ID string
	}{
		"foo",
	}

	if err = SendPayload(rr, payload, conf); err != nil {
		t.Fatalf("should not have payload error\n")
	}

	if err = SendPayload(rr, make(chan int), conf); err == nil {
		t.Fatalf("should have error\n")
	}
}

func TestGetMapSliceRowItems(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf(err.Error())
	}

	scanErr := fmt.Errorf("scan error")
	dbx := sqlx.NewDb(db, Postgres)
	status := 406

	serverErrCfg := ServerErrorConfig{
		RecoverConfig: RecoverConfig{
			RecoverDB: func(err error) (*sqlx.DB, error) {
				return dbx, nil
			},
		},
	}

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"value", "text"}).
			AddRow(nil, "bar").
			AddRow("hey", "there").
			RowError(1, scanErr),
	)

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(1),
	)

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"value", "text"}).
			AddRow(nil, "bar").
			AddRow("hey", "there").
			RowError(1, scanErr),
	)

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(1),
	)

	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/url", nil)

	if _, _, err = GetMapSliceRowItems(
		rr,
		r,
		dbx,
		status,
		func(db DBInterface) (*sqlx.Rows, int, error) {
			return GetQueriedAndCountResults(
				"select",
				"select",
				nil,
				DbFields{},
				r,
				dbx,
				ParamConfig{},
				QueryConfig{},
			)
		},
		serverErrCfg,
	); err == nil {
		t.Errorf("should have error")
	} else if !errors.Is(err, scanErr) {
		t.Errorf("should hvae scan err; got %s\n", err.Error())
	}

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"value", "text"}).
			AddRow(nil, "bar"),
	)

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).
			AddRow(1),
	)

	rr = httptest.NewRecorder()

	if _, _, err = GetMapSliceRowItems(
		rr,
		r,
		dbx,
		status,
		func(db DBInterface) (*sqlx.Rows, int, error) {
			return GetQueriedAndCountResults(
				"select",
				"select",
				nil,
				DbFields{},
				r,
				dbx,
				ParamConfig{},
				QueryConfig{},
			)
		},
		serverErrCfg,
	); err != nil {
		t.Errorf("should not have error;got %v\n", err)
	}
}

func TestGetMapSliceRowItemsWithRow(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf(err.Error())
	}

	scanErr := fmt.Errorf("scan error")
	dbx := sqlx.NewDb(db, Postgres)
	status := 406

	serverErrCfg := ServerErrorConfig{
		RecoverConfig: RecoverConfig{
			RecoverDB: func(err error) (*sqlx.DB, error) {
				return dbx, nil
			},
		},
	}

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"value", "text"}).
			AddRow(nil, "bar").
			AddRow("hey", "there").
			RowError(1, scanErr),
	)

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"id", "count"}).
			AddRow(1, 1),
	)

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"value", "text"}).
			AddRow(nil, "bar").
			AddRow("hey", "there").
			RowError(1, scanErr),
	)

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"id", "count"}).
			AddRow(1, 1),
	)

	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/url", nil)

	if _, _, err = GetMapSliceRowItemsWithRow(
		rr,
		r,
		dbx,
		status,
		func(db DBInterface) (*sqlx.Rows, *sqlx.Row, error) {
			return GetQueriedAndCountRowResults(
				"select",
				"select",
				nil,
				DbFields{},
				r,
				dbx,
				ParamConfig{},
				QueryConfig{},
			)
		},
		serverErrCfg,
	); err == nil {
		t.Errorf("should have error")
	} else if !errors.Is(err, scanErr) {
		t.Errorf("should hvae scan err; got %s\n", err.Error())
	}

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"value", "text"}).
			AddRow(nil, "bar"),
	)

	mock.ExpectQuery("select").WillReturnRows(
		sqlmock.NewRows([]string{"id", "count"}).
			AddRow(1, 1),
	)

	rr = httptest.NewRecorder()

	if _, _, err = GetMapSliceRowItemsWithRow(
		rr,
		r,
		dbx,
		status,
		func(db DBInterface) (*sqlx.Rows, *sqlx.Row, error) {
			return GetQueriedAndCountRowResults(
				"select",
				"select",
				nil,
				DbFields{},
				r,
				dbx,
				ParamConfig{},
				QueryConfig{},
			)
		},
		serverErrCfg,
	); err != nil {
		t.Errorf("should not have error;got %v\n", err)
	}
}
