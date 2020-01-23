package webutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	sessions "github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	testifymock "github.com/stretchr/testify/mock"
)

const (
	decodeErr         = "decodeErr"
	internalErr       = "internalErr"
	noRowsErr         = "noRowsErr"
	generalErr        = "generalErr"
	invalidJSONErr    = "invalidJSONErr"
	recoverUserErr    = "recoverUserErr"
	recoverSessionErr = "recoverSessionErr"
	recoverErr        = "recoverErr"

	queryUser    = "queryUser"
	querySession = "querySession"

	errType1 = "errType1"
	errType2 = "errType2"
	errType3 = "errType3"

	defaultURL = "/url"
)

var (
	// This should be used for read only
	mUser = MiddlewareUser{
		ID:    "1",
		Email: "someemail@email.com",
	}
)

type middlewareState struct {
	useSessionCounter bool

	pRegex string

	groupError string
	//routeError string

	sessionCounter int
	userCounter    int
	groupCounter   int
	//routeCounter   int

	hasRecoverError bool
	errType         string
	errCounter      int
}

func (a *middlewareState) queryForUser(r *http.Request, db Querier) ([]byte, error) {
	if r.Header.Get(queryUser) == decodeErr {
		cookieError := &Error{}
		cookieError.On("IsDecode").Return(true)
		return nil, cookieError
	}

	if r.Header.Get(queryUser) == internalErr {
		cookieError := &Error{}
		cookieError.On("IsDecode").Return(false)
		return nil, cookieError
	}

	if r.Header.Get(queryUser) == noRowsErr {
		return nil, sql.ErrNoRows
	}

	if r.Header.Get(queryUser) == generalErr {
		return nil, errors.New(generalErr)
	}

	if r.Header.Get(queryUser) == invalidJSONErr {
		sMap := []string{"foobar"}
		return json.Marshal(sMap)
	}

	if r.Header.Get(queryUser) == recoverUserErr {
		a.userCounter++

		fmt.Printf("counter: %d\n", a.userCounter)

		if a.userCounter == 1 {
			return nil, errors.New("errors")
		}

		return nil, sql.ErrNoRows
	}

	u := mUser

	if r.Header.Get(querySession) == noRowsErr {
		u.ID = "0"
	}

	if r.Header.Get(querySession) == generalErr {
		u.ID = "-1"
	}

	return json.Marshal(&u)
}

func (a *middlewareState) queryForSession(db Querier, userID string) (sessionID string, err error) {
	if a.useSessionCounter {
		a.sessionCounter++
		fmt.Printf("blah counter: %d\n", a.sessionCounter)

		if a.sessionCounter == 1 {
			return "", errors.New("errors")
		}

		return "", sql.ErrNoRows
	}

	return "sessionID", nil
}

func (a *middlewareState) queryForGroups(r *http.Request, db Querier) ([]byte, error) {
	if a.groupError == noRowsErr {
		return nil, sql.ErrNoRows
	}

	if a.groupError == recoverErr {
		a.groupCounter++

		if a.groupCounter == 1 {
			return nil, errors.New("errors")
		}

		return nil, sql.ErrNoRows
	}

	if a.groupError == generalErr {
		return nil, errors.New("errors")
	}

	if a.groupError == invalidJSONErr {
		return json.Marshal([]string{"invalid"})
	}

	return json.Marshal(&map[string]bool{
		"Admin": true,
	})
}

func (a *middlewareState) queryForRoutes(r *http.Request, db Querier) ([]byte, error) {
	if a.errType == errType1 {
		return nil, sql.ErrNoRows
	}
	if a.errType == errType2 {
		if a.errCounter == 0 {
			a.errCounter++
			return nil, errors.New("errors")
		}

		return json.Marshal(map[string]bool{
			"Admin": true,
		})
	}
	if a.errType == errType3 {
		if a.errCounter == 0 {
			a.errCounter++
			return nil, errors.New("errors")
		}

		return nil, sql.ErrNoRows
	}
	if a.errType == invalidJSONErr {
		return json.Marshal([]string{"foo"})
	}

	return json.Marshal(map[string]bool{
		defaultURL: true,
	})
}

func (a *middlewareState) pathRegex(r *http.Request) (string, error) {
	if a.pRegex == generalErr {
		return "", errors.New("errors")
	}

	return a.pRegex, nil
}

func (a *middlewareState) reset() {
	a.sessionCounter = 0
	a.userCounter = 0
	a.groupCounter = 0

	a.useSessionCounter = false
	a.groupError = ""
	a.pRegex = ""

	a.errType = ""
	a.errCounter = 0
	a.hasRecoverError = false
}

func TestAuthHandlerUnitTest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/url", nil)

	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	config := AuthHandlerConfig{
		SessionConfig: SessionConfig{
			SessionName: "user",
			Keys: SessionKeys{
				UserKey: "user",
			},
		},
	}

	mockStore := &Store{}
	session := sessions.NewSession(mockStore, config.SessionConfig.SessionName)
	session.IsNew = true

	middlewareState := &middlewareState{}
	authHandler := NewAuthHandler(db, middlewareState.queryForUser, config)

	mockHandler := &Handler{}
	mockHandler.On("ServeHTTP", testifymock.Anything, testifymock.Anything)
	h := authHandler.MiddlewareFunc(mockHandler)

	// Testing default settings without cache
	rr := httptest.NewRecorder()

	defer mockHandler.AssertExpectations(t)
	mockHandler.On("ServeHTTP", rr, testifymock.Anything)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing query for user returns cookie decode error
	rr = httptest.NewRecorder()
	req.Header.Set(queryUser, decodeErr)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf(statusErrTxt, http.StatusBadRequest, rr.Code)
	}

	// Testing query for user returns cookie internal error
	rr = httptest.NewRecorder()
	req.Header.Set(queryUser, internalErr)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing query for user returns sql.ErrNoRows
	rr = httptest.NewRecorder()
	req.Header.Set(queryUser, noRowsErr)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing when database recovers from error
	rr = httptest.NewRecorder()
	req.Header.Set(queryUser, recoverUserErr)

	config.RecoverDB = func(err error) (*sqlx.DB, error) {
		return &sqlx.DB{}, nil
	}
	authHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing when RecoverDB is not set
	config.RecoverDB = nil
	authHandler.config = config
	rr = httptest.NewRecorder()
	req.Header.Set(queryUser, generalErr)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing invalid json unmarshal
	rr = httptest.NewRecorder()
	req.Header.Set(queryUser, invalidJSONErr)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing with cache set but session returns error
	// and resorts to db
	rr = httptest.NewRecorder()
	req.Header.Set(queryUser, "")
	mockSessionStore1 := &MockSessionStore{}
	defer mockSessionStore1.AssertExpectations(t)
	mockSessionStore1.On("Get", testifymock.Anything, testifymock.Anything).
		Return(nil, errors.New("errors"))
	config.SessionStore = mockSessionStore1
	authHandler.config = config
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing new session but no user cookie in request
	rr = httptest.NewRecorder()
	mockSessionStore2 := &MockSessionStore{}
	defer mockSessionStore2.AssertExpectations(t)
	mockSessionStore2.On("Get", testifymock.Anything, testifymock.Anything).
		Return(session, nil)
	config.SessionStore = mockSessionStore2
	authHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing new session with user cookie but get sql.ErrNoRows
	// when using QueryForSession
	rr = httptest.NewRecorder()
	req.AddCookie(sessions.NewCookie(
		config.SessionConfig.SessionName,
		"session",
		&sessions.Options{},
	))
	mockSessionStore3 := &MockSessionStore{}
	defer mockSessionStore3.AssertExpectations(t)
	mockSessionStore3.On("Get", testifymock.Anything, testifymock.Anything).
		Return(session, nil)
	mockSessionStore3.On("Ping").Return(nil)
	config.SessionStore = mockSessionStore3
	config.QueryForSession = func(db Querier, id string) (string, error) {
		return "", sql.ErrNoRows
	}
	authHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing QueryForSession() will return valid but will fail
	// on getting new session from SessionStore but will
	// continue to next handler
	rr = httptest.NewRecorder()
	middlewareState.reset()
	req.AddCookie(sessions.NewCookie(
		config.SessionConfig.SessionName,
		"session",
		&sessions.Options{},
	))
	mockSessionStore5 := &MockSessionStore{}
	mockSessionStore5.On("Get", testifymock.Anything, testifymock.Anything).
		Return(session, nil)
	mockSessionStore5.On("Ping").Return(nil)
	mockSessionStore5.On("New", testifymock.Anything, testifymock.Anything).
		Return(nil, errors.New("errors"))
	config.SessionStore = mockSessionStore5
	config.RecoverDB = func(err error) (*sqlx.DB, error) {
		return &sqlx.DB{}, nil
	}
	config.QueryForSession = middlewareState.queryForSession
	authHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing QueryForSession() will return valid but will fail
	// on getting new session from SessionStore but will
	// continue to next handler
	rr = httptest.NewRecorder()
	middlewareState.reset()
	req.AddCookie(sessions.NewCookie(
		config.SessionConfig.SessionName,
		"session",
		&sessions.Options{},
	))
	mockSessionStore6 := &MockSessionStore{}
	defer mockSessionStore6.AssertExpectations(t)

	mockStore = &Store{}
	mockStore.On("Save", testifymock.Anything, testifymock.Anything, testifymock.Anything).
		Return(nil)
	session = sessions.NewSession(mockStore, config.SessionConfig.SessionName)
	session.IsNew = true

	mockSessionStore6.On("Get", testifymock.Anything, testifymock.Anything).
		Return(session, nil)
	mockSessionStore6.On("Ping").Return(nil)
	mockSessionStore6.On("New", testifymock.Anything, testifymock.Anything).
		Return(session, nil)

	config.SessionStore = mockSessionStore6
	config.RecoverDB = func(err error) (*sqlx.DB, error) {
		return &sqlx.DB{}, nil
	}
	config.QueryForSession = middlewareState.queryForSession
	authHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing when a session is not new but can't get user
	// value from session so get it from db
	rr = httptest.NewRecorder()

	session = sessions.NewSession(mockStore, config.SessionConfig.SessionName)
	mockSessionStore7 := &MockSessionStore{}
	defer mockSessionStore7.AssertExpectations(t)

	mockStore = &Store{}
	session = sessions.NewSession(mockStore, config.SessionConfig.SessionName)

	mockSessionStore7.On("Get", testifymock.Anything, testifymock.Anything).
		Return(session, nil)
	config.SessionStore = mockSessionStore7
	authHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing when a session is not new and get value from session
	rr = httptest.NewRecorder()

	session = sessions.NewSession(mockStore, config.SessionConfig.SessionName)
	mockSessionStore9 := &MockSessionStore{}
	defer mockSessionStore9.AssertExpectations(t)

	mUser := MiddlewareUser{
		ID:    "1",
		Email: "user@email.com",
	}

	mUserBytes, err := json.Marshal(&mUser)

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	mockStore = &Store{}
	session = sessions.NewSession(mockStore, config.SessionConfig.SessionName)
	session.Values[config.SessionConfig.Keys.UserKey] = mUserBytes

	mockSessionStore9.On("Get", testifymock.Anything, testifymock.Anything).
		Return(session, nil)
	config.SessionStore = mockSessionStore9
	authHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}
}

func TestGroupHandlerUnitTest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/url", nil)
	mUser := MiddlewareUser{
		ID:    "1",
		Email: "email@email.com",
	}
	ctx := context.WithValue(req.Context(), MiddlewareUserCtxKey, mUser)
	req = req.WithContext(ctx)

	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	state := &middlewareState{}
	config := ServerErrorCacheConfig{
		ServerErrorConfig: ServerErrorConfig{
			RecoverConfig: RecoverConfig{
				RecoverDB: func(err error) (*sqlx.DB, error) {
					return &sqlx.DB{}, nil
				},
			},
		},
	}
	mockHandler := &Handler{}
	mockHandler.On("ServeHTTP", testifymock.Anything, testifymock.Anything)
	defer mockHandler.AssertExpectations(t)

	groupHandler := NewGroupHandler(db, state.queryForGroups, config)
	h := groupHandler.MiddlewareFunc(mockHandler)

	// Testing default settings without cache and recoverdb
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing returning no sql.ErrNoRows error
	rr = httptest.NewRecorder()
	state.reset()
	state.groupError = noRowsErr
	groupHandler.queryForGroups = state.queryForGroups

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing getting invalid db response with error
	rr = httptest.NewRecorder()
	state.reset()
	state.groupError = generalErr
	config.RecoverDB = nil
	groupHandler.config = config
	groupHandler.queryForGroups = state.queryForGroups

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing invalid json
	rr = httptest.NewRecorder()
	state.reset()
	state.groupError = invalidJSONErr
	groupHandler.config = config
	groupHandler.queryForGroups = state.queryForGroups

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing cachestore returns general error but gets info from db
	rr = httptest.NewRecorder()

	mockCacheStore := &MockCacheStore{}
	defer mockCacheStore.AssertExpectations(t)

	mockCacheStore.On("Get", testifymock.Anything).
		Return(nil, errors.New("errors"))

	state.reset()
	config.Cache = mockCacheStore
	groupHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing cachestore returns nil cache error but gets info from db
	// because we ignore the nil cache
	rr = httptest.NewRecorder()
	state.reset()

	mockCacheStore2 := &MockCacheStore{}
	defer mockCacheStore2.AssertExpectations(t)

	mockCacheStore2.On("Get", testifymock.Anything).
		Return(nil, ErrCacheNil)

	config.Cache = mockCacheStore2
	config.IgnoreCacheNil = true
	groupHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing cachestore returns nil cache error but we don't ignore
	// nil cache so move to next handler
	rr = httptest.NewRecorder()
	state.reset()

	config.Cache = mockCacheStore2
	config.IgnoreCacheNil = false
	groupHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing cachestore returns results but with json error
	rr = httptest.NewRecorder()
	state.reset()

	mockCacheStore3 := &MockCacheStore{}
	defer mockCacheStore3.AssertExpectations(t)

	mockCacheStore3.On("Get", testifymock.Anything).
		Return([]byte("invalid"), nil)

	config.Cache = mockCacheStore3
	groupHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing not having user
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/url", nil)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}
}

func TestRoutingHandlerUnitTest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/url", nil)
	mUser := MiddlewareUser{
		ID:    "1",
		Email: "email@email.com",
	}
	ctx := context.WithValue(req.Context(), MiddlewareUserCtxKey, mUser)
	req = req.WithContext(ctx)

	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	state := &middlewareState{}
	config := ServerErrorCacheConfig{
		ServerErrorConfig: ServerErrorConfig{
			RecoverConfig: RecoverConfig{
				RecoverDB: func(err error) (*sqlx.DB, error) {
					return &sqlx.DB{}, nil
				},
			},
		},
	}

	mockHandler := &Handler{}
	defer mockHandler.AssertExpectations(t)
	mockHandler.On("ServeHTTP", testifymock.Anything, testifymock.Anything)

	state.pRegex = generalErr
	routingHandler := NewRoutingHandler(
		db,
		state.queryForGroups,
		state.pathRegex,
		map[string]bool{
			defaultURL: true,
		},
		config,
	)
	h := routingHandler.MiddlewareFunc(mockHandler)

	// Testing pathexp failing
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing sql.ErrNoRows is returned and that we get client error
	rr = httptest.NewRecorder()
	state.reset()
	state.errType = errType1
	routingHandler.queryRoutes = state.queryForRoutes

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf(statusErrTxt, http.StatusForbidden, rr.Code)
	}

	// Testing that RecoverDB works and that we get a valid
	// url map but doesn't match so get forbidden status
	rr = httptest.NewRecorder()
	state.reset()
	state.errType = errType2
	routingHandler.queryRoutes = state.queryForRoutes

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf(statusErrTxt, http.StatusForbidden, rr.Code)
	}

	// Testing that RecoverDB works and that we query a valid
	// url map but doesn't match so get forbidden status
	rr = httptest.NewRecorder()
	state.reset()
	state.errType = errType2
	routingHandler.queryRoutes = state.queryForRoutes

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf(statusErrTxt, http.StatusForbidden, rr.Code)
	}

	// Testing that RecoverDB works but we get sql.ErrnoRows error
	// so get forbidden status
	rr = httptest.NewRecorder()
	state.reset()
	state.errType = errType3
	routingHandler.queryRoutes = state.queryForRoutes

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf(statusErrTxt, http.StatusForbidden, rr.Code)
	}

	// Testing we get server error because we don't have recoverdb
	// and err occurs
	rr = httptest.NewRecorder()
	state.reset()
	state.errType = errType2
	config.RecoverDB = nil
	routingHandler.queryRoutes = state.queryForRoutes
	routingHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing invalid json
	rr = httptest.NewRecorder()
	state.reset()
	state.errType = invalidJSONErr
	config.RecoverDB = nil
	routingHandler.queryRoutes = state.queryForRoutes

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing we query valid url map and find url
	rr = httptest.NewRecorder()
	state.reset()
	state.pRegex = defaultURL
	config.RecoverDB = nil

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing cache returns error but we get urls from db
	rr = httptest.NewRecorder()
	state.reset()
	state.pRegex = defaultURL
	config.RecoverDB = func(err error) (*sqlx.DB, error) {
		return &sqlx.DB{}, nil
	}

	mockCache1 := &MockCacheStore{}
	mockCache1.On("Get", testifymock.Anything).Return(nil, errors.New("errors"))
	config.Cache = mockCache1
	routingHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing cache returns ErrCacheNil but we ignore nil and
	// query from db
	rr = httptest.NewRecorder()
	state.reset()
	state.pRegex = defaultURL

	mockCache2 := &MockCacheStore{}
	mockCache2.On("Get", testifymock.Anything).Return(nil, ErrCacheNil)
	config.IgnoreCacheNil = true
	config.Cache = mockCache2
	routingHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

	// Testing cache returns ErrCacheNil and we don't ignore cache
	// so we should get forbidden err
	rr = httptest.NewRecorder()
	state.reset()
	state.pRegex = defaultURL

	config.IgnoreCacheNil = false
	config.Cache = mockCache2
	routingHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf(statusErrTxt, http.StatusForbidden, rr.Code)
	}

	// Testing cache returns invalid json
	rr = httptest.NewRecorder()
	state.reset()
	state.pRegex = defaultURL

	mockCache3 := &MockCacheStore{}
	mockCache3.On("Get", testifymock.Anything).Return([]byte("invalid"), nil)
	config.Cache = mockCache3
	routingHandler.config = config

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf(statusErrTxt, http.StatusInternalServerError, rr.Code)
	}

	// Testing when user not logged in that non user urls works
	// and user gets through
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/url", nil)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}
}
