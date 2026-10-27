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
	table := DbTable{
		Fields: DbFields{
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err == nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err == nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err == nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err == nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err == nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err == nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err == nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err == nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err == nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
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

	if _, err = GetQueryBuilder(req, builder, table, cfg); err != nil {
		t.Errorf("should not have error; got %s\n", err.Error())
	}
}

func TestGetQueryBuilderFilterGroups(t *testing.T) {
	cfg := QueryConfig{
		FilterParam: "filters",
	}
	table := DbTable{
		Fields: DbFields{
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

		builder, err := GetQueryBuilder(newReq(filterJSON), newBuilder(), table, cfg)
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

func TestGetQueryBuilderSearch(t *testing.T) {
	var req *http.Request
	var err error
	var urlVals url.Values
	var builder sq.SelectBuilder
	var query string
	var args []any

	cfg := QueryConfig{
		FilterParam: "filters",
		OrderParam:  "sorts",
		LimitParam:  "take",
		OffsetParam: "skip",

		SearchParam:  "q",
		DefaultOrder: "worker.id",

		Limit:  100,
		OffSet: 10000,
	}

	// Deliberately declared out of alphabetical order: the built expression
	// must come out sorted by DBField so it matches the expression index.
	table := DbTable{
		Fields: DbFields{
			"worker.zip": FieldConfig{
				DBField:      "worker.zip",
				OperationCfg: OperationConfig{CanSearchBy: true},
			},
			"worker.name": FieldConfig{
				DBField: "worker.name",
				OperationCfg: OperationConfig{
					CanFilterBy: true,
					CanSortBy:   true,
					CanSearchBy: true,
				},
			},
			"worker.address1": FieldConfig{
				DBField:      "worker.address1",
				OperationCfg: OperationConfig{CanSearchBy: true},
			},
		},
		SearchIndex: "worker_idx_bm25",
	}

	wantExpr := "(coalesce(worker.address1::text, '') || ' ' || coalesce(worker.name::text, '')" +
		" || ' ' || coalesce(worker.zip::text, ''))"

	newBuilder := func() sq.SelectBuilder {
		return sq.Select("worker.id").From("worker")
	}

	// ----------------------------------------------------------------------------------
	// Active search with no explicit sort: one bound predicate over the sorted
	// expression, relevance ordering (scores are negative, never DESC) with
	// DefaultOrder as tiebreak.

	urlVals = url.Values{}
	urlVals.Add("q", `Worke(r) "Phone"`)

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder, err = GetQueryBuilder(req, newBuilder(), table, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if query, args, err = builder.ToSql(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(query, wantExpr+" @@ to_tsquery(?, ?)") {
		t.Errorf("missing search predicate; got: %s", query)
	}

	if !strings.Contains(query, "ORDER BY "+wantExpr+" <@> to_bm25query(?, ?), worker.id") {
		t.Errorf("missing relevance then default ordering; got: %s", query)
	}

	if strings.Contains(query, "DESC") {
		t.Errorf("bm25 scores are negative: ordering must not use DESC; got: %s", query)
	}

	wantArgs := []any{"english", "worke:* | r:* | phone:*", "worke:* | r:* | phone:*", "worker_idx_bm25"}
	if len(args) != len(wantArgs) {
		t.Fatalf("expected args %v; got %v", wantArgs, args)
	}

	for i, want := range wantArgs {
		if args[i] != want {
			t.Errorf("arg %d: expected %q; got %v", i, want, args[i])
		}
	}

	if strings.Contains(query, "worke:*") {
		t.Errorf("search tokens must not be inlined into SQL; got: %s", query)
	}

	// ----------------------------------------------------------------------------------
	// An explicit sort always wins over relevance, but the predicate stays.

	urlVals = url.Values{}
	urlVals.Add("q", "anything")
	urlVals.Add(cfg.OrderParam, `[{"dir":"asc","field":"worker.name"}]`)

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder, err = GetQueryBuilder(req, newBuilder(), table, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if query, _, err = builder.ToSql(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(query, "ORDER BY worker.name asc") {
		t.Errorf("explicit sort must be applied; got: %s", query)
	}

	if strings.Contains(query, "to_bm25query") {
		t.Errorf("explicit sort must win over relevance; got: %s", query)
	}

	if !strings.Contains(query, "@@ to_tsquery(?, ?)") {
		t.Errorf("predicate must survive an explicit sort; got: %s", query)
	}

	// ----------------------------------------------------------------------------------
	// Without a `q` the builder only gets the plain DefaultOrder — byte-for-byte
	// the pre-search behaviour.

	req = httptest.NewRequest(http.MethodGet, "/url", nil)
	builder, err = GetQueryBuilder(req, newBuilder(), table, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if query, args, err = builder.ToSql(); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(query, "to_tsquery") || strings.Contains(query, "to_bm25query") {
		t.Errorf("expected no search SQL without a q; got: %s", query)
	}

	if !strings.Contains(query, "ORDER BY worker.id") {
		t.Errorf("expected plain default ordering; got: %s", query)
	}

	if len(args) != 0 {
		t.Errorf("expected no args; got %v", args)
	}

	// ----------------------------------------------------------------------------------
	// A config without OrderParam (count queries, grouped subqueries) gets the
	// predicate but never an ordering, so grouped builders stay valid.

	urlVals = url.Values{}
	urlVals.Add("q", "anything")

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder, err = GetQueryBuilder(
		req,
		newBuilder(),
		table,
		QueryConfig{SearchParam: "q"},
	)
	if err != nil {
		t.Fatal(err)
	}

	if query, _, err = builder.ToSql(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(query, "@@ to_tsquery(?, ?)") {
		t.Errorf("count-style config must still get the predicate; got: %s", query)
	}

	if strings.Contains(query, "ORDER BY") {
		t.Errorf("count-style config must not order; got: %s", query)
	}

	// ----------------------------------------------------------------------------------
	// Tables whose Fields mark nothing searchable stay untouched even with a
	// `q` present (an unwired table can never produce invalid SQL).

	urlVals = url.Values{}
	urlVals.Add("q", "anything")

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder, err = GetQueryBuilder(req, newBuilder(), DbTable{
		Fields: DbFields{
			"worker.name": FieldConfig{
				DBField:      "worker.name",
				OperationCfg: OperationConfig{CanFilterBy: true},
			},
		},
		SearchIndex: "worker_idx_bm25",
	}, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if query, _, err = builder.ToSql(); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(query, "to_tsquery") || strings.Contains(query, "to_bm25query") {
		t.Errorf("unsearchable fields must stay untouched; got: %s", query)
	}

	// ----------------------------------------------------------------------------------
	// Active search without SearchIndex is a configuration error and fails fast.

	urlVals = url.Values{}
	urlVals.Add("q", "anything")

	noIndexTable := table
	noIndexTable.SearchIndex = ""

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	if _, err = GetQueryBuilder(req, newBuilder(), noIndexTable, cfg); err == nil {
		t.Error("expected error for missing SearchIndex")
	} else if !strings.Contains(err.Error(), "SearchIndex") {
		t.Errorf("error should mention SearchIndex; got '%s'", err.Error())
	}

	// ----------------------------------------------------------------------------------
	// Token caps: at most maxSearchTokens tokens of maxSearchTokenLen runes,
	// rendered into one bound query string.

	urlVals = url.Values{}
	urlVals.Add("q", "a b c d e f g h i j k l m n o")

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder, err = GetQueryBuilder(req, newBuilder(), table, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if _, args, err = builder.ToSql(); err != nil {
		t.Fatal(err)
	}

	wantQuery := strings.Join([]string{
		"a:*", "b:*", "c:*", "d:*", "e:*",
		"f:*", "g:*", "h:*", "i:*", "j:*",
	}, " | ")

	if args[1] != wantQuery {
		t.Errorf("expected capped query %q; got %v", wantQuery, args[1])
	}

	urlVals = url.Values{}
	urlVals.Add("q", strings.Repeat("x", maxSearchTokenLen+20))

	req = httptest.NewRequest(http.MethodGet, "/url?"+urlVals.Encode(), nil)
	builder, err = GetQueryBuilder(req, newBuilder(), table, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if _, args, err = builder.ToSql(); err != nil {
		t.Fatal(err)
	}

	if args[1] != strings.Repeat("x", maxSearchTokenLen)+":*" {
		t.Errorf("expected %d-rune token; got %v", maxSearchTokenLen, args[1])
	}
}
