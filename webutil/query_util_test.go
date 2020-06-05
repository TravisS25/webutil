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
)

////////////////////////////////////////////////////////////
// REPLACEMENT FUNCTION TESTS
////////////////////////////////////////////////////////////

func TestGetValueResultsTest(t *testing.T) {
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

	invalidJSONEncoded := url.QueryEscape(string(invalidJSONDecoding))
	filterEncoded := url.QueryEscape(string(fBytes))
	groupEncoded := url.QueryEscape(string(gBytes))
	sortEncoded := url.QueryEscape(string(sBytes))

	invalidFilterURL := "/url?" + filtersParam + "=" + invalidJSONEncoded
	invalidGroupURL := "&" + groupsParam + "=" + invalidJSONEncoded
	invalidSortURL := "&" + sortsParam + "=" + invalidJSONEncoded

	filterURL := "/url?" + filtersParam + "=" + filterEncoded
	groupURL := "&" + groupsParam + "=" + groupEncoded
	sortURL := "&" + sortsParam + "=" + sortEncoded
	takeURL := "&" + takeParam + "=" + take + "&" + skipParam + "=" + skip

	req := httptest.NewRequest(http.MethodGet, invalidFilterURL, nil)

	if _, err = getValueResults(
		&query,
		nil,
		true,
		req,
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
	req = httptest.NewRequest(http.MethodGet, filterURL+invalidGroupURL, nil)

	if _, err = getValueResults(
		&query,
		nil,
		true,
		req,
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
	req = httptest.NewRequest(http.MethodGet, filterURL+groupURL+invalidSortURL, nil)

	if _, err = getValueResults(
		&query,
		nil,
		true,
		req,
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
	req = httptest.NewRequest(http.MethodGet, filterURL+groupURL+sortURL+takeURL, nil)

	if _, err = getValueResults(
		&query,
		nil,
		true,
		req,
		ParamConfig{},
		QueryConfig{},
		filterFields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	query = defaultQuery
	req = httptest.NewRequest(http.MethodGet, filterURL+sortURL+groupURL, nil)

	if _, err = getValueResults(
		&query,
		nil,
		true,
		req,
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

	/////////////////////////////////////////////////////////////////
	//------------------Testing GetPreQueryResults------------------
	/////////////////////////////////////////////////////////////////

	query = defaultQuery

	if _, err = GetPreQueryResults(
		&query,
		nil,
		invalidSortFields,
		req,
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

	if _, err = GetPreQueryResults(
		&query,
		nil,
		filterFields,
		req,
		ParamConfig{},
		QueryConfig{},
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	query = defaultQuery

	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	newDB := &sqlx.DB{
		DB: db,
	}

	if _, err = GetQueriedResults(
		query,
		nil,
		invalidSortFields,
		req,
		newDB,
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
	rows := sqlmock.NewRows([]string{"id", "foo"}).AddRow(1, "test")
	mockDB.ExpectQuery("select").WillReturnRows(rows)

	req = httptest.NewRequest(http.MethodGet, filterURL, nil)

	if _, err = GetQueriedResults(
		query,
		nil,
		filterFields,
		req,
		newDB,
		ParamConfig{},
		QueryConfig{},
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	query = defaultQuery
	rows = sqlmock.NewRows([]string{"count"}).AddRow(10)
	mockDB.ExpectQuery("select").WillReturnRows(rows)

	if _, err = GetCountResults(
		query,
		nil,
		filterFields,
		req,
		newDB,
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
	takeParam := "take"
	skipParam := "skip"
	take := 20
	skip := 0

	validURL := "/url?" + takeParam + "=" + strconv.Itoa(take) + "&" + skipParam +
		"=" + strconv.Itoa(skip)
	invalidURL := "/url?" + takeParam + "=foo" + "&" + skipParam + "=bar"

	req := httptest.NewRequest(http.MethodGet, "/url", nil)

	if _, err = GetLimitWithOffsetValues(
		req,
		&query,
		takeParam,
		skipParam,
		100,
	); err != nil {
		t.Errorf("should not have errors\n")
		t.Errorf("er: %s\n", err.Error())
	}

	req = httptest.NewRequest(http.MethodGet, invalidURL, nil)

	if _, err = GetLimitWithOffsetValues(
		req,
		&query,
		takeParam,
		skipParam,
		100,
	); err == nil {
		t.Errorf("should have errors\n")
	}

	req = httptest.NewRequest(http.MethodGet, validURL, nil)

	if limitOffset, err = GetLimitWithOffsetValues(
		req,
		&query,
		takeParam,
		skipParam,
		100,
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

	groups := []Group{
		{
			Field: idField,
		},
	}
	group := Group{
		Field: "foo",
	}

	gBytes, err := json.Marshal(&groups)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	invalidBytes, err := json.Marshal(&group)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	groupEncoded := url.QueryEscape(string(gBytes))
	invalidEncoded := url.QueryEscape(string(invalidBytes))
	invalidGroupURL := "/url?" + groupParam + "=" + invalidEncoded
	groupURL := "/url?" + groupParam + "=" + groupEncoded

	req := httptest.NewRequest(http.MethodGet, invalidGroupURL, nil)

	if _, err = GetGroupReplacements(
		req,
		&query,
		groupParam,
		defaultConf,
		fields,
	); err == nil {
		t.Errorf("should have error\n")
	}

	query = defaultQuery
	req = httptest.NewRequest(http.MethodGet, groupURL, nil)

	if _, err = GetGroupReplacements(
		req,
		&query,
		groupParam,
		defaultConf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if !strings.Contains(query, "group by") {
			t.Errorf("query should contain 'group by'\n")
		}
	}

	query = defaultQuery
	query +=
		`
	group by
		foo.id
	`

	if _, err = GetGroupReplacements(
		req,
		&query,
		groupParam,
		defaultConf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if strings.Count(query, "group") > 1 {
			t.Errorf("query should not contain more than one group by clause\n")
		}
	}
}

func TestGetSortReplacementsUnitTest(t *testing.T) {
	var err error

	sortParam := "sorts"
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

	sorts := []Sort{
		{
			Field: idField,
			Dir:   "asc",
		},
	}
	sort := Sort{
		Field: "foo",
		Dir:   "asc",
	}

	sBytes, err := json.Marshal(&sorts)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	invalidBytes, err := json.Marshal(&sort)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
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

	sortEncoded := url.QueryEscape(string(sBytes))
	invalidEncoded := url.QueryEscape(string(invalidBytes))
	invalidSortURL := "/url?" + sortParam + "=" + invalidEncoded
	sortURL := "/url?" + sortParam + "=" + sortEncoded

	req := httptest.NewRequest(http.MethodGet, invalidSortURL, nil)

	if _, err = GetSortReplacements(
		req,
		&query,
		sortParam,
		defaultConf,
		fields,
	); err == nil {
		t.Errorf("should have error\n")
	}

	query = defaultQuery
	req = httptest.NewRequest(http.MethodGet, sortURL, nil)

	if _, err = GetSortReplacements(
		req,
		&query,
		sortParam,
		defaultConf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if !strings.Contains(query, "order by") {
			t.Errorf("query should contain 'order by'\n")
		}
	}

	query = defaultQuery
	query +=
		`
	order by
		foo.id
	`

	if _, err = GetSortReplacements(
		req,
		&query,
		sortParam,
		defaultConf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if strings.Count(query, "order") > 1 {
			t.Errorf("query should not contain more than one order by clause\n")
		}
	}
}

func TestGetFilterReplacementsUnitTest(t *testing.T) {
	var err error

	filterParam := "filters"
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

	filters := []Filter{
		{
			Field:    idField,
			Operator: "eq",
			Value:    2,
		},
	}
	filter := Filter{
		Field: "foo",
	}

	fBytes, err := json.Marshal(&filters)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	invalidBytes, err := json.Marshal(&filter)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
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

	filterEncoded := url.QueryEscape(string(fBytes))
	invalidEncoded := url.QueryEscape(string(invalidBytes))
	invalidFilterURL := "/url?" + filterParam + "=" + invalidEncoded
	filterURL := "/url?" + filterParam + "=" + filterEncoded

	req := httptest.NewRequest(http.MethodGet, invalidFilterURL, nil)

	if _, err = GetFilterReplacements(
		req,
		&query,
		filterParam,
		defaultConf,
		fields,
	); err == nil {
		t.Errorf("should have error\n")
	}

	query = defaultQuery
	req = httptest.NewRequest(http.MethodGet, filterURL, nil)

	if _, err = GetFilterReplacements(
		req,
		&query,
		filterParam,
		defaultConf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if !strings.Contains(query, "where") {
			t.Errorf("query should contain 'where'\n")
		}
	}

	query = defaultQuery
	query +=
		`
	where
		foo.id = 3
	`

	if _, err = GetFilterReplacements(
		req,
		&query,
		filterParam,
		defaultConf,
		fields,
	); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	} else {
		if strings.Count(query, "where") > 1 {
			t.Errorf("query should not contain more than one where clause\n")
		}
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

	filter := Filter{
		Field:    "foo.id",
		Operator: "eq",
		Value:    1,
	}

	invalidBytes, err := json.Marshal(&filter)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	filterParam := "filters"
	filterEncoded := url.QueryEscape(string(fBytes))
	invalidEncoded := url.QueryEscape(string(invalidBytes))

	filterURL := "/url?" + filterParam + "=" + filterEncoded
	invalidFilterURL := "/url?" + filterParam + "=" + invalidEncoded

	req := httptest.NewRequest(http.MethodGet, invalidFilterURL, nil)

	if _, err = DecodeFilters(req, filterParam); err == nil {
		t.Errorf("should have error\n")
	}

	req = httptest.NewRequest(http.MethodGet, filterURL, nil)

	if _, err = DecodeFilters(req, filterParam); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestDecodeSortsUnitTest(t *testing.T) {
	var err error

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

	sort := Sort{
		Field: "foo.id",
		Dir:   "asc",
	}

	invalidSortBytes, err := json.Marshal(&sort)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	sortParam := "sorts"
	sortEncoded := url.QueryEscape(string(sBytes))
	invalidEncoded := url.QueryEscape(string(invalidSortBytes))

	sortURL := "/url?" + sortParam + "=" + sortEncoded
	invalidSortURL := "/url?" + sortParam + "=" + invalidEncoded

	req := httptest.NewRequest(http.MethodGet, invalidSortURL, nil)

	if _, err = DecodeSorts(req, sortParam); err == nil {
		t.Errorf("should have error\n")
	}

	req = httptest.NewRequest(http.MethodGet, sortURL, nil)

	if _, err = DecodeSorts(req, sortParam); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
		t.Errorf("cause: %v\n", reflect.TypeOf(pkgerrors.Cause(err)))
	}
}

func TestDecodeGroupsUnitTest(t *testing.T) {
	var err error

	groups := []Group{
		{
			Field: "foo.id",
		},
	}

	group := Group{
		Field: "foo.id",
	}

	gBytes, err := json.Marshal(&groups)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	invalidBytes, err := json.Marshal(&group)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	groupParam := "groups"
	groupEncoded := url.QueryEscape(string(gBytes))
	invalidGroupEncoded := url.QueryEscape(string(invalidBytes))

	groupURL := "/url?" + groupParam + "=" + groupEncoded
	invalidGroupURL := "/url?" + groupParam + "=" + invalidGroupEncoded

	req := httptest.NewRequest(http.MethodGet, invalidGroupURL, nil)

	if _, err = DecodeGroups(req, groupParam); err == nil {
		t.Errorf("should have error\n")
	}

	req = httptest.NewRequest(http.MethodGet, groupURL, nil)

	if _, err = DecodeGroups(req, groupParam); err != nil {
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

	if err = ReplaceGroupFields(&query, groups, fields, false); err == nil {
		t.Errorf("should have error\n")
	}

	groups[0].Field = idField
	groups = append(groups, Group{Field: nameField})

	if err = ReplaceGroupFields(&query, groups, fields2, false); err == nil {
		t.Errorf("should have error\n")
	}

	if err = ReplaceGroupFields(&query, groups, fields, false); err != nil {
		t.Errorf("should not have error\n")
	}
}

func TestHasFilterOrServerErrorUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/url", nil)
	err := errors.New("error")
	status := http.StatusNotAcceptable
	conf := ServerErrorConfig{
		RecoverConfig: RecoverConfig{
			//DBInterfaceRecover: &testAPI{},
		},
	}

	if HasFilterOrServerError(rr, req, nil, nil, status, conf) {
		t.Errorf("should not have error\n")
	}

	filterErr := &FilterError{
		queryError: &queryError{},
	}
	rr = httptest.NewRecorder()

	if !HasFilterOrServerError(rr, req, filterErr, nil, status, conf) {
		t.Errorf("should have error\n")
	} else {
		if rr.Code != http.StatusNotAcceptable {
			t.Errorf("status should be http.StatusNotAcceptable\n")
		}
	}

	rr = httptest.NewRecorder()

	if !HasFilterOrServerError(rr, req, err, nil, status, conf) {
		t.Errorf("should have error\n")
	} else {
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status should be http.StatusInternalServerError\n")
		}
	}

	rr = httptest.NewRecorder()
	conf.RecoverDB = func(err error) (*sqlx.DB, error) {
		return nil, err
	}

	if !HasFilterOrServerError(rr, req, err, nil, status, conf) {
		t.Errorf("should have error\n")
	} else {
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status should be http.StatusInternalServerError\n")
		}
	}

	rr = httptest.NewRecorder()
	conf.RecoverDB = func(err error) (*sqlx.DB, error) {
		return &sqlx.DB{}, nil
	}

	if !HasFilterOrServerError(rr, req, err, nil, status, conf) {
		t.Errorf("should have error\n")
	} else {
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status should be http.StatusInternalServerError\n")
		}
	}
}
