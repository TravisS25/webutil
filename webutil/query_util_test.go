package webutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	reflect "reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	pkgerrors "github.com/pkg/errors"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

////////////////////////////////////////////////////////////
// REPLACEMENT FUNCTION TESTS
////////////////////////////////////////////////////////////

func TestQueryFunctionsUnitTest(t *testing.T) {
	var err error

	idField := "foo.id"
	nameField := "foo.name"
	filters := []Filter{
		{
			Field:    idField,
			Operator: "eq",
			Value:    1,
		},
	}
	groups := []Group{
		{
			Field: idField,
		},
	}
	sorts := []Sort{
		{
			Field: idField,
			Dir:   "asc",
		},
		{
			Field: nameField,
			Dir:   "asc",
		},
	}

	invalidSorts := []Sort{
		{
			Field: idField,
			Dir:   "invalid",
		},
	}

	fBytes, err := json.Marshal(&filters)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	gBytes, err := json.Marshal(&groups)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	sBytes, err := json.Marshal(&sorts)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	invalidSortBytes, err := json.Marshal(&invalidSorts)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	invalidJSONDecoding := "{ id: 1 }"
	filtersParam := "filters"
	sortsParam := "sorts"
	groupsParam := "groups"
	takeParam := "take"
	skipParam := "skip"
	take := "20"
	skip := "0"
	defaultQuery := "select foo.id from foo"
	query := defaultQuery

	filterFields := map[string]FieldConfig{
		idField: FieldConfig{
			DBField: idField,
			OperationConf: OperationConfig{
				CanSortBy:   true,
				CanFilterBy: true,
				CanGroupBy:  true,
			},
		},
		nameField: FieldConfig{
			DBField: nameField,
			OperationConf: OperationConfig{
				CanSortBy:   true,
				CanFilterBy: true,
				CanGroupBy:  true,
			},
		},
	}

	invalidSortFields := map[string]FieldConfig{
		idField: FieldConfig{
			DBField: idField,
			OperationConf: OperationConfig{
				CanSortBy:   false,
				CanFilterBy: true,
				CanGroupBy:  true,
			},
		},
		nameField: FieldConfig{
			DBField: nameField,
			OperationConf: OperationConfig{
				CanSortBy:   false,
				CanFilterBy: true,
				CanGroupBy:  true,
			},
		},
	}

	mockRequest1 := &MockFormRequest{}
	defer mockRequest1.AssertExpectations(t)
	mockRequest1.On("FormValue", filtersParam).Return(invalidJSONDecoding)

	if _, err = getValueResults(
		&query,
		true,
		mockRequest1,
		ParamConfig{},
		QueryConfig{},
		filterFields,
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if _, ok := pkgerrors.Cause(err).(*json.SyntaxError); !ok {
			t.Errorf("should have json.SyntaxError{} instance error\n")
		}
	}

	// query = defaultQuery
	// mockRequest.On("FormValue", filtersParam).Return(string(fBytes))
	// mockRequest.On("FormValue", filtersParam).Return(invalidJSONDecoding)

	// if _, err = getValueResults(
	// 	&query,
	// 	true,
	// 	mockRequest,
	// 	ParamConfig{},
	// 	QueryConfig{},
	// 	filterFields,
	// ); err == nil {
	// 	t.Errorf("should have error\n")
	// } else {
	// 	if _, ok := pkgerrors.Cause(err).(*json.SyntaxError); !ok {
	// 		t.Errorf("should have json.SyntaxError{} instance error\n")
	// 	}
	// }

	query = defaultQuery
	mockRequest2 := &MockFormRequest{}
	defer mockRequest2.AssertExpectations(t)
	mockRequest2.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest2.On("FormValue", groupsParam).Return(invalidJSONDecoding)

	if _, err = getValueResults(
		&query,
		true,
		mockRequest2,
		ParamConfig{},
		QueryConfig{},
		filterFields,
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if _, ok := pkgerrors.Cause(err).(*json.SyntaxError); !ok {
			t.Errorf("should have json.SyntaxError{} instance error\n")
		}
	}

	query = defaultQuery
	mockRequest3 := &MockFormRequest{}
	defer mockRequest3.AssertExpectations(t)
	mockRequest3.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest3.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest3.On("FormValue", sortsParam).Return(invalidJSONDecoding)

	if _, err = getValueResults(
		&query,
		true,
		mockRequest3,
		ParamConfig{},
		QueryConfig{},
		filterFields,
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if _, ok := pkgerrors.Cause(err).(*json.SyntaxError); !ok {
			t.Errorf("should have json.SyntaxError{} instance error\n")
		}
	}

	query = defaultQuery
	mockRequest4 := &MockFormRequest{}
	defer mockRequest4.AssertExpectations(t)
	mockRequest4.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest4.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest4.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest4.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest4.On("FormValue", takeParam).Return(string(take))
	mockRequest4.On("FormValue", skipParam).Return(string(skip))

	if _, err = getValueResults(
		&query,
		true,
		mockRequest4,
		ParamConfig{},
		QueryConfig{},
		filterFields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	query = defaultQuery
	mockRequest5 := &MockFormRequest{}
	defer mockRequest5.AssertExpectations(t)
	mockRequest5.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest5.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest5.On("FormValue", sortsParam).Return(string(sBytes))

	if _, err = getValueResults(
		&query,
		true,
		mockRequest5,
		ParamConfig{},
		QueryConfig{},
		invalidSortFields,
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if cErr, ok := pkgerrors.Cause(err).(*SortError); !ok {
			t.Errorf("should have SortError{} instance error\n")
			t.Errorf("err type: %s\n", reflect.TypeOf(err))
		} else {
			if !cErr.IsOperationError() {
				t.Errorf("should be operation error\n")
			}
		}
	}

	query = defaultQuery
	mockRequest6 := &MockFormRequest{}
	defer mockRequest6.AssertExpectations(t)
	mockRequest6.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest6.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest6.On("FormValue", sortsParam).Return(string(invalidSortBytes))
	mockRequest6.On("FormValue", sortsParam).Return(string(invalidSortBytes))

	if _, err = getValueResults(
		&query,
		true,
		mockRequest6,
		ParamConfig{},
		QueryConfig{},
		filterFields,
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if cErr, ok := pkgerrors.Cause(err).(*SortError); !ok {
			t.Errorf("should have &SortError{} instance error\n")
			t.Errorf("err type: %s\n", reflect.TypeOf(err))
		} else {
			if !cErr.IsDirError() {
				t.Errorf("should be dir error\n")
			}
		}
	}

	query = defaultQuery
	mockRequest7 := &MockFormRequest{}
	defer mockRequest7.AssertExpectations(t)
	mockRequest7.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest7.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest7.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest7.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest7.On("FormValue", takeParam).Return(string("invalid"))
	mockRequest7.On("FormValue", skipParam).Return(string(skip))

	if _, err = getValueResults(
		&query,
		true,
		mockRequest7,
		ParamConfig{},
		QueryConfig{},
		filterFields,
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if _, ok := pkgerrors.Cause(err).(*strconv.NumError); !ok {
			t.Errorf("should have &strconv.NumError{} instance error\n")
			t.Errorf("err type: %s\n", reflect.TypeOf(err))
		}
	}

	query = defaultQuery
	mockRequest8 := &MockFormRequest{}
	defer mockRequest1.AssertExpectations(t)
	mockRequest8.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest8.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest8.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest8.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest8.On("FormValue", takeParam).Return(string(take))
	mockRequest8.On("FormValue", skipParam).Return(string("invalid"))

	if _, err = getValueResults(
		&query,
		true,
		mockRequest8,
		ParamConfig{},
		QueryConfig{},
		filterFields,
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if _, ok := pkgerrors.Cause(err).(*strconv.NumError); !ok {
			t.Errorf("should have &strconv.NumError{} instance error\n")
			t.Errorf("err type: %s\n", reflect.TypeOf(err))
		}
	}

	/////////////////////////////////////////////////////////////////
	//------------------Testing GetPreQueryResults------------------
	/////////////////////////////////////////////////////////////////

	query = defaultQuery
	mockRequest9 := &MockFormRequest{}
	defer mockRequest9.AssertExpectations(t)
	mockRequest9.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest9.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest9.On("FormValue", sortsParam).Return(string(sBytes))

	if _, err = GetPreQueryResults(
		&query,
		invalidSortFields,
		mockRequest9,
		ParamConfig{},
		QueryConfig{},
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if cErr, ok := pkgerrors.Cause(err).(*SortError); !ok {
			t.Errorf("should have &SortError{} instance error\n")
		} else {
			if !cErr.IsOperationError() {
				t.Errorf("should be operation error\n")
			}
		}
	}

	query = defaultQuery
	mockRequest10 := &MockFormRequest{}
	defer mockRequest10.AssertExpectations(t)
	mockRequest10.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest10.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest10.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest10.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest10.On("FormValue", takeParam).Return(string(take))
	mockRequest10.On("FormValue", skipParam).Return(string(skip))

	if _, err = GetPreQueryResults(
		&query,
		filterFields,
		mockRequest10,
		ParamConfig{},
		QueryConfig{},
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	query = defaultQuery
	mockRequest11 := &MockFormRequest{}
	defer mockRequest11.AssertExpectations(t)
	mockRequest11.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest11.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest11.On("FormValue", sortsParam).Return(string(sBytes))

	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	if _, err = GetQueriedResults(
		&query,
		invalidSortFields,
		mockRequest11,
		db,
		ParamConfig{},
		QueryConfig{},
	); err == nil {
		t.Errorf("should have error\n")
	} else {
		if err, ok := pkgerrors.Cause(err).(*SortError); ok {
			if !err.IsOperationError() {
				t.Errorf("should be operation error\n")
			}
		} else {
			t.Errorf("should have &SortError{} instance\n")
		}
	}

	query = defaultQuery
	mockRequest12 := &MockFormRequest{}
	defer mockRequest12.AssertExpectations(t)
	mockRequest12.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest12.On("FormValue", groupsParam).Return(string(gBytes))
	mockRequest12.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest12.On("FormValue", sortsParam).Return(string(sBytes))
	mockRequest12.On("FormValue", takeParam).Return(string(take))
	mockRequest12.On("FormValue", skipParam).Return(string(skip))

	rows := sqlmock.NewRows([]string{"total"}).AddRow(20)
	mockDB.ExpectQuery("select").WillReturnRows(rows)

	if _, err = GetQueriedResults(
		&query,
		filterFields,
		mockRequest12,
		db,
		ParamConfig{},
		QueryConfig{},
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	query = defaultQuery
	mockRequest13 := &MockFormRequest{}
	defer mockRequest13.AssertExpectations(t)
	mockRequest13.On("FormValue", filtersParam).Return(string(fBytes))
	mockRequest13.On("FormValue", groupsParam).Return(string(gBytes))
	mockDB.ExpectQuery("select").WillReturnRows(rows)

	if _, err = GetCountResults(
		&query,
		filterFields,
		mockRequest13,
		db,
		ParamConfig{},
		QueryConfig{},
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestGetLimitWithOffsetValuesUnitTest(t *testing.T) {
	var err error
	var limitOffset *LimitOffset

	query := "select"
	invalidValue := "invalid"
	takeParam := "takeParam"
	skipParam := "skipParam"
	take := 20
	skip := 0

	mockRequest1 := &MockFormRequest{}
	defer mockRequest1.AssertExpectations(t)
	mockRequest1.On("FormValue", takeParam).Return("")
	mockRequest1.On("FormValue", skipParam).Return("")

	if _, err = GetLimitWithOffsetValues(
		mockRequest1,
		&query,
		takeParam,
		skipParam,
		100,
		false,
	); err != nil {
		t.Errorf("should not have errors\n")
		t.Errorf("er: %s\n", err.Error())
	}

	mockRequest2 := &MockFormRequest{}
	defer mockRequest2.AssertExpectations(t)
	mockRequest2.On("FormValue", takeParam).Return(invalidValue)
	mockRequest2.On("FormValue", skipParam).Return(invalidValue)

	if _, err = GetLimitWithOffsetValues(
		mockRequest2,
		&query,
		takeParam,
		skipParam,
		100,
		false,
	); err == nil {
		t.Errorf("should have errors\n")
	}

	mockRequest3 := &MockFormRequest{}
	defer mockRequest3.AssertExpectations(t)
	mockRequest3.On("FormValue", takeParam).Return(strconv.Itoa(take))
	mockRequest3.On("FormValue", skipParam).Return(strconv.Itoa(skip))

	if limitOffset, err = GetLimitWithOffsetValues(
		mockRequest3,
		&query,
		takeParam,
		skipParam,
		100,
		false,
	); err != nil {
		t.Errorf("should not have errors\n")
		t.Errorf("err: %s\n", err.Error())
	}

	if limitOffset.Take != take {
		t.Errorf("take should equal %d\n", take)
	}

	if limitOffset.Skip != skip {
		t.Errorf("skip should equal %d\n", skip)
	}
}

func TestGetGroupReplacementsUnitTest(t *testing.T) {
	var err error

	groupParam := "groups"
	invalid := "invalid"

	defaultQuery :=
		`
	select
		foo.id
	from
		foo
	`
	query := defaultQuery

	idField := "foo.id"
	bindVar := sqlx.DOLLAR
	limit := 100
	defaultConf := QueryConfig{
		SQLBindVar: &bindVar,
		TakeLimit:  &limit,
	}
	conf := defaultConf
	conf.PrependGroupFields = []Group{
		{
			Field: idField,
		},
	}

	fields := map[string]FieldConfig{
		idField: FieldConfig{
			DBField: idField,
			OperationConf: OperationConfig{
				CanFilterBy: true,
				CanGroupBy:  true,
				CanSortBy:   true,
			},
		},
	}

	mockRequest1 := &MockFormRequest{}
	defer mockRequest1.AssertExpectations(t)
	mockRequest1.On("FormValue", groupParam).Return(invalid)

	if _, err = GetGroupReplacements(
		mockRequest1,
		&query,
		groupParam,
		conf,
		fields,
	); err == nil {
		t.Errorf("should have error\n")
	}

	query = defaultQuery
	conf = defaultConf
	groups := []Group{
		{
			Field: idField,
		},
	}

	gBytes, err := json.Marshal(&groups)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	mockRequest2 := &MockFormRequest{}
	defer mockRequest2.AssertExpectations(t)
	mockRequest2.On("FormValue", groupParam).Return(string(gBytes))

	if _, err = GetGroupReplacements(
		mockRequest2,
		&query,
		groupParam,
		conf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if !strings.Contains(query, "group by") {
			t.Errorf("query should contain 'group by'\n")
		}
	}

	mockRequest3 := &MockFormRequest{}
	defer mockRequest3.AssertExpectations(t)
	mockRequest3.On("FormValue", groupParam).Return(string(gBytes))
	query = defaultQuery
	query +=
		`
	group by
		foo.id
	`

	if _, err = GetGroupReplacements(
		mockRequest3,
		&query,
		groupParam,
		conf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if strings.Count(query, "group") > 1 {
			t.Errorf("query should not contain more than one group by clause\n")
		}
	}

	conf.ExcludeGroups = true
	query = defaultQuery

	mockRequest4 := &MockFormRequest{}
	defer mockRequest4.AssertExpectations(t)
	if _, err = GetGroupReplacements(
		mockRequest4,
		&query,
		groupParam,
		conf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestGetSortReplacementsUnitTest(t *testing.T) {
	var err error

	sortParam := "sorts"
	invalid := "invalid"
	defaultQuery :=
		`
	select
		foo.id
	from
		foo
	`
	query := defaultQuery
	idField := "foo.id"
	bindVar := sqlx.DOLLAR
	limit := 100
	defaultConf := QueryConfig{
		SQLBindVar: &bindVar,
		TakeLimit:  &limit,
	}
	conf := defaultConf
	conf.PrependSortFields = []Sort{
		{
			Field: idField,
			Dir:   "asc",
		},
	}

	fields := map[string]FieldConfig{
		idField: FieldConfig{
			DBField: idField,
			OperationConf: OperationConfig{
				CanFilterBy: true,
				CanGroupBy:  true,
				CanSortBy:   true,
			},
		},
	}

	mockRequest1 := &MockFormRequest{}
	defer mockRequest1.AssertExpectations(t)
	mockRequest1.On("FormValue", sortParam).Return(invalid)

	if _, err = GetSortReplacements(
		mockRequest1,
		&query,
		sortParam,
		conf,
		fields,
	); err == nil {
		t.Errorf("should have error\n")
	}

	query = defaultQuery
	conf = defaultConf
	sorts := []Sort{
		{
			Field: idField,
			Dir:   "asc",
		},
	}

	sBytes, err := json.Marshal(&sorts)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	mockRequest2 := &MockFormRequest{}
	defer mockRequest2.AssertExpectations(t)
	mockRequest2.On("FormValue", sortParam).Return(string(sBytes))

	if _, err = GetSortReplacements(
		mockRequest2,
		&query,
		sortParam,
		conf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if !strings.Contains(query, "order by") {
			t.Errorf("query should contain 'order by'\n")
		}
	}

	mockRequest3 := &MockFormRequest{}
	defer mockRequest3.AssertExpectations(t)
	mockRequest3.On("FormValue", sortParam).Return(string(sBytes))

	query = defaultQuery
	query +=
		`
	order by
		foo.id
	`

	if _, err = GetSortReplacements(
		mockRequest3,
		&query,
		sortParam,
		conf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if strings.Count(query, "order") > 1 {
			t.Errorf("query should not contain more than one order by clause\n")
		}
	}

	conf.ExcludeSorts = true
	query = defaultQuery

	mockRequest4 := &MockFormRequest{}
	defer mockRequest4.AssertExpectations(t)

	if _, err = GetSortReplacements(
		mockRequest4,
		&query,
		sortParam,
		conf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestGetFilterReplacementsUnitTest(t *testing.T) {
	var err error

	filterParam := "filters"
	invalid := "invalid"
	defaultQuery :=
		`
	select
		foo.id
	from
		foo
	`
	query := defaultQuery
	idField := "foo.id"
	bindVar := sqlx.DOLLAR
	limit := 100
	defaultConf := QueryConfig{
		SQLBindVar: &bindVar,
		TakeLimit:  &limit,
	}
	conf := defaultConf
	conf.PrependFilterFields = []Filter{
		{
			Field:    idField,
			Operator: "eq",
			Value:    1,
		},
	}

	fields := map[string]FieldConfig{
		idField: FieldConfig{
			DBField: idField,
			OperationConf: OperationConfig{
				CanFilterBy: true,
				CanGroupBy:  true,
				CanSortBy:   true,
			},
		},
	}

	mockRequest1 := &MockFormRequest{}
	defer mockRequest1.AssertExpectations(t)
	mockRequest1.On("FormValue", filterParam).Return(invalid)

	if _, _, err = GetFilterReplacements(
		mockRequest1,
		&query,
		filterParam,
		conf,
		fields,
	); err == nil {
		t.Errorf("should have error\n")
	}

	query = defaultQuery
	conf = defaultConf
	filters := []Filter{
		{
			Field:    idField,
			Operator: "eq",
			Value:    2,
		},
	}

	fBytes, err := json.Marshal(&filters)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	mockRequest2 := &MockFormRequest{}
	defer mockRequest2.AssertExpectations(t)
	mockRequest2.On("FormValue", filterParam).Return(string(fBytes))

	if _, _, err = GetFilterReplacements(
		mockRequest2,
		&query,
		filterParam,
		conf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if !strings.Contains(query, "where") {
			t.Errorf("query should contain 'where'\n")
		}
	}

	mockRequest3 := &MockFormRequest{}
	defer mockRequest3.AssertExpectations(t)
	mockRequest3.On("FormValue", filterParam).Return(string(fBytes))
	query = defaultQuery
	query +=
		`
	where
		foo.id = 3
	`

	if _, _, err = GetFilterReplacements(
		mockRequest3,
		&query,
		filterParam,
		conf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if strings.Count(query, "where") > 1 {
			t.Errorf("query should not contain more than one where clause\n")
		}
	}

	mockRequest4 := &MockFormRequest{}
	defer mockRequest4.AssertExpectations(t)
	conf.ExcludeFilters = true
	query = defaultQuery

	if _, _, err = GetFilterReplacements(
		mockRequest4,
		&query,
		filterParam,
		conf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

////////////////////////////////////////////////////////////
// APPLY FUNCTION TESTS
////////////////////////////////////////////////////////////

func TestApplyOrderingUnitTest(t *testing.T) {
	query := "select"
	ApplyOrdering(&query, &Sort{Field: "foo.id", Dir: "asc"})

	if !strings.Contains(query, "order") {
		t.Fatalf("query should contain 'order'\n")
	}
}

func TestApplyLimitUnitTest(t *testing.T) {
	query := "select"
	ApplyLimit(&query)

	if !strings.Contains(query, "limit") {
		t.Errorf("query should contain 'limit'\n")
	}
}

func TestApplyFilterUnitTest(t *testing.T) {
	query :=
		`
	select
		foo.id
	from
		foo
	`
	temp := query
	filter := Filter{
		Operator: "eq",
		Value:    []interface{}{1},
		Field:    "foo.test",
	}
	ApplyFilter(&query, filter, true)

	if !strings.Contains(query, " and") {
		t.Errorf("did not apply `and` to query\n")
	}

	if !strings.Contains(query, " in ") {
		t.Errorf("did not apply `in` to query\n")
	}

	query = temp
	filter.Value = 1
	ApplyFilter(&query, filter, true)

	if !strings.Contains(query, filter.Field+" = ?") {
		t.Errorf("query should have equal operator\n")
	}
}

func TestApplySortUnitTest(t *testing.T) {
	query :=
		`
	select
		foo.id
	from
		foo
	`
	temp := query
	sort := Sort{
		Dir:   "asc",
		Field: "foo.test",
	}

	ApplySort(&query, sort, true)

	if !strings.Contains(query, "asc") {
		t.Errorf("query should contain `asc`\n")
	}

	query = temp
	sort.Dir = "desc"
	ApplySort(&query, sort, true)

	if !strings.Contains(query, "desc") {
		t.Errorf("query should contain `desc`\n")
	}
}

func TestApplyGroupUnitTest(t *testing.T) {
	query := "select"
	ApplyGroup(&query, Group{Field: "foo.test"}, true)

	if !strings.Contains(query, ",") {
		t.Errorf("query should contain `,`\n")
	}
}

////////////////////////////////////////////////////////////
// CHECK FUNCTION TESTS
////////////////////////////////////////////////////////////

func TestSortCheckUnitTest(t *testing.T) {
	var err error

	sort := Sort{
		Field: "foo.test",
		Dir:   "invalid",
	}

	if err = SortCheck(sort); err == nil {
		t.Fatalf("should be err\n")
	}

	sort.Dir = "asc"

	if err = SortCheck(sort); err != nil {
		t.Fatalf("should not have error\n")
	}
}

func TestFilterCheckUnitTest(t *testing.T) {
	var err error

	filter := Filter{
		Field:    "foo.id",
		Operator: "eq",
		Value:    []interface{}{errors.New("wrong type")},
	}

	if _, err = FilterCheck(filter); err == nil {
		t.Errorf("should have error\n")
	} else {
		if _, ok := err.(*SliceError); !ok {
			t.Errorf("should have *SliceError instance\n")
		}
	}

	filter.Value = []interface{}{1}

	if _, err = FilterCheck(filter); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	filter.Value = nil

	if _, err = FilterCheck(filter); err == nil {
		t.Errorf("should have error\n")
	} else {
		if fErr, ok := err.(*FilterError); !ok {
			t.Errorf("should have *FilterError instance\n")
		} else {
			if !fErr.IsValueError() {
				t.Errorf("should be value error\n")
			}
		}
	}

	filter.Value = errors.New("error")

	if _, err = FilterCheck(filter); err == nil {
		t.Errorf("should have error\n")
	} else {
		if fErr, ok := err.(*FilterError); !ok {
			t.Errorf("should have *FilterError instance\n")
		} else {
			if !fErr.IsValueError() {
				t.Errorf("should be value error\n")
			}
		}
	}

	filter.Value = 1

	if _, err = FilterCheck(filter); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

////////////////////////////////////////////////////////////
// DECODE FUNCTION TESTS
////////////////////////////////////////////////////////////

func TestDecodeFiltersUnitTest(t *testing.T) {
	var err error

	mockRequest := &MockFormRequest{}
	defer mockRequest.AssertExpectations(t)

	invalidParam := "invalidParam"
	param := "param"
	mockRequest.On("FormValue", invalidParam).Return("invalid")

	if _, err = DecodeFilters(mockRequest, invalidParam); err == nil {
		t.Errorf("should have error\n")
	}

	filters := []Filter{
		{
			Field:    "foo.id",
			Operator: "eq",
			Value:    1,
		},
	}

	fBytes, err := json.Marshal(&filters)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	paramValues := url.QueryEscape(string(fBytes))
	mockRequest.On("FormValue", param).Return(paramValues)

	if _, err = DecodeFilters(mockRequest, param); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestDecodeSortsUnitTest(t *testing.T) {
	var err error

	mockRequest := &MockFormRequest{}
	defer mockRequest.AssertExpectations(t)

	invalidParam := "invalidParam"
	param := "param"

	mockRequest.On("FormValue", invalidParam).Return("invalid")

	if _, err = DecodeSorts(mockRequest, invalidParam); err == nil {
		t.Errorf("should have error\n")
	}

	sorts := []Sort{
		{
			Field: "foo.id",
			Dir:   "asc",
		},
	}

	sBytes, err := json.Marshal(&sorts)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	paramValues := url.QueryEscape(string(sBytes))
	fmt.Printf("%s\n", paramValues)
	mockRequest.On("FormValue", param).Return(paramValues)

	if _, err = DecodeSorts(mockRequest, param); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
		t.Errorf("cause: %v\n", reflect.TypeOf(pkgerrors.Cause(err)))
	}
}

func TestDecodeGroupsUnitTest(t *testing.T) {
	var err error

	mockRequest := &MockFormRequest{}
	defer mockRequest.AssertExpectations(t)

	invalidParam := "invalidParam"
	param := "param"
	mockRequest.On("FormValue", invalidParam).Return("invalid")

	if _, err = DecodeGroups(mockRequest, invalidParam); err == nil {
		t.Errorf("should have error\n")
	}

	groups := []Group{
		{
			Field: "foo.id",
		},
	}

	gBytes, err := json.Marshal(&groups)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	paramValues := url.QueryEscape(string(gBytes))
	mockRequest.On("FormValue", param).Return(paramValues)

	if _, err = DecodeGroups(mockRequest, param); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

////////////////////////////////////////////////////////////
// UTIL FUNCTION TESTS
////////////////////////////////////////////////////////////

func TestReplaceGroupFieldsUnitTest(t *testing.T) {
	var err error

	idField := "foo.id"
	nameField := "foo.name"
	query :=
		`
	select
		foo.id
	from
		foo
	`
	groups := []Group{
		{
			Field: "invalid",
		},
	}
	fields := map[string]FieldConfig{
		idField: FieldConfig{
			DBField: idField,
			OperationConf: OperationConfig{
				CanGroupBy:  true,
				CanFilterBy: true,
				CanSortBy:   true,
			},
		},
		nameField: FieldConfig{
			DBField: nameField,
			OperationConf: OperationConfig{
				CanGroupBy:  true,
				CanFilterBy: true,
				CanSortBy:   true,
			},
		},
	}
	fields2 := map[string]FieldConfig{
		idField: FieldConfig{
			DBField: idField,
			OperationConf: OperationConfig{
				CanGroupBy:  false,
				CanFilterBy: true,
				CanSortBy:   true,
			},
		},
	}

	if err = ReplaceGroupFields(&query, groups, fields); err == nil {
		t.Errorf("should have error\n")
	}

	groups[0].Field = idField
	groups = append(groups, Group{Field: nameField})

	if err = ReplaceGroupFields(&query, groups, fields2); err == nil {
		t.Errorf("should have error\n")
	}

	if err = ReplaceGroupFields(&query, groups, fields); err != nil {
		t.Errorf("should not have error\n")
	}
}

func TestSetRowerResultsUnitTest(t *testing.T) {
	// var err error

	// generalErr := errors.New("error")
	// columns := []string{"id", "name"}
	// values := []interface{}{
	// 	int64(1),
	// 	"name",
	// }

	// mockCtrl := gomock.NewController(t)
	// defer mockCtrl.Finish()

	// mockRower := NewMockRower(mockCtrl)
	// mockRower.EXPECT().Columns().Return(nil, generalErr)

	// if err = SetRowerResults(mockRower, nil, CacheSetup{}); err == nil {
	// 	t.Fatalf("should have error\n")
	// }

	// mockRower.EXPECT().Columns().Return(columns)
	// mockRower.EXPECT().Next().Return(true)
	// mockRower.EXPECT().Scan().Return(generalErr)

	// if err = SetRowerResults(mockRower, nil, CacheSetup{}); err == nil {
	// 	t.Fatalf("should have error\n")
	// }

	// mockRower.EXPECT().Columns().Return(columns)
	// mockRower.EXPECT().Next().Return(true)
	// mockRower.EXPECT().Scan(values)

	// mockCacheStore := NewMockCacheStore(mockCtrl)
	// mockCacheStore.EXPECT()
}

func TestHasFilterOrServerErrorUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	err := errors.New("error")
	conf := ServerErrorConfig{}

	if HasFilterOrServerError(rr, nil, conf) {
		t.Errorf("should not have error\n")
	}

	filterErr := &FilterError{
		queryError: &queryError{},
	}
	rr = httptest.NewRecorder()

	if !HasFilterOrServerError(rr, filterErr, conf) {
		t.Errorf("should have error\n")
	} else {
		if rr.Code != http.StatusNotAcceptable {
			t.Errorf("status should be http.StatusNotAcceptable\n")
		}
	}

	rr = httptest.NewRecorder()

	if !HasFilterOrServerError(rr, err, conf) {
		t.Errorf("should have error\n")
	} else {
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status should be http.StatusInternalServerError\n")
		}
	}

	rr = httptest.NewRecorder()
	conf.RecoverDB = func(err error) error {
		return err
	}

	if !HasFilterOrServerError(rr, err, conf) {
		t.Errorf("should have error\n")
	} else {
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status should be http.StatusInternalServerError\n")
		}
	}

	rr = httptest.NewRecorder()
	conf.RecoverDB = func(err error) error {
		return nil
	}
	conf.RetryDB = func() error {
		return err
	}

	if !HasFilterOrServerError(rr, err, conf) {
		t.Errorf("should have error\n")
	} else {
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status should be http.StatusInternalServerError\n")
		}
	}

	rr = httptest.NewRecorder()
	conf.RecoverDB = func(err error) error {
		return nil
	}
	conf.RetryDB = func() error {
		return nil
	}

	if HasFilterOrServerError(rr, err, conf) {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}
