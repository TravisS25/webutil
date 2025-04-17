package webutil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"reflect"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/pkg/errors"
)

var (
	// When set, the final query of query builder function being executed will print to stdout
	// This is to help visualize what the query builder is sending to database to troubleshoot
	// any queries
	DebugPrintQueryOutput = false
)

//////////////////////////////////////////////////////////////////
//----------------------- INTERFACES -------------------------
//////////////////////////////////////////////////////////////////

type ColScanner interface {
	Columns() ([]string, error)
	Scan(dest ...any) error
	Err() error
}

type Database interface {
	qrm.DB
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

//////////////////////////////////////////////////////////////////
//-------------------------- TYPES --------------------------
//////////////////////////////////////////////////////////////////

// DbFields is type used in various querying functions
// The key value should be fields that will be sent in a
// url query and the value is the configuration of each field
// allowing a user the flexibility to determine if a certain field
// can be filtered, sortable, and/or groupable and will return
// appropriate error if settings don't match
type DbFields map[string]FieldConfig

//////////////////////////////////////////////////////////////////
//-------------------------- STRUCTS --------------------------
//////////////////////////////////////////////////////////////////

type QueryConfig struct {
	FilterParam string
	OrderParam  string
	LimitParam  string
	OffsetParam string

	Limit  uint64
	OffSet uint64

	CanMultiColumnOrder bool
	CanMultiColumnGroup bool
}

type DataInputParams struct {
	QueryCfg   QueryConfig
	CustomFunc func(map[string]any) error
}

type CountInputParams struct {
	QueryCfg QueryConfig
}

type SelectItem struct {
	Value string `json:"value" mapstructure:"value" db:"value" alias:"select.value"`
	Text  string `json:"text" mapstructure:"text" db:"text" alias:"select.text"`
}

// FilteredResults is struct used for dynamically filtered results
type FilteredResults struct {
	Data  any `json:"data"`
	Total int `json:"total"`
}

type QueryBuilderError struct {
	errorMsg string
}

func (q QueryBuilderError) Error() string {
	return q.errorMsg
}

// OperationConfig is used in conjunction with FieldConfig{}
// to determine if the field associated can perform certain
// sql actions
type OperationConfig struct {
	// CanFilterBy determines whether field can have filters applied
	CanFilterBy bool

	// CanSortBy determines whether field can be sorted
	CanSortBy bool

	// CanGroupBy determines whether field can be grouped
	CanGroupBy bool
}

// ParamConfig is for extracting expected query params from url
// to be passed to the server
type ParamConfig struct {
	// Filter is for query param that will be applied
	// to "where" clause of query
	Filter *string

	// Sort is for query param that will be applied
	// to "order by" clause of query
	Sort *string

	// Take is for query param that will be applied
	// to "limit" clause of query
	Take *string

	// Skip is for query param that will be applied
	// to "offset" clause of query
	Skip *string

	// Group is for query param that will be applied
	// to "group by" clause of query
	Group *string
}

// FieldConfig is meant to be a per database field config
// to determine if a user can perform a certain sql action
// and if user tries to perform action not allowed, throw error
type FieldConfig struct {
	// DBField should be the name of the database field
	// to apply configurations to
	DBField string

	// ValueOverride should be used to override and return a different value for field
	ValueOverride func(value any) (any, error)

	// OperationCfg is config to set to determine which sql
	// operations can be performed on DBField
	OperationCfg OperationConfig
}

// Filter is the filter config struct for server side filtering
type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// Sort is the sort config struct for server side sorting
type Sort struct {
	Dir   string `json:"dir"`
	Field string `json:"field"`
}

type Order = Sort

// Group is the group config struct for server side grouping
type Group struct {
	Field string `json:"field"`
}

// LimitOffset is config struct to set limit and offset of a query
type LimitOffset struct {
	Take int `json:"take"`
	Skip int `json:"skip"`
}

////////////////////////////////////////////////////////////
// --------------------- FUNCTIONS ---------------------
////////////////////////////////////////////////////////////

func In(query string, args ...any) (string, []any, error) {
	// argMeta stores reflect.Value and length for slices and
	// the value itself for non-slice arguments
	type argMeta struct {
		v      reflect.Value
		i      any
		length int
	}

	var flatArgsCount int
	var anySlices bool

	var stackMeta [32]argMeta

	var meta []argMeta
	if len(args) <= len(stackMeta) {
		meta = stackMeta[:len(args)]
	} else {
		meta = make([]argMeta, len(args))
	}

	for i, arg := range args {
		if a, ok := arg.(driver.Valuer); ok {
			var err error
			arg, err = a.Value()
			if err != nil {
				return "", nil, err
			}
		}

		if v, ok := asSliceForIn(arg); ok {
			meta[i].length = v.Len()
			meta[i].v = v

			anySlices = true
			flatArgsCount += meta[i].length

			if meta[i].length == 0 {
				return "", nil, errors.New("empty slice passed to 'in' query")
			}
		} else {
			meta[i].i = arg
			flatArgsCount++
		}
	}

	// don't do any parsing if there aren't any slices;  note that this means
	// some errors that we might have caught below will not be returned.
	if !anySlices {
		return query, args, nil
	}

	newArgs := make([]any, 0, flatArgsCount)

	var buf strings.Builder
	buf.Grow(len(query) + len(", ?")*flatArgsCount)

	var arg, offset int

	for i := strings.IndexByte(query[offset:], '?'); i != -1; i = strings.IndexByte(query[offset:], '?') {
		if arg >= len(meta) {
			// if an argument wasn't passed, lets return an error;  this is
			// not actually how database/sql Exec/Query works, but since we are
			// creating an argument list programmatically, we want to be able
			// to catch these programmer errors earlier.
			return "", nil, errors.New("number of bindVars exceeds arguments")
		}

		argMeta := meta[arg]
		arg++

		// not a slice, continue.
		// our questionmark will either be written before the next expansion
		// of a slice or after the loop when writing the rest of the query
		if argMeta.length == 0 {
			offset = offset + i + 1
			newArgs = append(newArgs, argMeta.i)
			continue
		}

		// write everything up to and including our ? character
		buf.WriteString(query[:offset+i+1])

		for si := 1; si < argMeta.length; si++ {
			buf.WriteString(", ?")
		}

		newArgs = appendReflectSlice(newArgs, argMeta.v, argMeta.length)

		// slice the query and reset the offset. this avoids some bookkeeping for
		// the write after the loop
		query = query[offset+i+1:]
		offset = 0
	}

	buf.WriteString(query)

	if arg < len(meta) {
		return "", nil, errors.New("number of bindVars less than number arguments")
	}

	return buf.String(), newArgs, nil
}

func Rebind(bindType int, query string) string {
	switch bindType {
	case QUESTION_SQL_BIND_VAR, UNKNOWN_SQL_BIND_VAR:
		return query
	}

	// Add space enough for 10 params before we have to allocate
	rqb := make([]byte, 0, len(query)+10)

	var i, j int

	for i = strings.Index(query, "?"); i != -1; i = strings.Index(query, "?") {
		rqb = append(rqb, query[:i]...)

		switch bindType {
		case DOLLAR_SQL_BIND_VAR:
			rqb = append(rqb, '$')
		case NAMED_SQL_BIND_VAR:
			rqb = append(rqb, ':', 'a', 'r', 'g')
		case AT_SQL_BIND_VAR:
			rqb = append(rqb, '@', 'p')
		}

		j++
		rqb = strconv.AppendInt(rqb, int64(j), 10)

		query = query[i+1:]
	}

	return string(append(rqb, query...))
}

func InQueryRebind(bindType int, query string, args ...any) (string, []any, error) {
	query, args, err := In(query, args...)
	if err != nil {
		return query, args, err
	}

	query = Rebind(bindType, query)
	return query, args, nil
}

func MapScanner(r ColScanner, dest map[string]any) error {
	columns, values, err := scanColVals(r)

	if err != nil {
		return err
	}

	// getInnerMap takes in colMap and colWords and gets the inner most map and returns it
	getInnerMap := func(colMap map[string]any, colWords []string) map[string]any {
		if len(colWords) == 0 {
			return nil
		}

		var innerMap map[string]any

		for i := 0; i < len(colWords); i++ {
			if i == 0 {
				innerMap = colMap[colWords[i]].(map[string]any)
			} else {
				innerMap = innerMap[colWords[i]].(map[string]any)
			}
		}

		return innerMap
	}

	for i := range columns {
		colWords := strings.Split(columns[i], ".")

		for idx := range colWords {
			if idx == len(colWords)-1 {
				innerMap := getInnerMap(dest, colWords[:idx])

				if innerMap == nil {
					dest[colWords[idx]] = *(values[i].(*any))
				} else {
					innerMap[colWords[idx]] = *(values[i].(*any))
				}
			} else {
				innerMap := getInnerMap(dest, colWords[:idx])

				if innerMap == nil {
					if _, ok := dest[colWords[idx]]; !ok {
						dest[colWords[idx]] = make(map[string]any)
					}
				} else {
					if _, ok := innerMap[colWords[idx]]; !ok {
						innerMap[colWords[idx]] = make(map[string]any)
					}
				}
			}
		}
	}

	return r.Err()
}

////////////////////////////////////////////////////////////
// --------------------- QUERY FUNCTIONS ----------------
////////////////////////////////////////////////////////////

func QueryDB(
	ctx context.Context,
	db qrm.Queryable,
	bindType int,
	query string,
	args []any,
	rowUpdate func(row any) error,
	destPtr any,
) error {
	var err error

	newQuery, newArgs, err := InQueryRebind(bindType, query, args...)
	if err != nil {
		return fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), query, args)
	}

	destVal := reflect.ValueOf(destPtr)

	rows, err := db.QueryContext(ctx, newQuery, newArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("webutil: error getting number of columns: %w", err)
	}

	var box any
	isArr := false

	if destVal.Elem().Kind() == reflect.Slice {
		box = make([]any, 0)
		isArr = true
	}

	hasRow := false

	for rows.Next() {
		var row any
		hasRow = true

		if len(cols) > 1 {
			val := make(map[string]any)

			if err = MapScanner(rows, val); err != nil {
				return errors.WithStack(err)
			}

			if rowUpdate != nil {
				if err = rowUpdate(&val); err != nil {
					return err
				}
			}

			row = val
		} else {
			if err = rows.Scan(&row); err != nil {
				return errors.WithStack(err)
			}

			if rowUpdate != nil {
				if err = rowUpdate(&row); err != nil {
					return err
				}
			}
		}

		if isArr {
			box = append(box.([]any), row)
		} else {
			box = row
			break
		}
	}

	if !isArr && !hasRow {
		return sql.ErrNoRows
	}

	jsonBytes, err := json.Marshal(box)
	if err != nil {
		return errors.WithStack(err)
	}

	if err = json.Unmarshal(jsonBytes, destPtr); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

////////////////////////////////////////////////////////////
// --------------------- BUILDER FUNCTIONS ---------------
////////////////////////////////////////////////////////////

func QueryDataResult(
	ctx context.Context,
	req *http.Request,
	builder sq.SelectBuilder,
	dbFields DbFields,
	db qrm.Queryable,
	bindVar int,
	dest any,
	params DataInputParams,
) error {
	data, ok := dest.(*[]map[string]any)
	if !ok {
		return fmt.Errorf("dest parameter must pointer to []map[string]any; got %s", reflect.TypeOf(dest))
	}

	var err error

	if builder, err = GetQueryBuilder(
		req,
		builder,
		dbFields,
		params.QueryCfg,
	); err != nil {
		return errors.WithStack(err)
	}

	rows, err := getRowsFromBuilder(ctx, builder, db, bindVar)
	if err != nil {
		return err
	}

	for rows.Next() {
		row := map[string]any{}

		if err = MapScanner(rows, row); err != nil {
			return errors.WithStack(err)
		}

		if params.CustomFunc != nil {
			if err = params.CustomFunc(row); err != nil {
				return errors.WithStack(err)
			}
		}

		*data = append(*data, row)
	}

	return nil
}

func QueryCountResult(
	ctx context.Context,
	req *http.Request,
	builder sq.SelectBuilder,
	dbFields DbFields,
	db qrm.Queryable,
	bindVar int,
	dest any,
	params CountInputParams,
) error {
	count, ok := dest.(*uint64)
	if !ok {
		return fmt.Errorf("dest parameter must pointer to uint64; got %s", reflect.TypeOf(dest))
	}

	var err error

	if builder, err = GetQueryBuilder(
		req,
		builder,
		dbFields,
		params.QueryCfg,
	); err != nil {
		return errors.WithStack(err)
	}

	rows, err := getRowsFromBuilder(ctx, builder, db, bindVar)
	if err != nil {
		return err
	}

	for rows.Next() {
		if err = rows.Scan(count); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func QueryDataAndCountResults(
	ctx context.Context,
	req *http.Request,
	dataBuilder sq.SelectBuilder,
	countBuilder sq.SelectBuilder,
	dbFields DbFields,
	db qrm.Queryable,
	bindVar int,
	dataDest any,
	countDest *uint64,
	dataParams DataInputParams,
	countParams CountInputParams,
) error {
	err := QueryDataResult(
		ctx,
		req,
		dataBuilder,
		dbFields,
		db,
		bindVar,
		dataDest,
		dataParams,
	)
	if err != nil {
		return err
	}

	err = QueryCountResult(
		ctx,
		req,
		countBuilder,
		dbFields,
		db,
		bindVar,
		countDest,
		countParams,
	)
	if err != nil {
		return err
	}

	return nil
}

func GetInnerBuilderResults(
	r *http.Request,
	builder sq.SelectBuilder,
	dbFields DbFields,
	defaultOrderBy string,
	queryCfg QueryConfig,
) (string, []any, error) {
	innerDataBuilder, err := GetQueryBuilder(
		r,
		builder,
		dbFields,
		queryCfg,
	)
	if err != nil {
		return "", nil, errors.WithStack(err)
	}

	if defaultOrderBy != "" && r.FormValue(queryCfg.OrderParam) == "" {
		innerDataBuilder = innerDataBuilder.OrderBy(defaultOrderBy)
	}

	return innerDataBuilder.ToSql()
}

func GetQueryBuilder(
	r *http.Request,
	builder sq.SelectBuilder,
	dbFields DbFields,
	cfg QueryConfig,
) (sq.SelectBuilder, error) {
	var err error
	var ok bool

	filterParam := r.FormValue(cfg.FilterParam)
	orderParam := r.FormValue(cfg.OrderParam)
	limitParam := r.FormValue(cfg.LimitParam)
	offsetParam := r.FormValue(cfg.OffsetParam)

	if filterParam != "" {
		var filters []Filter

		if err = json.Unmarshal([]byte(filterParam), &filters); err != nil {
			return sq.SelectBuilder{}, errors.WithStack(QueryBuilderError{
				errorMsg: "invalid filter parameter"},
			)
		}

		for _, filter := range filters {
			var dbField FieldConfig
			invalidFilterValue := ""

			if dbField, ok = dbFields[filter.Field]; !ok {
				return sq.SelectBuilder{}, errors.WithStack(
					QueryBuilderError{errorMsg: fmt.Sprintf("invalid field %q for filter parameter", filter.Field)},
				)
			}

			if !dbField.OperationCfg.CanFilterBy {
				return sq.SelectBuilder{}, errors.WithStack(
					QueryBuilderError{fmt.Sprintf("field %q can not be filtered", filter.Field)},
				)
			}

			if filter.Value == nil {
				if filter.Operator != "isnull" && filter.Operator != "isnotnull" {
					return sq.SelectBuilder{}, QueryBuilderError{
						errorMsg: fmt.Sprintf("field %q does not contain value", filter.Field),
					}
				}
			}

			fieldValue := filter.Value

			if dbField.ValueOverride != nil {
				if fieldValue, err = dbField.ValueOverride(fieldValue); err != nil {
					return sq.SelectBuilder{}, errors.WithStack(QueryBuilderError{errorMsg: fmt.Sprintf("invalid value %q for field %q", fieldValue, filter.Field)})
				}
			}

			switch filter.Operator {
			case "eq":
				builder = builder.Where(sq.Eq{
					dbField.DBField: fieldValue,
				})
			case "neq":
				builder = builder.Where(sq.NotEq{
					dbField.DBField: fieldValue,
				})
			case "startswith":
				builder = builder.Where(sq.ILike{
					dbField.DBField: fmt.Sprintf("%v%%", fieldValue),
				})
			case "endswith":
				builder = builder.Where(sq.ILike{
					dbField.DBField: fmt.Sprintf("%%%v", fieldValue),
				})
			case "contains":
				builder = builder.Where(sq.ILike{
					dbField.DBField: fmt.Sprintf("%%%v%%", fieldValue),
				})
			case "doesnotcontain":
				builder = builder.Where(sq.NotILike{
					dbField.DBField: fmt.Sprintf("%%%v%%", fieldValue),
				})
			case "isnull":
				builder = builder.Where(sq.Eq{
					dbField.DBField: nil,
				})
			case "isnotnull":
				builder = builder.Where(sq.NotEq{
					dbField.DBField: nil,
				})
			case "isempty":
				builder = builder.Where(sq.Eq{
					dbField.DBField: "",
				})
			case "isnotempty":
				builder = builder.Where(sq.NotEq{
					dbField.DBField: "",
				})
			case "lt":
				builder = builder.Where(sq.Lt{
					dbField.DBField: fmt.Sprintf("%v", fieldValue),
				})
			case "lte":
				builder = builder.Where(sq.LtOrEq{
					dbField.DBField: fmt.Sprintf("%v", fieldValue),
				})
			case "gt":
				builder = builder.Where(sq.Gt{
					dbField.DBField: fmt.Sprintf("%v", fieldValue),
				})
			case "gte":
				builder = builder.Where(sq.GtOrEq{
					dbField.DBField: fmt.Sprintf("%v", fieldValue),
				})
			default:
				return sq.SelectBuilder{}, errors.WithStack(
					QueryBuilderError{errorMsg: fmt.Sprintf("invalid operator for field %q", filter.Field)},
				)
			}

			if invalidFilterValue != "" {
				return sq.SelectBuilder{}, errors.WithStack(
					QueryBuilderError{errorMsg: fmt.Sprintf("invalid filter value for field %q", invalidFilterValue)},
				)
			}
		}
	}

	if orderParam != "" {
		var sorts []Order

		if err = json.Unmarshal([]byte(orderParam), &sorts); err != nil {
			return sq.SelectBuilder{}, errors.WithStack(
				QueryBuilderError{errorMsg: "invalid order parameter"},
			)
		}

		for _, sort := range sorts {
			var dbField FieldConfig

			if sort.Dir != "asc" && sort.Dir != "desc" {
				return sq.SelectBuilder{}, errors.WithStack(
					QueryBuilderError{errorMsg: fmt.Sprintf("invalid sort dir for field %q", sort.Field)},
				)
			}

			if dbField, ok = dbFields[sort.Field]; !ok {
				return sq.SelectBuilder{}, errors.WithStack(
					QueryBuilderError{errorMsg: fmt.Sprintf("invalid field %q for order parameter", sort.Field)},
				)
			}

			if !dbField.OperationCfg.CanSortBy {
				return sq.SelectBuilder{}, errors.WithStack(
					QueryBuilderError{errorMsg: fmt.Sprintf("field %q can not be ordered", sort.Field)},
				)
			}

			builder = builder.OrderByClause(dbField.DBField + " " + sort.Dir)

			if !cfg.CanMultiColumnOrder {
				break
			}
		}
	}

	if limitParam != "" {
		var limit uint64

		if limit, err = strconv.ParseUint(
			limitParam,
			INT_BASE,
			INT_BIT_SIZE,
		); err != nil {
			return sq.SelectBuilder{}, errors.WithStack(
				QueryBuilderError{errorMsg: "invalid limit"},
			)
		}

		if cfg.Limit > 0 && limit > cfg.Limit {
			limit = cfg.Limit
		}

		builder = builder.Limit(limit)
	} else if cfg.Limit > 0 {
		builder = builder.Limit(cfg.Limit)
	}

	if offsetParam != "" {
		var offset uint64

		if offset, err = strconv.ParseUint(
			offsetParam,
			INT_BASE,
			INT_BIT_SIZE,
		); err != nil {
			return sq.SelectBuilder{}, errors.WithStack(
				QueryBuilderError{errorMsg: "invalid offset"},
			)
		}

		if cfg.OffSet > 0 && offset > cfg.OffSet {
			offset = cfg.OffSet
		}

		builder = builder.Offset(offset)
	}

	return builder, nil
}

func scanColVals(r ColScanner) ([]string, []any, error) {
	// ignore r.started, since we needn't use reflect for anything.
	columns, err := r.Columns()
	if err != nil {
		return nil, nil, err
	}

	values := make([]any, len(columns))
	for i := range values {
		values[i] = new(any)
	}

	err = r.Scan(values...)
	return columns, values, err
}

func appendReflectSlice(args []any, v reflect.Value, vlen int) []any {
	switch val := v.Interface().(type) {
	case []any:
		args = append(args, val...)
	case []int:
		for i := range val {
			args = append(args, val[i])
		}
	case []string:
		for i := range val {
			args = append(args, val[i])
		}
	default:
		for si := 0; si < vlen; si++ {
			args = append(args, v.Index(si).Interface())
		}
	}

	return args
}

func asSliceForIn(i any) (v reflect.Value, ok bool) {
	if i == nil {
		return reflect.Value{}, false
	}

	v = reflect.ValueOf(i)
	t := deref(v.Type())

	// Only expand slices
	if t.Kind() != reflect.Slice {
		return reflect.Value{}, false
	}

	// []byte is a driver.Value type so it should not be expanded
	if t == reflect.TypeOf([]byte{}) {
		return reflect.Value{}, false

	}

	return v, true
}

// Deref is Indirect for reflect.Types
func deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func getRowsFromBuilder(ctx context.Context, builder sq.SelectBuilder, db qrm.Queryable, bindVar int) (*sql.Rows, error) {
	var query string
	var err error
	var args []any

	if query, args, err = builder.ToSql(); err != nil {
		return nil, errors.WithStack(err)
	}

	resQuery := query
	resArgs := args

	if DebugPrintQueryOutput {
		fmt.Printf("webutil: debug query: %s\n", resQuery)
		fmt.Printf("webutil: debug args: %+v\n", resArgs...)
	}

	if query, args, err = InQueryRebind(bindVar, query, args...); err != nil {
		return nil, errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), resQuery, resArgs))
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), resQuery, resArgs))
	}

	return rows, nil
}
