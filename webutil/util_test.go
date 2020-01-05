package webutil

import (
	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const (
	// WebUtilConfig is env var that should be set to point to file that
	// contains test config settings for integration tests
	WebUtilConfig = "WEB_UTIL_CONFIG"
)

var (
	sqlAnyMatcher = sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		return nil
	})
)
