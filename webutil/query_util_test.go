package webutil

import (
	"encoding/json"
	"errors"
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
	gomock "github.com/golang/mock/gomock"
)

////////////////////////////////////////////////////////////
// REPLACEMENT FUNCTION TESTS
////////////////////////////////////////////////////////////

func TestQueryFunctionsUnitTest(t *testing.T) {
	var err error

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

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

	mockRequest := NewMockFormRequest(mockCtrl)
	mockRequest.EXPECT().FormValue(filtersParam).Return(invalidJSONDecoding)

	if _, err = getValueResults(
		&query,
		true,
		mockRequest,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(invalidJSONDecoding)

	if _, err = getValueResults(
		&query,
		true,
		mockRequest,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(invalidJSONDecoding)

	if _, err = getValueResults(
		&query,
		true,
		mockRequest,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(invalidJSONDecoding)

	if _, err = getValueResults(
		&query,
		true,
		mockRequest,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(takeParam).Return(take)
	mockRequest.EXPECT().FormValue(skipParam).Return(skip)

	if _, err = getValueResults(
		&query,
		true,
		mockRequest,
		ParamConfig{},
		QueryConfig{},
		filterFields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	query = defaultQuery
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))

	if _, err = getValueResults(
		&query,
		true,
		mockRequest,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(invalidSortBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(invalidSortBytes))

	if _, err = getValueResults(
		&query,
		true,
		mockRequest,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(takeParam).Return("invalid")
	mockRequest.EXPECT().FormValue(skipParam).Return(skip)

	if _, err = getValueResults(
		&query,
		true,
		mockRequest,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(takeParam).Return(take)
	mockRequest.EXPECT().FormValue(skipParam).Return("invalid")

	if _, err = getValueResults(
		&query,
		true,
		mockRequest,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))

	if _, err = GetPreQueryResults(
		&query,
		invalidSortFields,
		mockRequest,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(takeParam).Return(take)
	mockRequest.EXPECT().FormValue(skipParam).Return(skip)

	if _, err = GetPreQueryResults(
		&query,
		filterFields,
		mockRequest,
		ParamConfig{},
		QueryConfig{},
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	query = defaultQuery
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))

	mockQuerier := NewMockQuerier(mockCtrl)

	if _, err = GetQueriedResults(
		&query,
		invalidSortFields,
		mockRequest,
		mockQuerier,
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
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(sortsParam).Return(string(sBytes))
	mockRequest.EXPECT().FormValue(takeParam).Return(take)
	mockRequest.EXPECT().FormValue(skipParam).Return(skip)

	mockQuerier.EXPECT().Query(gomock.Any(), gomock.Any())

	if _, err = GetQueriedResults(
		&query,
		filterFields,
		mockRequest,
		mockQuerier,
		ParamConfig{},
		QueryConfig{},
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	query = defaultQuery
	mockRequest.EXPECT().FormValue(filtersParam).Return(string(fBytes))
	mockRequest.EXPECT().FormValue(groupsParam).Return(string(gBytes))

	matcher := sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		return nil
	})

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	rows := sqlmock.NewRows([]string{"total"}).AddRow(20)
	mock.ExpectQuery("select").WillReturnRows(rows)

	if _, err = GetCountResults(
		&query,
		filterFields,
		mockRequest,
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

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	query := "select"
	invalidValue := "invalid"
	takeParam := "takeParam"
	skipParam := "skipParam"
	take := 20
	skip := 0
	mockRequest := NewMockFormRequest(mockCtrl)
	mockRequest.EXPECT().FormValue(takeParam).Return("")
	mockRequest.EXPECT().FormValue(skipParam).Return("")

	if _, err = GetLimitWithOffsetValues(
		mockRequest,
		&query,
		takeParam,
		skipParam,
		100,
		false,
	); err != nil {
		t.Errorf("should not have errors\n")
		t.Errorf("er: %s\n", err.Error())
	}

	mockRequest.EXPECT().FormValue(takeParam).Return(invalidValue)
	mockRequest.EXPECT().FormValue(skipParam).Return(invalidValue)

	if _, err = GetLimitWithOffsetValues(
		mockRequest,
		&query,
		takeParam,
		skipParam,
		100,
		false,
	); err == nil {
		t.Errorf("should have errors\n")
	}

	mockRequest.EXPECT().FormValue(takeParam).Return(invalidValue)
	mockRequest.EXPECT().FormValue(skipParam).Return(invalidValue)

	if _, err = GetLimitWithOffsetValues(
		mockRequest,
		&query,
		takeParam,
		skipParam,
		100,
		false,
	); err == nil {
		t.Errorf("should have errors\n")
	}

	mockRequest.EXPECT().FormValue(takeParam).Return(strconv.Itoa(take))
	mockRequest.EXPECT().FormValue(skipParam).Return(strconv.Itoa(skip))

	if limitOffset, err = GetLimitWithOffsetValues(
		mockRequest,
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

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	groupParam := "groups"
	invalid := "invalid"
	mockRequest := NewMockFormRequest(mockCtrl)
	mockRequest.EXPECT().FormValue(groupParam).Return(invalid)
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

	if _, err = GetGroupReplacements(
		mockRequest,
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

	mockRequest.EXPECT().FormValue(groupParam).Return(string(gBytes))

	if _, err = GetGroupReplacements(
		mockRequest,
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

	mockRequest.EXPECT().FormValue(groupParam).Return(string(gBytes))
	query = defaultQuery
	query +=
		`
	group by
		foo.id
	`

	if _, err = GetGroupReplacements(
		mockRequest,
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

	if _, err = GetGroupReplacements(
		mockRequest,
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

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

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

	mockRequest := NewMockFormRequest(mockCtrl)
	mockRequest.EXPECT().FormValue(sortParam).Return(invalid)

	if _, err = GetSortReplacements(
		mockRequest,
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

	mockRequest.EXPECT().FormValue(sortParam).Return(string(sBytes))

	if _, err = GetSortReplacements(
		mockRequest,
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

	mockRequest.EXPECT().FormValue(sortParam).Return(string(sBytes))
	query = defaultQuery
	query +=
		`
	order by
		foo.id
	`

	if _, err = GetSortReplacements(
		mockRequest,
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

	if _, err = GetSortReplacements(
		mockRequest,
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

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

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

	mockRequest := NewMockFormRequest(mockCtrl)
	mockRequest.EXPECT().FormValue(filterParam).Return(invalid)

	if _, _, err = GetFilterReplacements(
		mockRequest,
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

	mockRequest.EXPECT().FormValue(filterParam).Return(string(fBytes))

	if _, _, err = GetFilterReplacements(
		mockRequest,
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

	mockRequest.EXPECT().FormValue(filterParam).Return(string(fBytes))
	query = defaultQuery
	query +=
		`
	where
		foo.id = 3
	`

	if _, _, err = GetFilterReplacements(
		mockRequest,
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

	conf.ExcludeFilters = true
	query = defaultQuery

	if _, _, err = GetFilterReplacements(
		mockRequest,
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

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRequest := NewMockFormRequest(mockCtrl)
	mockRequest.EXPECT().FormValue("test").Return("invalid")

	if _, err = DecodeFilters(mockRequest, "test"); err == nil {
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

	params := url.QueryEscape(string(fBytes))
	mockRequest.EXPECT().FormValue("test").Return(params)

	if _, err = DecodeFilters(mockRequest, "test"); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestDecodeSortsUnitTest(t *testing.T) {
	var err error

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRequest := NewMockFormRequest(mockCtrl)
	mockRequest.EXPECT().FormValue("test").Return("invalid")

	if _, err = DecodeSorts(mockRequest, "test"); err == nil {
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

	params := url.QueryEscape(string(sBytes))
	mockRequest.EXPECT().FormValue("test").Return(params)

	if _, err = DecodeSorts(mockRequest, "test"); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestDecodeGroupsUnitTest(t *testing.T) {
	var err error

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRequest := NewMockFormRequest(mockCtrl)
	mockRequest.EXPECT().FormValue("test").Return("invalid")

	if _, err = DecodeGroups(mockRequest, "test"); err == nil {
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

	params := url.QueryEscape(string(gBytes))
	mockRequest.EXPECT().FormValue("test").Return(params)

	if _, err = DecodeGroups(mockRequest, "test"); err != nil {
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

func TestHasFilterErrorUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	err := errors.New("error")
	conf := ErrorResponse{}

	if HasFilterOrServerError(rr, nil, conf) {
		t.Errorf("should not have error\n")
	}

	filterErr := &FilterError{}
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
}
