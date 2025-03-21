package webutiltest

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type TestLog interface {
	Errorf(string, ...any)
	Fatalf(string, ...any)
	Helper()
}
