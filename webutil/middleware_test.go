package webutil

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TravisS25/httputil/cacheutil/cachetest"
	gomock "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

const (
	decodeErr      = "decodeErr"
	internalErr    = "internalErr"
	noRowsErr      = "noRowsErr"
	generalErr     = "generalErr"
	invalidJSONErr = "invalidJSONErr"
)

var (
	// This should be used for read only
	mUser = middlewareUser{
		ID:    "1",
		Email: "someemail@email.com",
	}

// queryDB = func(w http.ResponseWriter, req *http.Request, db Querier) ([]byte, error) {
// 	return nil, errors.New("errors")
// }

// This should be used for read only
// mockHandler = &webutiltest.MockHandler{
// 	ServeHTTPFunc: func(w http.ResponseWriter, r *http.Request) {},
// }
)

type mockHandler struct{}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func TestAuthHandlerUnitTest(t *testing.T) {
	//var err error

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	queryUser := "queryUser"
	querySession := "querySession"
	mockDB := NewMockQuerier(mockCtrl)
	// mockSessionStore := NewMockSessionStore(mockCtrl)
	req := httptest.NewRequest(http.MethodGet, "/url", nil)
	mHandler := &mockHandler{}
	queryForUser := func(w http.ResponseWriter, r *http.Request, db Querier) ([]byte, error) {
		if r.Header.Get(queryUser) == decodeErr {
			return nil, cachetest.NewMockSessionError(nil, "Decode cookie error", false, true, false)
		}

		if r.Header.Get(queryUser) == internalErr {
			fmt.Printf("made to internal error\n")
			return nil, cachetest.NewMockSessionError(nil, "Internal cookie error", false, false, true)
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

		u := mUser

		if r.Header.Get(querySession) == noRowsErr {
			u.ID = "0"
		}

		if r.Header.Get(querySession) == generalErr {
			u.ID = "-1"
		}

		return json.Marshal(&u)
	}

	authHandler := NewAuthHandler(mockDB, queryForUser, AuthHandlerConfig{})

	// Testing default settings without cache
	rr := httptest.NewRecorder()
	h := authHandler.MiddlewareFunc(mHandler)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(statusErrTxt, http.StatusOK, rr.Code)
	}

}
