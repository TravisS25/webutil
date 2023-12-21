package webutil

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	"reflect"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

//////////////////////////////////////////////////////////////////
//------------------------ STRING CONSTS -----------------------
//////////////////////////////////////////////////////////////////

const (
	// Select string for queries
	Select = "select "
)

var (
	validFilterTypes = map[string]struct{}{"string": {}, "int": {}, "float": {}, "float64": {}, "int64": {}}
)

//////////////////////////////////////////////////////////////////
//----------------------- CUSTOM ERRORS -------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrInvalidSort is error returned if client tries
	// to pass filter parameter that is not sortable
	ErrInvalidSort = errors.New("webutil: invalid sort")

	// ErrInvalidArray is error returned if client tries
	// to pass array parameter that is invalid array type
	ErrInvalidArray = errors.New("webutil: invalid array for field")

	// ErrInvalidValue is error returned if client tries
	// to pass filter parameter that had invalid field
	// value for certain field
	ErrInvalidValue = errors.New("webutil: invalid field value")
)

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

type InnerBuilderResult struct {
	DataQuery  string
	CountQuery string
	Args       []interface{}
}

type QueryBuilderConfig struct {
	FilterParam string
	OrderParam  string
	LimitParam  string
	OffsetParam string

	Limit  uint64
	OffSet uint64

	CanMultiColumnOrder bool
	CanMultiColumnGroup bool

	IsCountQuery bool
}

type QueryBuilderErrorResponse struct {
	QueryBuilderErrorStatus int
	DatabaseErrorStatus     int
	DatabaseErrorResponse   []byte
}

// SelectItem is struct used in conjunction with primeng's library api
// type SelectItem struct {
// 	Value interface{} `json:"value" db:"value" alias:"value"`
// 	Text  string      `json:"text" db:"text" alias:"text"`
// }

type SelectItem struct {
	Value string `json:"value" mapstructure:"value" db:"value" alias:"select.value"`
	Text  string `json:"text" mapstructure:"text" db:"text" alias:"select.text"`
}

// FilteredResults is struct used for dynamically filtered results
type FilteredResults struct {
	Data  interface{} `json:"data"`
	Total int         `json:"total"`
}

type QueryBuilderError struct {
	errorMsg string
}

func (q *QueryBuilderError) Error() string {
	return q.errorMsg
}

//////////////////////////////////////////////////////////////////
//-------------------------- STRUCTS --------------------------
//////////////////////////////////////////////////////////////////

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

	// FieldTypeValidator is custom function that allows user
	// to determine if db field value sent by user is valid
	// and if not, return an error
	FieldTypeValidator func(value interface{}) error

	// OperationCfg is config to set to determine which sql
	// operations can be performed on DBField
	OperationCfg OperationConfig
}

// Filter is the filter config struct for server side filtering
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
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

// CountSelect take column string and applies count select
func CountSelect(column string) string {
	return fmt.Sprintf("count(%s) as total", column)
}

func In(query string, args ...interface{}) (string, []interface{}, error) {
	// argMeta stores reflect.Value and length for slices and
	// the value itself for non-slice arguments
	type argMeta struct {
		v      reflect.Value
		i      interface{}
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

	newArgs := make([]interface{}, 0, flatArgsCount)

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
	case QUESTION, UNKNOWN:
		return query
	}

	// Add space enough for 10 params before we have to allocate
	rqb := make([]byte, 0, len(query)+10)

	var i, j int

	for i = strings.Index(query, "?"); i != -1; i = strings.Index(query, "?") {
		rqb = append(rqb, query[:i]...)

		switch bindType {
		case DOLLAR:
			rqb = append(rqb, '$')
		case NAMED:
			rqb = append(rqb, ':', 'a', 'r', 'g')
		case AT:
			rqb = append(rqb, '@', 'p')
		}

		j++
		rqb = strconv.AppendInt(rqb, int64(j), 10)

		query = query[i+1:]
	}

	return string(append(rqb, query...))
}

// InQueryRebind is wrapper function for combining sqlx.In() and sqlx.Rebind()
// to handle passing database bind type along with handling errors
func InQueryRebind(bindType int, query string, args ...interface{}) (string, []interface{}, error) {
	query, args, err := sqlx.In(query, args...)

	if err != nil {
		return query, nil, err
	}

	query = sqlx.Rebind(bindType, query)
	return query, args, nil
}

// QuerySelectItems is utility function for querying against a table and returning
// a list of SelectItem structs to be used with primeng's library components
func QuerySelectItems(db SqlxDB, bindVar int, query string, args ...interface{}) ([]SelectItem, error) {
	var err error
	var items []SelectItem

	if err = db.SelectRebind(&items, bindVar, query, args...); err != nil {
		return nil, err
	}

	return items, nil
}

// QuerySingleColumn is utility function used to query for single column
// Will return error if length of *sql.Rows#Columns does not return 1
func QuerySingleColumn(db Querier, bindVar int, query string, args ...interface{}) ([]interface{}, error) {
	var err error

	if query, args, err = InQueryRebind(bindVar, query, args...); err != nil {
		return nil, err
	}

	items := make([]interface{}, 0)
	rows, err := db.Queryx(query, args...)

	if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()

	if err != nil {
		return nil, err
	}

	if len(cols) != 1 {
		return nil, fmt.Errorf("webutil: query should only return one column")
	}

	for rows.Next() {
		var item interface{}

		if err = rows.Scan(&item); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func GetQueryBuilder(
	r *http.Request,
	dbFields DbFields,
	builder sq.SelectBuilder,
	cfg QueryBuilderConfig,
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
			return sq.SelectBuilder{}, errors.WithStack(&QueryBuilderError{
				errorMsg: "invalid filter parameter"},
			)
		}

		for _, filter := range filters {
			var dbField FieldConfig
			var fieldType string
			invalidFilterValue := ""

			if dbField, ok = dbFields[filter.Field]; !ok {
				return sq.SelectBuilder{}, errors.WithStack(
					&QueryBuilderError{errorMsg: fmt.Sprintf("invalid field '%s' for filter parameter", filter.Field)},
				)
			}

			if !dbField.OperationCfg.CanFilterBy {
				return sq.SelectBuilder{}, errors.WithStack(
					&QueryBuilderError{fmt.Sprintf("field '%s' can not be filtered", filter.Field)},
				)
			}

			if filter.Value == nil {
				if filter.Operator != "isnull" && filter.Operator != "isnotnull" {
					return sq.SelectBuilder{}, &QueryBuilderError{
						errorMsg: fmt.Sprintf("field '%s' does not contain value", filter.Field),
					}
				}
			} else {
				fieldType = reflect.TypeOf(filter.Value).String()
			}

			switch filter.Operator {
			case "eq":
				builder = builder.Where(sq.Eq{
					dbField.DBField: filter.Value,
				})
			case "neq":
				builder = builder.Where(sq.NotEq{
					dbField.DBField: filter.Value,
				})
			case "startswith":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.ILike{
						dbField.DBField: fmt.Sprintf("%v%%", filter.Value),
					})
				}
			case "endswith":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.ILike{
						dbField.DBField: fmt.Sprintf("%%%v", filter.Value),
					})
				}
			case "contains":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.ILike{
						dbField.DBField: fmt.Sprintf("%%%v%%", filter.Value),
					})
				}
			case "doesnotcontain":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.NotILike{
						dbField.DBField: fmt.Sprintf("%%%v%%", filter.Value),
					})
				}
			case "isnull":
				builder = builder.Where(sq.Eq{
					dbField.DBField: nil,
				})
			case "isnotnull":
				builder = builder.Where(sq.NotEq{
					dbField.DBField: nil,
				})
			case "isempty":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.Eq{
						dbField.DBField: "",
					})
				}
			case "isnotempty":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.NotEq{
						dbField.DBField: "",
					})
				}
			case "lt":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.Lt{
						dbField.DBField: fmt.Sprintf("%v", filter.Value),
					})
				}
			case "lte":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.LtOrEq{
						dbField.DBField: fmt.Sprintf("%v", filter.Value),
					})
				}
			case "gt":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.Gt{
						dbField.DBField: fmt.Sprintf("%v", filter.Value),
					})
				}
			case "gte":
				if _, ok = validFilterTypes[fieldType]; !ok {
					invalidFilterValue = filter.Field
				} else {
					builder = builder.Where(sq.GtOrEq{
						dbField.DBField: fmt.Sprintf("%v", filter.Value),
					})
				}
			default:
				return sq.SelectBuilder{}, errors.WithStack(
					&QueryBuilderError{errorMsg: fmt.Sprintf("invalid operator for field '%s'", filter.Field)},
				)
			}

			if invalidFilterValue != "" {
				return sq.SelectBuilder{}, errors.WithStack(
					&QueryBuilderError{errorMsg: fmt.Sprintf("invalid filter value for field '%s'", invalidFilterValue)},
				)
			}
		}
	}

	if orderParam != "" {
		var sorts []Order

		if err = json.Unmarshal([]byte(orderParam), &sorts); err != nil {
			return sq.SelectBuilder{}, errors.WithStack(
				&QueryBuilderError{errorMsg: "invalid order parameter"},
			)
		}

		for _, sort := range sorts {
			var dbField FieldConfig

			if sort.Dir != "asc" && sort.Dir != "desc" {
				return sq.SelectBuilder{}, errors.WithStack(
					&QueryBuilderError{errorMsg: fmt.Sprintf("invalid sort dir for field '%s'", sort.Field)},
				)
			}

			if dbField, ok = dbFields[sort.Field]; !ok {
				return sq.SelectBuilder{}, errors.WithStack(
					&QueryBuilderError{errorMsg: fmt.Sprintf("invalid field '%s' for order parameter", sort.Field)},
				)
			}

			if !dbField.OperationCfg.CanSortBy {
				return sq.SelectBuilder{}, errors.WithStack(
					&QueryBuilderError{errorMsg: fmt.Sprintf("field '%s' can not be ordered", sort.Field)},
				)
			}

			builder = builder.OrderByClause(dbField.DBField + " " + sort.Dir)

			if !cfg.CanMultiColumnOrder {
				break
			}
		}
	}

	if limitParam != "" || cfg.IsCountQuery {
		var limit uint64

		if cfg.IsCountQuery {
			limit = cfg.Limit
		} else {
			if limit, err = strconv.ParseUint(
				limitParam,
				IntBase,
				IntBitSize,
			); err != nil {
				return sq.SelectBuilder{}, errors.WithStack(
					&QueryBuilderError{errorMsg: "invalid limit"},
				)
			}

			if cfg.Limit > 0 && limit > cfg.Limit {
				limit = cfg.Limit
			}
		}

		builder = builder.Limit(limit)
	}

	if offsetParam != "" {
		var offset uint64

		if offset, err = strconv.ParseUint(
			offsetParam,
			IntBase,
			IntBitSize,
		); err != nil {
			return sq.SelectBuilder{}, errors.WithStack(
				&QueryBuilderError{errorMsg: "invalid offset"},
			)
		}

		if cfg.OffSet > 0 && offset > cfg.OffSet {
			offset = cfg.OffSet
		}

		builder = builder.Offset(offset)
	}

	return builder, nil
}

func GetQueryBuilderResult(
	req *http.Request,
	db Querier,
	bindvar int,
	dbFields DbFields,
	builder sq.SelectBuilder,
	cfg QueryBuilderConfig,
) (interface{}, error) {
	var err error
	var query string
	var args []interface{}

	if builder, err = GetQueryBuilder(
		req,
		dbFields,
		builder,
		cfg,
	); err != nil {
		return nil, errors.WithStack(err)
	}

	if query, args, err = builder.ToSql(); err != nil {
		return nil, errors.WithStack(err)
	}

	if cfg.IsCountQuery {
		var row *sqlx.Row
		var count uint64

		if row, err = db.QueryRowxRebind(bindvar, query, args...); err != nil {
			return nil, errors.WithStack(err)
		}

		if row.Scan(&count); err != nil {
			return nil, errors.WithStack(err)
		}

		return count, nil
	}

	var rows *sqlx.Rows

	data := make([]map[string]interface{}, 0, cfg.Limit)

	if rows, err = db.QueryxRebind(bindvar, query, args...); err != nil {
		return nil, errors.WithStack(err)
	}

	for rows.Next() {
		row := map[string]interface{}{}

		if err = rows.MapScanMultiLvl(row); err != nil {
			return nil, errors.WithStack(err)
		}

		data = append(data, row)
	}

	return data, nil
}

func GetQueryBuilderResultL(
	req *http.Request,
	db Querier,
	bindvar int,
	dbFields DbFields,
	logFunc func(error),
	customFunc func(map[string]interface{}) error,
	builder sq.SelectBuilder,
	cfg QueryBuilderConfig,
) (interface{}, error) {
	var err error
	var query string
	var args []interface{}

	if builder, err = GetQueryBuilder(
		req,
		dbFields,
		builder,
		cfg,
	); err != nil {
		return nil, errors.WithStack(err)
	}

	if query, args, err = builder.ToSql(); err != nil {
		return nil, errors.WithStack(err)
	}

	if cfg.IsCountQuery {
		var row *sqlx.Row
		var count uint64

		if row, err = db.QueryRowxRebind(bindvar, query, args...); err != nil {
			return nil, errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), query, args))
		}

		if row.Scan(&count); err != nil {
			return nil, errors.WithStack(err)
		}

		return count, nil
	}

	var rows *sqlx.Rows

	data := make([]map[string]interface{}, 0, cfg.Limit)

	if rows, err = db.QueryxRebind(bindvar, query, args...); err != nil {
		return nil, errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), query, args))
	}

	for rows.Next() {
		row := map[string]interface{}{}

		if err = rows.MapScanMultiLvl(row); err != nil {
			return nil, errors.WithStack(err)
		}

		if customFunc != nil {
			if err = customFunc(row); err != nil {
				return nil, errors.WithStack(err)
			}
		}

		data = append(data, row)
	}

	return data, nil
}

func GetDataAndCountBuilderResult(
	req *http.Request,
	db Querier,
	bindvar int,
	dbFields DbFields,
	dataBuilder sq.SelectBuilder,
	countBuilder sq.SelectBuilder,
	dataCfg QueryBuilderConfig,
	countCfg QueryBuilderConfig,
) ([]map[string]interface{}, uint64, error) {
	data, err := GetQueryBuilderResult(
		req,
		db,
		bindvar,
		dbFields,
		dataBuilder,
		dataCfg,
	)

	if err != nil {
		return nil, 0, err
	}

	count, err := GetQueryBuilderResult(
		req,
		db,
		bindvar,
		dbFields,
		countBuilder,
		dataCfg,
	)

	if err != nil {
		return nil, 0, err
	}

	return data.([]map[string]interface{}), count.(uint64), nil
}

func GetDataAndCountBuilderResultL(
	w http.ResponseWriter,
	req *http.Request,
	db Querier,
	bindvar int,
	dbFields DbFields,
	logFunc func(error),
	customFunc func(map[string]interface{}) error,
	dataBuilder sq.SelectBuilder,
	countBuilder sq.SelectBuilder,
	dataCfg QueryBuilderConfig,
	countCfg QueryBuilderConfig,
	statusCfg QueryBuilderErrorResponse,
) ([]map[string]interface{}, uint64, error) {
	errFunc := func(e error) {
		if logFunc != nil {
			logFunc(e)
		}

		var valErr *QueryBuilderError

		if errors.As(e, &valErr) {
			w.WriteHeader(statusCfg.QueryBuilderErrorStatus)
			w.Write([]byte(e.Error()))
		} else {
			w.WriteHeader(statusCfg.DatabaseErrorStatus)
			w.Write(statusCfg.DatabaseErrorResponse)
		}
	}

	data, err := GetQueryBuilderResultL(
		req,
		db,
		bindvar,
		dbFields,
		logFunc,
		customFunc,
		dataBuilder,
		dataCfg,
	)

	if err != nil {
		errFunc(err)
		return nil, 0, err
	}

	count, err := GetQueryBuilderResultL(
		req,
		db,
		bindvar,
		dbFields,
		logFunc,
		customFunc,
		countBuilder,
		countCfg,
	)

	if err != nil {
		errFunc(err)
		return nil, 0, err
	}

	return data.([]map[string]interface{}), count.(uint64), nil
}

func GetInnerBuilderResults(
	r *http.Request,
	builder sq.SelectBuilder,
	dbFields DbFields,
	dataCfg QueryBuilderConfig,
	countCfg QueryBuilderConfig,
) (InnerBuilderResult, error) {
	innerDataBuilder, err := GetQueryBuilder(
		r,
		dbFields,
		builder,
		dataCfg,
	)

	if err != nil {
		return InnerBuilderResult{}, errors.WithStack(err)
	}

	innerDataQuery, innerArgs, err := innerDataBuilder.ToSql()

	if err != nil {
		return InnerBuilderResult{}, errors.WithStack(err)
	}

	innerCountBuilder, err := GetQueryBuilder(
		r,
		dbFields,
		builder,
		countCfg,
	)

	if err != nil {
		return InnerBuilderResult{}, errors.WithStack(err)
	}

	innerCountQuery, _, err := innerCountBuilder.ToSql()

	if err != nil {
		return InnerBuilderResult{}, errors.WithStack(err)
	}

	return InnerBuilderResult{
		DataQuery:  innerDataQuery,
		CountQuery: innerCountQuery,
		Args:       innerArgs,
	}, nil
}
