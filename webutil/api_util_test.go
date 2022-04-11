package webutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
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
