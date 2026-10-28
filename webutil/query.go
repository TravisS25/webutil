package webutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
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

// DbTable describes the query surface of one database table: the request-facing
// field configs plus the table-level full-text search settings shared by every
// builder that queries the table (data, count, distinct).
type DbTable struct {
	// Fields are the filterable/sortable/searchable field configs, keyed by
	// the name a request sends in the filter/sort params.
	Fields DbFields

	// SearchIndex is the table's bm25 expression index (e.g.
	// "worker_idx_bm25"). Required whenever QueryConfig.SearchParam is
	// configured, the request carries search tokens and Fields hold
	// CanSearchBy fields: the relevance ordering resolves the index
	// explicitly because the driver prepares these statements. The index
	// must be built over the concatenation of the CanSearchBy fields sorted
	// by DBField.
	SearchIndex string

	// SearchTextConfig is the PostgreSQL text search configuration parsed by
	// to_tsquery (e.g. "english"); empty means SearchDefaultTextConfig. It
	// must match the text_config SearchIndex was created with.
	SearchTextConfig string

	// SearchLookups are lookup tables joined by a foreign key on the searched
	// table whose text should also match the global text search (e.g. a
	// dropdown's name column). A row matches when either its own expression
	// matches or its FK points at a lookup row matching the query: the
	// lookup side resolves to ids with an inline subquery, so the searched
	// table's own bm25 index still drives the scan. Only honored while the
	// search is active, which requires CanSearchBy fields and SearchIndex as
	// above; the lookup tables themselves need no bm25 index (they are small
	// and evaluated by the subquery).
	SearchLookups []SearchLookup
}

// SearchLookup describes one lookup table searchable through the foreign key
// on the searched table. All identifiers are server-side configuration (the
// same trust level as FieldConfig.DBField), never request input.
type SearchLookup struct {
	// FKField is the foreign key column on the searched table (e.g.
	// "bid_option.area_id").
	FKField string

	// Table is the lookup table the FK references (e.g. "area"); its
	// "id" column is what FKField is matched against.
	Table string

	// Fields are the lookup table's text columns matched by the query (e.g.
	// ["area.name"]). Concatenated in the given order, each coalesced to ''.
	Fields []string
}

//////////////////////////////////////////////////////////////////
//-------------------------- STRUCTS --------------------------
//////////////////////////////////////////////////////////////////

type QueryConfig struct {
	FilterParam string
	OrderParam  string
	LimitParam  string
	OffsetParam string

	// SearchParam is the request parameter carrying the global text search
	// (e.g. "q"). Empty disables search handling entirely for this config.
	// The predicate is applied whenever the param has tokens and the
	// DbTable's Fields hold CanSearchBy fields, so data, count and distinct
	// builders of an endpoint all see the same rows; set it on each config
	// those builders use.
	SearchParam string

	// DefaultOrder is applied to the ORDER BY clause when OrderParam is
	// configured (i.e. this config owns ordering) and the request carried no
	// explicit sort. With an active search the relevance score is ordered
	// first and DefaultOrder breaks ties; without a search it is the whole
	// ordering. Configs that must not order (count queries, grouped
	// subqueries) leave both OrderParam and DefaultOrder empty.
	DefaultOrder string

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

// type SelectItem struct {
// 	Value string `json:"value" mapstructure:"value" db:"value" alias:"select.value"`
// 	Text  string `json:"text" mapstructure:"text" db:"text" alias:"select.text"`
// }

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

	// CanSearchBy determines whether the field participates in the global
	// text search (QueryConfig.SearchParam): GetQueryBuilder concatenates
	// every CanSearchBy field into one BM25 expression predicate over the
	// caller's text search index. A field can be searchable without being
	// filterable or sortable (and vice versa).
	CanSearchBy bool
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

// FilterGroup combines a list of filters and/or nested filter groups
// using a boolean logic operator ("and" or "or"). An empty Logic
// defaults to "and", allowing callers to nest groups to build
// arbitrary AND/OR expressions.
type FilterGroup struct {
	Logic string       `json:"logic"`
	Items []FilterItem `json:"items"`
}

// FilterItem represents a single entry within a FilterGroup. It is
// either a leaf Filter (when a "field" key is present) or a nested
// FilterGroup (when no "field" key is present).
type FilterItem struct {
	*Filter
	*FilterGroup
}

func (item *FilterItem) UnmarshalJSON(data []byte) error {
	var probe map[string]json.RawMessage

	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}

	if _, ok := probe["field"]; ok {
		var filter Filter

		if err := json.Unmarshal(data, &filter); err != nil {
			return err
		}

		item.Filter = &filter
		item.FilterGroup = nil
		return nil
	}

	var group FilterGroup

	if err := json.Unmarshal(data, &group); err != nil {
		return err
	}

	item.FilterGroup = &group
	item.Filter = nil
	return nil
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
		// if a, ok := arg.(driver.Valuer); ok {
		// 	var err error
		// 	arg, err = a.Value()
		// 	if err != nil {
		// 		return "", nil, err
		// 	}
		// }

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

func ExecInsert(
	ctx context.Context,
	db qrm.Executable,
	bindVar int,
	table string,
	columns []string,
	values ...any,
) error {
	if len(columns) != len(values) {
		return fmt.Errorf("columns (%d) and values (%d) count mismatch", len(columns), len(values))
	}

	placeholders := make([]string, len(values))
	for i := range values {
		switch bindVar {
		case DOLLAR_SQL_BIND_VAR:
			placeholders[i] = "?"
		case QUESTION_SQL_BIND_VAR:
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := db.ExecContext(ctx, query, values...)
	return err
}

func ExecDB(
	ctx context.Context,
	db qrm.Executable,
	bindType int,
	query string,
	args []any,
) (sql.Result, error) {
	var err error

	newQuery, newArgs, err := InQueryRebind(bindType, query, args...)
	if err != nil {
		return nil, fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), query, args)
	}

	return db.ExecContext(ctx, newQuery, newArgs...)
}

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

	rows, err := db.QueryContext(ctx, newQuery, newArgs...)
	if err != nil {
		if DebugPrintQueryOutput {
			return fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), newQuery, args)
		}
		return err
	}
	defer rows.Close()

	return setRowResults(rows, rowUpdate, destPtr)
}

////////////////////////////////////////////////////////////
// --------------------- BUILDER FUNCTIONS ---------------
////////////////////////////////////////////////////////////

func ExecQueryBuilder(
	ctx context.Context,
	req *http.Request,
	builder sq.SelectBuilder,
	table DbTable,
	db qrm.Queryable,
	bindVar int,
	queryCfg QueryConfig,
	rowUpdate func(row any) error,
	destPtr any,
) error {
	var err error

	if builder, err = GetQueryBuilder(
		req,
		builder,
		table,
		queryCfg,
	); err != nil {
		return errors.WithStack(err)
	}

	rows, err := getRowsFromBuilder(ctx, builder, db, bindVar)
	if err != nil {
		return err
	}

	return setRowResults(rows, rowUpdate, destPtr)
}

func QueryDataAndCountResults(
	ctx context.Context,
	req *http.Request,
	dataBuilder sq.SelectBuilder,
	countBuilder sq.SelectBuilder,
	table DbTable,
	db qrm.Queryable,
	bindVar int,
	dataQueryCfg QueryConfig,
	countQueryCfg QueryConfig,
	rowUpdate func(row any) error,
	dataDest any,
	countDest *uint64,
) error {
	err := ExecQueryBuilder(
		ctx,
		req,
		dataBuilder,
		table,
		db,
		bindVar,
		dataQueryCfg,
		rowUpdate,
		dataDest,
	)
	if err != nil {
		return err
	}

	err = ExecQueryBuilder(
		ctx,
		req,
		countBuilder,
		table,
		db,
		bindVar,
		countQueryCfg,
		nil,
		countDest,
	)
	if err != nil {
		return err
	}

	return nil
}

func GetOrderByResults(
	r *http.Request,
	builder sq.SelectBuilder,
	table DbTable,
	orderByParam,
	defaultOrderBy string,
) (string, []any, error) {
	orderDataBuilder, err := GetQueryBuilder(
		r,
		builder,
		table,
		QueryConfig{
			OrderParam: orderByParam,
		},
	)
	if err != nil {
		return "", nil, errors.WithStack(err)
	}

	if defaultOrderBy != "" && r.FormValue(orderByParam) == "" {
		orderDataBuilder = orderDataBuilder.OrderBy(defaultOrderBy)
	}

	return orderDataBuilder.ToSql()
}

func parseFilterParam(filterParam string) (FilterGroup, error) {
	// A bare array is the legacy format and is treated as a top level
	// group whose items are combined with "and".
	if strings.HasPrefix(strings.TrimSpace(filterParam), "[") {
		var items []FilterItem

		if err := json.Unmarshal([]byte(filterParam), &items); err != nil {
			return FilterGroup{}, err
		}

		return FilterGroup{Logic: "and", Items: items}, nil
	}

	var group FilterGroup

	if err := json.Unmarshal([]byte(filterParam), &group); err != nil {
		return FilterGroup{}, err
	}

	return group, nil
}

func buildFilterGroupSqlizer(group FilterGroup, table DbTable) (sq.Sqlizer, error) {
	logic := strings.ToLower(group.Logic)

	if logic != "" && logic != "and" && logic != "or" {
		return nil, errors.WithStack(
			QueryBuilderError{errorMsg: fmt.Sprintf("invalid logic %q for filter group", group.Logic)},
		)
	}

	parts := make([]sq.Sqlizer, 0, len(group.Items))

	for i := range group.Items {
		item := group.Items[i]

		var (
			part sq.Sqlizer
			err  error
		)

		switch {
		case item.Filter != nil:
			part, err = buildFilterSqlizer(*item.Filter, table)
		case item.FilterGroup != nil:
			part, err = buildFilterGroupSqlizer(*item.FilterGroup, table)
		default:
			continue
		}

		if err != nil {
			return nil, err
		}

		if part != nil {
			parts = append(parts, part)
		}
	}

	switch len(parts) {
	case 0:
		return nil, nil
	case 1:
		return parts[0], nil
	}

	if logic == "or" {
		return sq.Or(parts), nil
	}

	return sq.And(parts), nil
}

func buildFilterSqlizer(filter Filter, table DbTable) (sq.Sqlizer, error) {
	dbField, ok := table.Fields[filter.Field]
	if !ok {
		return nil, errors.WithStack(
			QueryBuilderError{errorMsg: fmt.Sprintf("invalid field %q for filter parameter", filter.Field)},
		)
	}

	if !dbField.OperationCfg.CanFilterBy {
		return nil, errors.WithStack(
			QueryBuilderError{fmt.Sprintf("field %q can not be filtered", filter.Field)},
		)
	}

	if filter.Value == nil {
		if filter.Operator != "isnull" && filter.Operator != "isnotnull" {
			return nil, errors.WithStack(
				QueryBuilderError{errorMsg: fmt.Sprintf("field %q does not contain value", filter.Field)},
			)
		}
	}

	fieldValue := filter.Value

	if dbField.ValueOverride != nil {
		var err error

		if fieldValue, err = dbField.ValueOverride(fieldValue); err != nil {
			return nil, errors.WithStack(
				QueryBuilderError{errorMsg: fmt.Sprintf("invalid value %q for field %q", fieldValue, filter.Field)},
			)
		}
	}

	switch filter.Operator {
	case "eq":
		return sq.Eq{dbField.DBField: fieldValue}, nil
	case "neq":
		return sq.NotEq{dbField.DBField: fieldValue}, nil
	case "startswith":
		return sq.ILike{dbField.DBField: fmt.Sprintf("%v%%", fieldValue)}, nil
	case "endswith":
		return sq.ILike{dbField.DBField: fmt.Sprintf("%%%v", fieldValue)}, nil
	case "contains":
		return sq.ILike{dbField.DBField: fmt.Sprintf("%%%v%%", fieldValue)}, nil
	case "doesnotcontain":
		return sq.NotILike{dbField.DBField: fmt.Sprintf("%%%v%%", fieldValue)}, nil
	case "isnull":
		return sq.Eq{dbField.DBField: nil}, nil
	case "isnotnull":
		return sq.NotEq{dbField.DBField: nil}, nil
	case "isempty":
		return sq.Eq{dbField.DBField: ""}, nil
	case "isnotempty":
		return sq.NotEq{dbField.DBField: ""}, nil
	case "lt":
		return sq.Lt{dbField.DBField: fmt.Sprintf("%v", fieldValue)}, nil
	case "lte":
		return sq.LtOrEq{dbField.DBField: fmt.Sprintf("%v", fieldValue)}, nil
	case "gt":
		return sq.Gt{dbField.DBField: fmt.Sprintf("%v", fieldValue)}, nil
	case "gte":
		return sq.GtOrEq{dbField.DBField: fmt.Sprintf("%v", fieldValue)}, nil
	default:
		return nil, errors.WithStack(
			QueryBuilderError{errorMsg: fmt.Sprintf("invalid operator for field %q", filter.Field)},
		)
	}
}

// SearchDefaultTextConfig is the PostgreSQL text search configuration used to
// parse the global search query when DbTable.SearchTextConfig is empty.
// It must match the text_config of the BM25 index being searched.
const SearchDefaultTextConfig = "english"

// Caps that keep one search request from building an unbounded query: at most
// maxSearchTokens tokens, each at most maxSearchTokenLen runes.
const (
	maxSearchTokens   = 10
	maxSearchTokenLen = 64
)

// searchTokenPattern extracts the alphanumeric runs a search query is made of.
// PostgreSQL's text search parser tokenizes the indexed documents by splitting
// on punctuation and lowercasing, so dropping punctuation here keeps the query
// aligned with what the index can match and simultaneously removes every
// query-parser metacharacter before the value reaches SQL as a bind parameter.
var searchTokenPattern = regexp.MustCompile(`[\p{L}\p{N}]+`)

// searchableFields returns the CanSearchBy fields of table.Fields sorted by
// DBField. The sort makes the concatenated expression deterministic — a map
// has no order of its own — and the order must match the column order the
// table's BM25 expression index was built with.
func searchableFields(table DbTable) []FieldConfig {
	fields := make([]FieldConfig, 0, len(table.Fields))

	for _, field := range table.Fields {
		if field.OperationCfg.CanSearchBy {
			fields = append(fields, field)
		}
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].DBField < fields[j].DBField
	})

	return fields
}

// searchExpr renders the sorted searchable fields as one concatenated text
// expression — the exact shape the BM25 expression index is built on, so both
// the predicate and the relevance ordering resolve that index.
func searchExpr(fields []FieldConfig) string {
	parts := make([]string, 0, len(fields))

	for _, field := range fields {
		parts = append(parts, fmt.Sprintf("coalesce(%s::text, '')", field.DBField))
	}

	return "(" + strings.Join(parts, " || ' ' || ") + ")"
}

// lookupSearchExpr renders a lookup's Fields as one concatenated text
// expression over the lookup table (each field coalesced to ”). Unlike the
// searched table's own expression this one is evaluated by the inline
// subquery, so it needs no matching expression index.
func lookupSearchExpr(lookup SearchLookup) string {
	parts := make([]string, 0, len(lookup.Fields))

	for _, field := range lookup.Fields {
		parts = append(parts, fmt.Sprintf("coalesce(%s::text, '')", field))
	}

	return "(" + strings.Join(parts, " || ' ' || ") + ")"
}

// searchTokens splits a raw search value into at most maxSearchTokens
// alphanumeric tokens of at most maxSearchTokenLen runes each. Tokens are
// lowercased because to_tsquery matches the lowercased lexemes the index was
// built with.
func searchTokens(q string) []string {
	tokens := searchTokenPattern.FindAllString(q, maxSearchTokens)

	for i, token := range tokens {
		runes := []rune(token)

		if len(runes) > maxSearchTokenLen {
			token = string(runes[:maxSearchTokenLen])
		}

		tokens[i] = strings.ToLower(token)
	}

	return tokens
}

// bm25Query renders tokens as the query text both full-text operators
// receive: every token gets a prefix wildcard and the tokens are OR-joined
// ("tok1:* | tok2:*"). Tokens are alphanumeric (searchTokens), so the string
// carries no query syntax beyond the wildcards added here, and both call
// sites pass it as a bind parameter.
func bm25Query(tokens []string) string {
	return strings.Join(tokens, ":* | ") + ":*"
}

// searchTextConfig returns the text search configuration to parse queries
// with: the table's own value or SearchDefaultTextConfig.
func searchTextConfig(table DbTable) string {
	if table.SearchTextConfig != "" {
		return table.SearchTextConfig
	}

	return SearchDefaultTextConfig
}

func GetQueryBuilder(
	r *http.Request,
	builder sq.SelectBuilder,
	table DbTable,
	cfg QueryConfig,
) (sq.SelectBuilder, error) {
	var err error
	var ok bool

	filterParam := r.FormValue(cfg.FilterParam)
	orderParam := r.FormValue(cfg.OrderParam)
	limitParam := r.FormValue(cfg.LimitParam)
	offsetParam := r.FormValue(cfg.OffsetParam)

	if filterParam != "" {
		filterGroup, err := parseFilterParam(filterParam)
		if err != nil {
			return sq.SelectBuilder{}, errors.WithStack(QueryBuilderError{
				errorMsg: "invalid filter parameter"},
			)
		}

		filterSqlizer, err := buildFilterGroupSqlizer(filterGroup, table)
		if err != nil {
			return sq.SelectBuilder{}, errors.WithStack(err)
		}

		if filterSqlizer != nil {
			builder = builder.Where(filterSqlizer)
		}
	}

	// Global text search over the fields marked CanSearchBy: one BM25
	// predicate over the concatenated expression, bound as a single
	// parameter. It applies to every builder this config is used for — data,
	// count and distinct subqueries agree on the matched rows — while the
	// relevance ordering further below only runs for configs that own
	// ordering (OrderParam set).
	searchActive := false
	var searchExprSql, searchQuery string

	if cfg.SearchParam != "" {
		tokens := searchTokens(r.FormValue(cfg.SearchParam))
		fields := searchableFields(table)

		if len(tokens) > 0 && len(fields) > 0 {
			if table.SearchIndex == "" {
				return sq.SelectBuilder{}, errors.WithStack(QueryBuilderError{
					errorMsg: "searchable fields require DbTable.SearchIndex",
				})
			}

			searchExprSql = searchExpr(fields)
			searchQuery = bm25Query(tokens)
			searchActive = true

			// The predicate ORs the table's own BM25 match with one inline
			// id-resolution subquery per lookup: a row whose FK points at a
			// lookup row matching the same tokens also matches. The table's
			// expression stays the first disjunct so the bm25 index scan
			// remains reachable; the lookup subqueries are tiny lookups over
			// the small lookup tables.
			predParts := []string{searchExprSql + " @@ to_tsquery(?, ?)"}
			predArgs := []any{searchTextConfig(table), searchQuery}

			for _, lookup := range table.SearchLookups {
				if lookup.FKField == "" || lookup.Table == "" || len(lookup.Fields) == 0 {
					return sq.SelectBuilder{}, errors.WithStack(QueryBuilderError{
						errorMsg: "search lookups require FKField, Table and Fields",
					})
				}

				predParts = append(predParts, fmt.Sprintf(
					"%s = any(select %s.id from %s where %s @@ to_tsquery(?, ?))",
					lookup.FKField,
					lookup.Table,
					lookup.Table,
					lookupSearchExpr(lookup),
				))
				predArgs = append(predArgs, searchTextConfig(table), searchQuery)
			}

			if len(predParts) == 1 {
				builder = builder.Where(sq.Expr(predParts[0], predArgs...))
			} else {
				builder = builder.Where(sq.Expr(
					"("+strings.Join(predParts, " or ")+")",
					predArgs...,
				))
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

			if dbField, ok = table.Fields[sort.Field]; !ok {
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
	} else if cfg.OrderParam != "" {
		// This config owns ordering and the request sent no explicit sort:
		// BM25 relevance first while the search is active (scores are
		// negative, so ascending order ranks the best match first), then
		// DefaultOrder as tiebreak; without a search, just DefaultOrder.
		// Relevance requires the ordered BM25 index scan, which enumerates
		// only the documents the bm25 posting lists contain — rows matched
		// solely through a lookup subquery would be dropped by its
		// post-filter before scoring. Tables with SearchLookups therefore
		// take DefaultOrder only: the search buys findability, not
		// relevance, and every match rows up by the table's own ordering.
		if searchActive && len(table.SearchLookups) == 0 {
			builder = builder.OrderByClause(
				searchExprSql+" <@> to_bm25query(?, ?)",
				searchQuery,
				table.SearchIndex,
			)
		}

		if cfg.DefaultOrder != "" {
			builder = builder.OrderBy(cfg.DefaultOrder)
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

func setRowResults(rows *sql.Rows, rowUpdate func(row any) error, destPtr any) error {
	defer rows.Close()

	var box any
	isArr := false

	destVal := reflect.ValueOf(destPtr)

	if destVal.IsValid() && destVal.Elem().Kind() == reflect.Slice {
		box = make([]any, 0)
		isArr = true
	}

	hasRow := false
	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("webutil: error getting number of columns: %w", err)
	}

	for rows.Next() {
		var row any
		var err error
		hasRow = true

		if len(cols) > 1 {
			val := make(map[string]any)

			if err = MapScanner(rows, val); err != nil {
				return errors.WithStack(err)
			}

			if rowUpdate != nil {
				if err = rowUpdate(&val); err != nil {
					return errors.WithStack(err)
				}
			}

			row = val
		} else {
			if err = rows.Scan(&row); err != nil {
				return errors.WithStack(err)
			}

			if rowUpdate != nil {
				if err = rowUpdate(&row); err != nil {
					return errors.WithStack(err)
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
		return errors.WithStack(sql.ErrNoRows)
	}

	jsonBytes, err := json.Marshal(box)
	if err != nil {
		return errors.WithStack(err)
	}

	if destVal.IsValid() {
		if err = json.Unmarshal(jsonBytes, destPtr); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
