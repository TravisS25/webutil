package webutiltest

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type TestLog interface {
	Errorf(format string, args ...interface{})
	Helper()
}
