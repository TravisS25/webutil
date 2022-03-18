package webutiltest

type TestLog interface {
	Errorf(format string, args ...interface{})
	Helper()
}
