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
