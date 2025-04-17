package webutil

import (
	"net/http/httptest"
	"testing"
)

// const (
// 	cookieName   = "user"
// 	statusErrTxt = "Status should be %d; got %d"
// )

// var (
// 	keyPairs = [][]byte{
// 		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
// 		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
// 	}
// )

// func getMockSession(cookieName string) *sessions.Session {
// 	mockSession := &MockSessionStore{}
// 	return sessions.NewSession(mockSession, cookieName)
// }

// func TestSetSecureCookieUnitTest(t *testing.T) {
// 	rr := httptest.NewRecorder()
// 	session := getMockSession(cookieName)

// 	if _, err := SetSecureCookie(rr, session, keyPairs...); err != nil {
// 		t.Fatalf("err: %s\n", err.Error())
// 	}
// }

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
