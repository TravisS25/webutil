package webutiltest

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type TestLog interface {
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	Helper()
}
