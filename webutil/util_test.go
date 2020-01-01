package webutil

import sqlmock "github.com/DATA-DOG/go-sqlmock"

var (
	sqlAnyMatcher = sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		return nil
	})
)
