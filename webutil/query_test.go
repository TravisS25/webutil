package webutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestGetQueryBuilder(t *testing.T) {
	var req *http.Request
	var err error
	var urlVals url.Values
	var builder sq.SelectBuilder
	var jsonBytes []byte
	var filters []Filter
	var sorts []Order

	cfg := QueryConfig{
		FilterParam: "filters",
		OrderParam:  "sorts",
		LimitParam:  "take",
		OffsetParam: "skip",

		Limit:  100,
		OffSet: 10000,
	}
	dbFields := DbFields{
		"user.id": FieldConfig{
			DBField:      "user.name",
			OperationCfg: OperationConfig{},
		},
		"user.name": FieldConfig{
			DBField: "user.name",
			OperationCfg: OperationConfig{
				CanFilterBy: true,
				CanSortBy:   true,
			},
		},
		"phone.number": FieldConfig{
			DBField: "user.number",
			OperationCfg: OperationConfig{
				CanFilterBy: true,
			},
		},
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, "invalid")

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err == nil {
		t.Errorf("should have error\n")
	} else {
		if !strings.Contains(err.Error(), "invalid filter parameter") {
			t.Errorf("error should be '%s'; got '%s'", "invalid filter parameter", err.Error())
		}
	}

	// ----------------------------------------------------------------------------------

	urlVals = url.Values{}
	urlVals.Add(cfg.OrderParam, "invalid")

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err == nil {
		t.Errorf("should have error\n")
	} else {
		if !strings.Contains(err.Error(), "invalid order parameter") {
			t.Errorf("error should be '%s'; got '%s'", "invalid order parameter", err.Error())
		}
	}

	// ----------------------------------------------------------------------------------

	sorts = []Order{
		{
			Dir: "invalid",
		},
	}

	if jsonBytes, err = json.Marshal(sorts); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.OrderParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err == nil {
		t.Errorf("should have error\n")
	} else {
		if !strings.Contains(err.Error(), "invalid sort dir for field") {
			t.Errorf("error should be '%s'; got '%s'", "invalid sort dir for field", err.Error())
		}
	}

	// ----------------------------------------------------------------------------------

	sorts = []Order{
		{
			Dir: "desc",
		},
	}

	if jsonBytes, err = json.Marshal(sorts); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.OrderParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err == nil {
		t.Errorf("should have error\n")
	} else {
		if !strings.Contains(err.Error(), `invalid field "" for order parameter`) {
			t.Errorf("error should be '%s'; got '%s'", `invalid field "" for order parameter`, err.Error())
		}
	}

	// ----------------------------------------------------------------------------------

	sorts = []Order{
		{
			Dir:   "desc",
			Field: "user.id",
		},
	}

	if jsonBytes, err = json.Marshal(sorts); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.OrderParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err == nil {
		t.Errorf("should have error\n")
	} else {
		if !strings.Contains(err.Error(), "can not be ordered") {
			t.Errorf("error should be '%s'; got '%s'", "can not be ordered", err.Error())
		}
	}

	// ----------------------------------------------------------------------------------

	sorts = []Order{
		{
			Dir:   "desc",
			Field: "user.name",
		},
	}

	if jsonBytes, err = json.Marshal(sorts); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.OrderParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field: "invalid",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err == nil {
		t.Errorf("should have error\n")
	} else {
		if !strings.Contains(err.Error(), "invalid field") {
			t.Errorf("error should be '%s'; got '%s'", "invalid field", err.Error())
		}
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field: "user.id",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err == nil {
		t.Errorf("should have error\n")
	} else {
		if !strings.Contains(err.Error(), "can not be filtered") {
			t.Errorf("error should be '%s'; got '%s'", "can not be filtered", err.Error())
		}
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field: "user.name",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err == nil {
		t.Errorf("should have error\n")
	} else {
		if !strings.Contains(err.Error(), "does not contain value") {
			t.Errorf("error should be '%s'; got '%s'", "does not contain value", err.Error())
		}
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field: "user.name",
			Value: "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err == nil {
		t.Errorf("should have error\n")
	} else {
		if !strings.Contains(err.Error(), "invalid operator for field") {
			t.Errorf("error should be '%s'; got '%s'", "invalid operator for field", err.Error())
		}
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "eq",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "neq",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "startswith",
			Value:    []string{"foo"},
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "endswith",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "contains",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "doesnotcontain",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "isnull",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "isnotnull",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "isempty",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "isnotempty",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "lt",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "lte",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "gt",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	filters = []Filter{
		{
			Field:    "user.name",
			Operator: "gte",
			Value:    "foo",
		},
	}

	if jsonBytes, err = json.Marshal(filters); err != nil {
		t.Fatalf(err.Error())
	}

	urlVals = url.Values{}
	urlVals.Add(cfg.FilterParam, string(jsonBytes))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}

	// ----------------------------------------------------------------------------------

	urlVals = url.Values{}
	urlVals.Add(cfg.LimitParam, "1000")

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder = sq.Select(
		"user.id",
		"user.name",
		"phone.number",
	).
		From("user").
		Join("phone on phone.user_id = user.id")

	if _, err = GetQueryBuilder(req, builder, dbFields, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}
}

func TestGetQueryBuilderFilterGroups(t *testing.T) {
	cfg := QueryConfig{
		FilterParam: "filters",
	}
	dbFields := DbFields{
		"user.name": FieldConfig{
			DBField: "user.name",
			OperationCfg: OperationConfig{
				CanFilterBy: true,
			},
		},
		"phone.number": FieldConfig{
			DBField: "user.number",
			OperationCfg: OperationConfig{
				CanFilterBy: true,
			},
		},
		"user.id": FieldConfig{
			DBField:      "user.id",
			OperationCfg: OperationConfig{},
		},
	}

	newBuilder := func() sq.SelectBuilder {
		return sq.Select(
			"user.id",
			"user.name",
			"phone.number",
		).
			From("user").
			Join("phone on phone.user_id = user.id")
	}

	newReq := func(filterJSON string) *http.Request {
		urlVals := url.Values{}
		urlVals.Add(cfg.FilterParam, filterJSON)
		return httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	}

	buildSQL := func(t *testing.T, filterJSON string) (string, []any, error) {
		t.Helper()

		builder, err := GetQueryBuilder(newReq(filterJSON), newBuilder(), dbFields, cfg)
		if err != nil {
			return "", nil, err
		}

		return builder.ToSql()
	}

	t.Run("top level or", func(t *testing.T) {
		sqlStr, args, err := buildSQL(t, `{"logic":"or","items":[`+
			`{"field":"user.name","operator":"eq","value":"foo"},`+
			`{"field":"phone.number","operator":"contains","value":"5"}]}`)
		if err != nil {
			t.Fatalf("should not have error; got %s\n", err.Error())
		}

		if !strings.Contains(sqlStr, " OR ") {
			t.Errorf("sql should contain an OR clause; got %q", sqlStr)
		}

		if !strings.Contains(sqlStr, "user.name = ?") || !strings.Contains(sqlStr, "user.number ILIKE ?") {
			t.Errorf("sql should contain both conditions; got %q", sqlStr)
		}

		if len(args) != 2 || args[0] != "foo" || args[1] != "%5%" {
			t.Errorf("unexpected args; got %v", args)
		}
	})

	t.Run("nested and within or", func(t *testing.T) {
		sqlStr, args, err := buildSQL(t, `{"logic":"or","items":[`+
			`{"field":"user.name","operator":"eq","value":"foo"},`+
			`{"logic":"and","items":[`+
			`{"field":"phone.number","operator":"contains","value":"5"},`+
			`{"field":"user.name","operator":"isnotempty","value":""}]}]}`)
		if err != nil {
			t.Fatalf("should not have error; got %s\n", err.Error())
		}

		if !strings.Contains(sqlStr, " OR ") || !strings.Contains(sqlStr, " AND ") {
			t.Errorf("sql should contain both OR and AND clauses; got %q", sqlStr)
		}

		if !strings.Contains(sqlStr, "user.number ILIKE ?") || !strings.Contains(sqlStr, "user.name <> ?") {
			t.Errorf("sql should contain nested conditions; got %q", sqlStr)
		}

		if len(args) != 3 || args[0] != "foo" || args[1] != "%5%" || args[2] != "" {
			t.Errorf("unexpected args; got %v", args)
		}
	})

	t.Run("legacy flat array still ands", func(t *testing.T) {
		sqlStr, args, err := buildSQL(t, `[`+
			`{"field":"user.name","operator":"eq","value":"foo"},`+
			`{"field":"phone.number","operator":"contains","value":"5"}]`)
		if err != nil {
			t.Fatalf("should not have error; got %s\n", err.Error())
		}

		if strings.Contains(sqlStr, " OR ") {
			t.Errorf("sql should not contain an OR clause; got %q", sqlStr)
		}

		if !strings.Contains(sqlStr, " AND ") {
			t.Errorf("sql should contain an AND clause; got %q", sqlStr)
		}

		if len(args) != 2 || args[0] != "foo" || args[1] != "%5%" {
			t.Errorf("unexpected args; got %v", args)
		}
	})

	t.Run("invalid logic", func(t *testing.T) {
		_, _, err := buildSQL(t, `{"logic":"xor","items":[`+
			`{"field":"user.name","operator":"eq","value":"foo"},`+
			`{"field":"user.name","operator":"eq","value":"bar"}]}`)
		if err == nil {
			t.Fatal("should have error")
		}

		if !strings.Contains(err.Error(), "invalid logic") {
			t.Errorf("error should be '%s'; got '%s'", "invalid logic", err.Error())
		}
	})

	t.Run("invalid field within nested group", func(t *testing.T) {
		_, _, err := buildSQL(t, `{"logic":"or","items":[`+
			`{"logic":"and","items":[`+
			`{"field":"invalid","operator":"eq","value":"x"}]}]}`)
		if err == nil {
			t.Fatal("should have error")
		}

		if !strings.Contains(err.Error(), "invalid field") {
			t.Errorf("error should be '%s'; got '%s'", "invalid field", err.Error())
		}
	})

	t.Run("invalid operator within nested group", func(t *testing.T) {
		_, _, err := buildSQL(t, `{"logic":"or","items":[`+
			`{"logic":"and","items":[`+
			`{"field":"user.name","operator":"bad","value":"x"}]}]}`)
		if err == nil {
			t.Fatal("should have error")
		}

		if !strings.Contains(err.Error(), "invalid operator") {
			t.Errorf("error should be '%s'; got '%s'", "invalid operator", err.Error())
		}
	})

	t.Run("not filterable within group", func(t *testing.T) {
		_, _, err := buildSQL(t, `{"logic":"or","items":[`+
			`{"field":"user.id","operator":"eq","value":"x"}]}`)
		if err == nil {
			t.Fatal("should have error")
		}

		if !strings.Contains(err.Error(), "can not be filtered") {
			t.Errorf("error should be '%s'; got '%s'", "can not be filtered", err.Error())
		}
	})
}
