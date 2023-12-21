package webutil

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type mockMatch struct{}

func (m mockMatch) Match(expectedSQL, actualSQL string) error {
	return nil
}

func TestValidateUniquenessRuleJet(t *testing.T) {
	var err error
	var validator validateUniquenessRuleJet

	validator = validateUniquenessRuleJet{
		instanceValue: 1,
	}

	if err = validator.Validate("1"); err == nil {
		t.Errorf("should have error")
	} else {
		if !strings.Contains(err.Error(), "can't compare") {
			t.Errorf("should have compare error; got '%s'", err.Error())
		}
	}

	if err = validator.Validate(1); err != nil {
		t.Errorf("should not have error; got '%s'", err.Error())
	}

	validator = validateUniquenessRuleJet{
		instanceValue: []int64{1},
	}

	if err = validator.Validate([]int32{1}); err == nil {
		t.Errorf("should have error")
	} else {
		if !strings.Contains(err.Error(), "can't compare") {
			t.Errorf("should have compare error; got '%s'", err.Error())
		}
	}

	if err = validator.Validate([]int64{1}); err != nil {
		t.Errorf("should not have error; got '%s'", err.Error())
	}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(mockMatch{}))
	if err != nil {
		t.Fatalf(err.Error())
	}
	mockRows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery("select").WillReturnRows(mockRows)

	validator = validateUniquenessRuleJet{
		instanceValue: 1,
		validatorJet: &validatorJet{
			queryable: db,
		},
	}

	if err = validator.Validate(2); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}
}
