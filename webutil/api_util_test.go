package webutil

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/gorilla/sessions"

	"github.com/gorilla/securecookie"
)

const (
	cookieName = "user"
)

var (
	keyPairs = [][]byte{
		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
	}
)

func getMockSession(mockCtrl *gomock.Controller, cookieName string) *sessions.Session {
	mockSession := NewMockSessionStore(mockCtrl)
	return sessions.NewSession(mockSession, cookieName)
}

func TestDecodeCookieUnitTest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/url", nil)
	session := getMockSession(mockCtrl, cookieName)
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
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rr := httptest.NewRecorder()
	session := getMockSession(mockCtrl, cookieName)

	if err := SetSecureCookie(rr, session, keyPairs...); err != nil {
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
	user := &middlewareUser{
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

func TestHasBodyErrorUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/url", nil)

	if !HasBodyError(rr, req, HTTPResponseConfig{}) {
		t.Fatalf("should have body error\n")
	}

	buf := &bytes.Buffer{}
	r := ioutil.NopCloser(buf)
	req.Body = r

	if HasBodyError(rr, req, HTTPResponseConfig{}) {
		t.Fatalf("should not have body error\n")
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
