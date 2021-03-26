package webutil

//go:generate mockgen -source=query_util.go -destination=../webutilmock/query_util_mock.go -package=webutilmock
//go:generate mockgen -source=query_util.go -destination=query_util_mock_test.go -package=webutil

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/knq/snaker"

	"github.com/jmoiron/sqlx"

	"reflect"

	"github.com/pkg/errors"
)

//////////////////////////////////////////////////////////////////
//------------------------ STRING CONSTS -----------------------
//////////////////////////////////////////////////////////////////

const (
	// Select string for queries
	Select = "select "

	sort   = "sort"
	group  = "group"
	filter = "filter"
)

//////////////////////////////////////////////////////////////////
//-------------------------- ENUM TYPES ------------------------
//////////////////////////////////////////////////////////////////

// Aggregate Types
const (
	AggregateCount = iota + 1
	AggregateSum
	AggregateAverage
	AggregateMin
	AggregateMax
)

//////////////////////////////////////////////////////////////////
//----------------------- CUSTOM ERRORS -------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrInvalidSort is error returned if client tries
	// to pass filter parameter that is not sortable
	ErrInvalidSort = errors.New("invalid sort")

	// ErrInvalidArray is error returned if client tries
	// to pass array parameter that is invalid array type
	ErrInvalidArray = errors.New("invalid array for field")

	// ErrInvalidValue is error returned if client tries
	// to pass filter parameter that had invalid field
	// value for certain field
	ErrInvalidValue = errors.New("invalid field value")
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

// SelectItem is struct used in conjunction with primeng's library api
type SelectItem struct {
	Value interface{} `json:"value" db:"value"`
	Label string      `json:"label" db:"label"`
}

// FilteredResults is struct used for dynamically filtered results
type FilteredResults struct {
	Data  interface{} `json:"data"`
	Total int         `json:"total"`
}

type queryError struct {
	isInvalidField     bool
	isInvalidOperation bool

	invalidField     string
	invalidOperation string
}

func (q *queryError) Error() string {
	if q.isInvalidField {
		return fmt.Sprintf("invalid field: '%s'", q.invalidField)
	}
	if q.isInvalidOperation {
		return fmt.Sprintf("invalid operation (%s) for field: '%s'", q.invalidOperation, q.invalidField)
	}

	return ""
}

// IsFieldError returns whether a given field name does not
// match a key in DbFields type
func (q *queryError) IsFieldError() bool {
	return q.isInvalidField
}

// IsOperationError returns whether an operation was tried
// on a given field that was not allowed
//
// This will be thrown if OperationConf within the FieldConfig type
// is set not to be filterable, sortable, or groupable and
// a client tries to do one of these actions
func (q *queryError) IsOperationError() bool {
	return q.isInvalidOperation
}

func (q *queryError) setInvalidField(field string) {
	q.invalidField = field
	q.isInvalidField = true
}

func (q *queryError) setInvalidOperation(field string, operation string) {
	q.invalidField = field
	q.invalidOperation = operation
	q.isInvalidOperation = true
}

// FilterError is error struct used when an error occurs when trying
// to perform a filter action on a field
type FilterError struct {
	*queryError

	isInvalidValue bool

	invalidValue interface{}
}

func (f *FilterError) Error() string {
	if f.queryError.Error() != "" {
		return f.queryError.Error()
	}
	if f.isInvalidValue {
		return fmt.Sprintf("invalid value (%v) for field '%s'", f.invalidValue, f.invalidField)
	}

	return ""
}

// IsValueError returns whether an invalid value was given to field
func (f *FilterError) IsValueError() bool {
	return f.isInvalidValue
}

func (f *FilterError) setInvalidValueError(field string, value interface{}) {
	f.invalidField = field
	f.invalidValue = value
	f.isInvalidValue = true
}

// SortError is error struct used when an error occurs when trying
// to perform a sort action on a field
type SortError struct {
	*queryError

	isInvalidDir bool

	invalidValue interface{}
}

func (s *SortError) Error() string {
	if s.queryError.Error() != "" {
		return s.queryError.Error()
	}
	if s.isInvalidDir {
		return fmt.Sprintf("invalid sort dir (%s) for field '%s'", s.invalidValue, s.invalidField)
	}

	return ""
}

// IsDirError returns whether given dir value for sort field
// was either "asc" or "desc"
func (s *SortError) IsDirError() bool {
	return s.isInvalidDir
}

func (s *SortError) setInvalidDirError(field, dir string) {
	s.invalidField = field
	s.invalidField = dir
	s.isInvalidDir = true
}

// GroupError is error struct used when an error occurs when trying
// to perform a group action on a field
type GroupError struct {
	*queryError
}

func (g *GroupError) Error() string {
	if g.queryError.Error() != "" {
		return g.queryError.Error()
	}

	return ""
}

// SliceError is error struct used when an error occurs when trying
// to perform a group action on a field
type SliceError struct {
	invalidSlice bool

	fieldType    string
	invalidField string
}

func (s *SliceError) Error() string {
	if s.invalidSlice {
		return fmt.Sprintf("invalid type (%s) within array for field: '%s'", s.fieldType, s.invalidField)
	}

	return ""
}

// IsSliceError returns whether a slice error occured which
// happens if any value within a slice is not a primitive type
func (s *SliceError) IsSliceError() bool {
	return s.invalidSlice
}

func (s *SliceError) setInvalidSliceError(field, fieldType string) {
	s.invalidField = field
	s.fieldType = fieldType
	s.invalidSlice = true
}

type GeneralJSON map[string]interface{}

func (g GeneralJSON) Value() (driver.Value, error) {
	j, err := json.Marshal(g)
	return j, err
}

func (g *GeneralJSON) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*g, ok = i.(map[string]interface{})
	if !ok {
		arr, ok := i.([]interface{})

		if ok {
			newV := make(map[string]interface{})
			newV["array"] = arr
			*g = newV
		} else {
			return errors.New("Not valid json")
		}
	}

	return nil
}

//////////////////////////////////////////////////////////////////
//----------------------- CONFIG STRUCTS ------------------------
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

// FieldConfig is meant to be a per database field config
// to determine if a user can perform a certain sql action
// and if user tries to perform action not allowed, throw error
type FieldConfig struct {
	// DBField should be the name of the database field
	// to apply configurations to
	DBField string

	// OperationConf is config to set to determine which sql
	// operations can be performed on DBField
	OperationConf OperationConfig
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

// QueryConfig is config for how the overall execution of the query
// is supposed to be performed
type QueryConfig struct {
	// SQLBindVar is used to determines what query placeholder parameters
	// will be converted to depending on what database being used
	// This is based off of the sqlx library
	SQLBindVar *int

	// TakeLimit is used to set max limit on number of
	// records that are returned from query
	TakeLimit *int

	// PrependFilterFields prepends filters to query before
	// ones passed by url query params
	PrependFilterFields []Filter

	// PrependGroupFields prepends groups to query before
	// ones passed by url query params
	PrependGroupFields []Group

	// PrependSortFields prepends sorts to query before
	// ones passed by url query params
	PrependSortFields []Sort

	//////////////////////////////////////////////////////////////////////////
	//
	// The exclude settings below are generally used in the case of
	// inner queries
	//
	// If we have a query that we need to call group by on when
	// doing dynamic filtering, generally speaking, it's because
	// we are querying against a join table that has a many to many
	// relationship so we get multiple results back when querying
	// for one of the id relationships in the table which we generally
	// do not want; we just want a distinct row for id we are
	// querying for
	//
	// However when we group by whatever field we need to group by,
	// if the client sends a request to group by or sort by another field,
	// database will error because in sql, you must use aggregate functions
	// for group by
	//
	// So to avoid this, we use the settings below to be able to
	// exclude the client's request for group by or sort by for the
	// inner query and then apply them to outer query
	//
	//////////////////////////////////////////////////////////////////////////

	// ExcludeFilters determines whether to exclude applying
	// filters from url query params
	// The PrependFilterFields property is NOT effected by this
	ExcludeFilters bool

	// ExcludeGroups determines whether to exclude applying
	// groups from url query params
	// The PrependGroupFields property is NOT effected by this
	ExcludeGroups bool

	// ExcludeSorts determines whether to exclude applying
	// sorts from url query params
	// The PrependSortFields property is NOT effected by this
	ExcludeSorts bool

	// ExcludeLimitWithOffset determines whether to exclude applying
	// limit and offset from url query params
	ExcludeLimitWithOffset bool

	// DisableGroupMod is used to determine if a user wants to disable
	// a query from automatically being modified to accommodate a
	// group by with order by without the client having to explictly send
	// group by parameters along with order by
	//
	// In sql, if you have a group by and order by, the order by field(s)
	// also have to appear in group by
	// The GetPreQueryResults() function and functions that utilize it will
	// automatically add the order by fields to the group by clause if they are
	// needed unless DisableGroupMod is set true
	DisableGroupMod bool
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

// Group is the group config struct for server side grouping
type Group struct {
	// Dir        string       `json:"dir"`
	Field string `json:"field"`
	// Aggregates []*Aggregate `json:"aggregates"`
}

// LimitOffset is config struct to set limit and offset of
// a query
type LimitOffset struct {
	Take int `json:"take"`
	Skip int `json:"skip"`
}

//////////////////////////////////////////////////////////////////
//----------------------- FUNCTIONS -------------------------
//////////////////////////////////////////////////////////////////

func getValueResults(
	query *string,
	prependArgs []interface{},
	isSelectQuery bool,
	req *http.Request,
	paramConf ParamConfig,
	queryConf QueryConfig,
	fields DbFields,
) ([]interface{}, error) {
	var allValues []interface{}
	var sorts []Sort
	var groups []Group
	var filters []Filter
	var limitOffset *LimitOffset
	var err error

	f := "filters"
	sk := "skip"
	so := "sorts"
	t := "take"
	g := "groups"

	sql := sqlx.QUESTION
	limit := -1

	if paramConf.Filter == nil {
		paramConf.Filter = &f
	}
	if paramConf.Skip == nil {
		paramConf.Skip = &sk
	}
	if paramConf.Sort == nil {
		paramConf.Sort = &so
	}
	if paramConf.Take == nil {
		paramConf.Take = &t
	}
	if paramConf.Group == nil {
		paramConf.Group = &g
	}

	if queryConf.SQLBindVar == nil {
		queryConf.SQLBindVar = &sql
	}
	if queryConf.TakeLimit == nil {
		queryConf.TakeLimit = &limit
	}

	allValues = make([]interface{}, 0)

	if prependArgs != nil {
		for _, v := range prependArgs {
			allValues = append(allValues, v)
		}
	}

	if !queryConf.ExcludeFilters {
		// Get filters and append to query
		if filters, err = GetFilterReplacements(
			req,
			query,
			*paramConf.Filter,
			queryConf,
			fields,
		); err != nil {
			//fmt.Printf("err 1: %s\n", err.Error())
			return nil, errors.Wrap(err, "")
		}
	}

	for _, v := range filters {
		if v.Value != nil {
			allValues = append(allValues, v.Value)
		}
	}

	if !queryConf.ExcludeGroups {
		// Get groups and append to query
		if groups, err = GetGroupReplacements(
			req,
			query,
			*paramConf.Group,
			queryConf,
			fields,
		); err != nil {
			//fmt.Printf("err 2: %s\n", err.Error())
			return nil, errors.Wrap(err, "")
		}
	}

	// If currently executing on a select query, apply sort
	// Else skip as sort doesn't matter on count query
	if isSelectQuery {
		if sorts, err = DecodeSorts(req, *paramConf.Sort); err != nil {
			//fmt.Printf("err 3: %s\n", err.Error())
			return nil, errors.Wrap(err, "")
		}

		// If there are group and sort slices and user has decided NOT
		// to disable the group mod feature, loop through the sort slice
		// and determine if there is duplicate field in the group slice
		// and if not add to list to then append to query
		if len(groups) > 0 && len(sorts) > 0 && !queryConf.DisableGroupMod {
			groupFields := make([]string, 0)

			for _, v := range sorts {
				hasGroupInSort := false
				if conf, ok := fields[v.Field]; ok {
					// Check that we allow for the field to be sorted
					// If not, throw SortError{}
					if conf.OperationConf.CanSortBy {
						for _, k := range groups {
							if v.Field == k.Field {
								hasGroupInSort = true
							}
						}

						// If there is group field that does not match current
						// sort field, add to our list
						if !hasGroupInSort {
							groupFields = append(groupFields, fields[v.Field].DBField)
						}
					} else {
						sortErr := &SortError{queryError: &queryError{}}
						sortErr.setInvalidOperation(v.Field, sort)
						return nil, sortErr
					}
				} else {
					sortErr := &SortError{queryError: &queryError{}}
					sortErr.setInvalidField(v.Field)
					return nil, sortErr
				}
			}

			// Loop through list where fields were not found
			// in group slice but in sort slice and append
			// to query
			for i, v := range groupFields {
				if i == 0 {
					*query += ","
				}

				*query += " " + v

				if i != len(groupFields)-1 {
					*query += ","
				}
			}
		}

		if !queryConf.ExcludeSorts {
			if sorts, err = GetSortReplacements(
				req,
				query,
				*paramConf.Sort,
				queryConf,
				fields,
			); err != nil {
				//fmt.Printf("err 4: %s\n", err.Error())
				return nil, errors.Wrap(err, "")
			}
		}

		if !queryConf.ExcludeLimitWithOffset && *queryConf.TakeLimit > -1 {
			if limitOffset, err = GetLimitWithOffsetValues(
				req,
				query,
				*paramConf.Take,
				*paramConf.Skip,
				*queryConf.TakeLimit,
				//queryConf.ExcludeLimitWithOffset,
			); err != nil {
				//fmt.Printf("err 5: %s\n", err.Error())
				return nil, errors.Wrap(err, "")
			}

			if limitOffset != nil {
				allValues = append(allValues, limitOffset.Take, limitOffset.Skip)
			}
		}

		// fmt.Printf("select query: %s\n", *query)
	}
	//fmt.Printf("values: %v\n", allValues)

	if *query, allValues, err = InQueryRebind(
		*queryConf.SQLBindVar,
		*query,
		allValues...,
	); err != nil {
		// fmt.Printf("query: %s\n", *query)
		// fmt.Printf("args: %s\n", allValues)
		// fmt.Printf("err 6: %s\n", err.Error())
		return nil, errors.Wrap(err, "")
	}

	return allValues, nil
}

// GetQueriedAndCountResults is a wrapper function for GetQueriedResults()
// and GetCountResults() functions and simply returns the values for both
func GetQueriedAndCountResults(
	query string,
	countQuery string,
	prependArgs []interface{},
	fields DbFields,
	req *http.Request,
	db Querier,
	paramConf ParamConfig,
	queryConf QueryConfig,
) (*sqlx.Rows, int, error) {
	rower, err := GetQueriedResults(
		query,
		prependArgs,
		fields,
		req,
		db,
		paramConf,
		queryConf,
	)

	if err != nil {
		return nil, 0, errors.Wrap(err, "")
	}

	//fmt.Printf("past query results\n")

	count, err := GetCountResults(
		countQuery,
		prependArgs,
		fields,
		req,
		db,
		paramConf,
		queryConf,
	)

	if err != nil {
		return nil, 0, errors.Wrap(err, "")
	}

	return rower, count, nil
}

// GetCountResults should take in count query that will return
// single row and column with total count of results with all
// the filters applied to query
func GetCountResults(
	countQuery string,
	prependArgs []interface{},
	fields DbFields,
	req *http.Request,
	db Querier,
	paramConf ParamConfig,
	queryConf QueryConfig,
) (int, error) {
	var results []interface{}
	var err error

	if results, err = getValueResults(
		&countQuery,
		prependArgs,
		false,
		req,
		paramConf,
		queryConf,
		fields,
	); err != nil {
		return 0, errors.Wrap(err, "")
	}

	rower, err := db.Queryx(countQuery, results...)

	if err != nil {
		fmt.Printf("query: %s\n", countQuery)
		fmt.Printf("values: %v\n", results)
		return 0, err
	}

	totalCount := 0

	for rower.Next() {
		var count int
		err = rower.Scan(&count)

		if err != nil {
			if err == sql.ErrNoRows {
				return 0, nil
			}

			return 0, err
		}

		totalCount += count
	}

	return totalCount, nil
}

// GetPreQueryResults gathers all of the replacement values and
// appends all the neccessary clauses to query but doesn't execute
func GetPreQueryResults(
	query *string,
	prependArgs []interface{},
	fields DbFields,
	req *http.Request,
	paramConf ParamConfig,
	queryConf QueryConfig,
) ([]interface{}, error) {
	var results []interface{}
	var err error

	if results, err = getValueResults(
		query,
		prependArgs,
		true,
		req,
		paramConf,
		queryConf,
		fields,
	); err != nil {
		//fmt.Printf("query: %s\n", *query)
		return nil, errors.Wrap(err, "")
	}

	return results, nil
}

// GetPreCountQueryResults gathers all of the replacement values and
// appends all the neccessary clauses to query but doesn't execute
func GetPreCountQueryResults(
	countQuery *string,
	prependArgs []interface{},
	fields DbFields,
	req *http.Request,
	paramConf ParamConfig,
	queryConf QueryConfig,
) ([]interface{}, error) {
	var results []interface{}
	var err error

	if results, err = getValueResults(
		countQuery,
		prependArgs,
		false,
		req,
		paramConf,
		queryConf,
		fields,
	); err != nil {
		//fmt.Printf("get result err: %s\n", err.Error())
		return nil, errors.Wrap(err, "")
	}

	return results, nil
}

// GetPreQueryAndCountResults gathers all of the replacement values and
// appends all the neccessary clauses to query but doesn't execute
func GetPreQueryAndCountResults(
	query *string,
	countQuery *string,
	prependArgs []interface{},
	fields DbFields,
	req *http.Request,
	paramConf ParamConfig,
	queryConf QueryConfig,
) ([]interface{}, []interface{}, error) {
	var results, countResults []interface{}
	var err error

	if results, err = getValueResults(
		query,
		prependArgs,
		true,
		req,
		paramConf,
		queryConf,
		fields,
	); err != nil {
		//fmt.Printf("get result err: %s\n", err.Error())
		return nil, nil, errors.Wrap(err, "")
	}

	if countResults, err = getValueResults(
		countQuery,
		prependArgs,
		false,
		req,
		paramConf,
		queryConf,
		fields,
	); err != nil {
		//fmt.Printf("get result err: %s\n", err.Error())
		return nil, nil, errors.Wrap(err, "")
	}

	return results, countResults, err
}

// GetQueriedResults dynamically adds filters, sorts and groups to query
// based on query params passed given from url and returns queried results
//
// This is a wrapper function for GetPreQueryResults() function that
// executes the query and returns results
func GetQueriedResults(
	query string,
	prependArgs []interface{},
	fields DbFields,
	req *http.Request,
	db Querier,
	paramConf ParamConfig,
	queryConf QueryConfig,
) (*sqlx.Rows, error) {
	values, err := GetPreQueryResults(
		&query,
		prependArgs,
		fields,
		req,
		paramConf,
		queryConf,
	)

	//fmt.Printf("query: %s\n", query)
	// fmt.Printf("args: %s\n", values)

	if err != nil {
		// fmt.Printf("get queried err: %s\n", err.Error())
		// fmt.Printf("query: %s\n", query)
		return nil, errors.Wrap(err, "")
	}

	rows, err := db.Queryx(query, values...)

	if err != nil {
		// fmt.Printf("query: %s\n", query)
		// fmt.Printf("values: %v\n", values)
		return nil, errors.WithStack(err)
	}

	return rows, nil
}

////////////////////////////////////////////////////////////
// GET REPLACEMENT FUNCTIONS
////////////////////////////////////////////////////////////

// GetFilterReplacements will decode passed paramName parameter from *http.Request into []Filter
// It will then apply these filters to passed query and return extracted values
// Applies "where" or "and" to query string depending on whether the query string
// already contains a where clause
//
// Throws FilterError{} or json.SyntaxError{} error type if error occurs
func GetFilterReplacements(
	req *http.Request,
	query *string,
	paramName string,
	queryConf QueryConfig,
	fields DbFields,
) ([]Filter, error) {
	var err error
	var filters []Filter
	allFilters := make([]Filter, 0)

	if len(queryConf.PrependFilterFields) > 0 {
		ApplyFilterText(query)

		if filters, err = ReplaceFilterFields(
			query,
			queryConf.PrependFilterFields,
			fields,
			true,
		); err != nil {
			return nil, errors.Wrap(err, "")
		}

		allFilters = append(allFilters, filters...)
	} else {
		queryConf.PrependFilterFields = make([]Filter, 0)
	}

	if !queryConf.ExcludeFilters {
		if filters, err = DecodeFilters(req, paramName); err != nil {
			return nil, errors.Wrap(err, "")
		}

		if len(filters) > 0 {
			ApplyFilterText(query)

			if filters, err = ReplaceFilterFields(query, filters, fields, false); err != nil {
				return nil, errors.Wrap(err, "")
			}

			allFilters = append(allFilters, filters...)
		}
	}

	return allFilters, nil
}

// GetSortReplacements will decode passed paramName parameter from *http.Request into []Sort
// It will then apply these sorts to passed query and return extracted values
// Will apply "order by" text to query if not found
//
// Throws SortError{} or json.UnmarshalTypeError{} error type if error occurs
func GetSortReplacements(
	r *http.Request,
	query *string,
	paramName string,
	queryConf QueryConfig,
	fields DbFields,
) ([]Sort, error) {
	var sortSlice []Sort
	var err error

	if len(queryConf.PrependSortFields) > 0 {
		ApplySortText(query)

		if err = ReplaceSortFields(
			query,
			queryConf.PrependSortFields,
			fields,
			true,
		); err != nil {
			return nil, errors.Wrap(err, "")
		}
	} else {
		queryConf.PrependSortFields = make([]Sort, 0)
	}

	if !queryConf.ExcludeSorts {
		if sortSlice, err = DecodeSorts(r, paramName); err != nil {
			return nil, errors.Wrap(err, "")
		}

		if len(sortSlice) > 0 {
			ApplySortText(query)

			if err = ReplaceSortFields(query, sortSlice, fields, false); err != nil {
				return nil, errors.Wrap(err, "")
			}
		} else {
			sortSlice = make([]Sort, 0)
		}
	}

	allSorts := make([]Sort, 0, len(queryConf.PrependSortFields)+len(sortSlice))

	for _, v := range queryConf.PrependSortFields {
		allSorts = append(allSorts, v)
	}
	for _, v := range sortSlice {
		allSorts = append(allSorts, v)
	}

	return allSorts, nil
}

// GetGroupReplacements will decode passed paramName parameter from *http.Request into []Group
// It will then apply these groups to passed query and return extracted values
// Will apply "group by" text to query if not found
//
// Throws GroupError{} or json.UnmarshalTypeError{} error type if error occurs
func GetGroupReplacements(
	req *http.Request,
	query *string,
	paramName string,
	queryConf QueryConfig,
	fields DbFields,
) ([]Group, error) {
	var groupSlice []Group
	var err error

	if len(queryConf.PrependGroupFields) > 0 {
		ApplyGroupText(query)

		if err = ReplaceGroupFields(
			query,
			queryConf.PrependGroupFields,
			fields,
			true,
		); err != nil {
			return nil, errors.Wrap(err, "")
		}
	} else {
		queryConf.PrependGroupFields = make([]Group, 0)
	}

	if !queryConf.ExcludeGroups {
		if groupSlice, err = DecodeGroups(req, paramName); err != nil {
			return nil, errors.Wrap(err, "")
		}

		if len(groupSlice) > 0 {
			ApplyGroupText(query)

			if err = ReplaceGroupFields(query, groupSlice, fields, false); err != nil {
				return nil, errors.Wrap(err, "")
			}
		}
	} else {
		groupSlice = make([]Group, 0)
	}

	allGroups := make([]Group, 0, len(queryConf.PrependGroupFields)+len(groupSlice))

	for _, v := range queryConf.PrependGroupFields {
		allGroups = append(allGroups, v)
	}
	for _, v := range groupSlice {
		allGroups = append(allGroups, v)
	}

	return allGroups, nil
}

// GetLimitWithOffsetValues takes in take and skip query param with
// a take limit and applies to query and returns skip and take values
// from request
func GetLimitWithOffsetValues(
	req *http.Request,
	query *string,
	takeParam,
	skipParam string,
	takeLimit int,
) (*LimitOffset, error) {
	var err error
	var takeInt, skipInt int

	take := req.FormValue(takeParam)
	skip := req.FormValue(skipParam)

	if take == "" {
		takeInt = takeLimit
	} else {
		if takeInt, err = strconv.Atoi(take); err != nil {
			return nil, errors.Wrap(err, "")
		}

		if takeInt > takeLimit {
			takeInt = takeLimit
		}
	}

	if skip == "" {
		skipInt = 0
	} else {
		if skipInt, err = strconv.Atoi(skip); err != nil {
			return nil, errors.Wrap(err, "")
		}
	}

	ApplyLimit(query)
	return &LimitOffset{Take: takeInt, Skip: skipInt}, nil
}

////////////////////////////////////////////////////////////
// DECODE FUNCTIONS
////////////////////////////////////////////////////////////

// DecodeFilters will use passed paramName parameter to extract json encoded
// filter from passed *http.Request and decode into Filter
// If paramName is not found in *http.Request, error will be thrown
// Will also throw error if can't properly decode
func DecodeFilters(req *http.Request, paramName string) ([]Filter, error) {
	var filterSlice []Filter
	var err error

	if err = decodeQueryParams(req, paramName, &filterSlice); err != nil {
		return nil, errors.Wrap(err, "")
	}

	return filterSlice, nil
}

// DecodeSorts will use passed paramName parameter to extract json encoded
// sort from passed *http.Request and decode into Sort
// If paramName is not found in *http.Request, error will be thrown
// Will also throw error if can't properly decode
func DecodeSorts(req *http.Request, paramName string) ([]Sort, error) {
	var sortSlice []Sort
	var err error

	if err = decodeQueryParams(req, paramName, &sortSlice); err != nil {
		return nil, errors.Wrap(err, "")
	}

	return sortSlice, nil
}

// DecodeGroups will use passed paramName parameter to extract json encoded
// group from passed *http.Request and decode into Group
// If paramName is not found in *http.Request, error will be thrown
// Will also throw error if can't properly decode
func DecodeGroups(req *http.Request, paramName string) ([]Group, error) {
	var groupSlice []Group
	var err error

	if err = decodeQueryParams(req, paramName, &groupSlice); err != nil {
		return nil, err
	}

	return groupSlice, nil
}

func decodeQueryParams(req *http.Request, paramName string, val interface{}) error {
	formVal := req.FormValue(paramName)

	if formVal != "" {
		param, err := url.QueryUnescape(formVal)

		if err != nil {
			return errors.Wrap(err, "")
		}

		err = json.Unmarshal([]byte(param), &val)

		if err != nil {
			return errors.Wrap(err, "")
		}
	}

	return nil
}

////////////////////////////////////////////////////////////
// REPLACE FUNCTIONS
////////////////////////////////////////////////////////////

// ReplaceFilterFields is used to replace query field names and values from slice of filters
// along with verifying that they have right values and applying changes to query
// This function does not apply "where" string for query so one must do it before
// passing query
func ReplaceFilterFields(query *string, filters []Filter, fields DbFields, isPrependFilters bool) ([]Filter, error) {
	var err error
	validFilters := make([]Filter, 0)

	applyCheck := func(idx int, filterVal Filter, dbField string) error {
		if err = FilterCheck(&filterVal); err != nil {
			return errors.Wrap(err, "")
		}

		validFilters = append(validFilters, filterVal)
		applyAnd := true

		if idx == len(filters)-1 {
			applyAnd = false
		}

		filterVal.Field = dbField
		ApplyFilter(query, filterVal, applyAnd)

		return nil
	}

	for i, v := range filters {
		if !isPrependFilters {

			// Check if current filter is within our fields map
			// If it is, check that it is allowed to be filtered
			// by and then check if given parameters are valid
			// If valid, apply filter to query
			// Else throw error
			if conf, ok := fields[v.Field]; ok && conf.OperationConf.CanFilterBy {
				if err = applyCheck(i, v, conf.DBField); err != nil {
					return nil, err
				}
			} else {
				filterErr := &FilterError{queryError: &queryError{}}
				filterErr.setInvalidField(v.Field)
				return nil, errors.Wrap(filterErr, "")
			}
		} else {
			if err = applyCheck(i, v, fields[v.Field].DBField); err != nil {
				return nil, err
			}
		}
	}

	return validFilters, nil
}

// ReplaceSortFields is used to replace query field names and values from slice of sorts
// along with verifying that they have right values and applying changes to query
// This function does not apply "order by" string for query so one must do it before
// passing query
func ReplaceSortFields(query *string, sorts []Sort, fields DbFields, isPrependSort bool) error {
	var err error

	applyCheck := func(idx int, sortVal Sort, dbField string) error {
		if err = SortCheck(sortVal); err != nil {
			return errors.Wrap(err, "")
		}

		applyComma := true

		if idx == len(sorts)-1 {
			applyComma = false
		}

		sortVal.Field = dbField
		ApplySort(query, sortVal, applyComma)

		return nil
	}

	for i, v := range sorts {
		if !isPrependSort {
			// Check if current sort is within our fields map
			// If it is, check that it is allowed to be sorted
			// by and then check if given parameters are valid
			// If valid, apply sort to query
			// Else throw error
			if conf, ok := fields[v.Field]; ok && conf.OperationConf.CanSortBy {
				if err = applyCheck(i, v, conf.DBField); err != nil {
					return err
				}
			} else {
				sortErr := &SortError{queryError: &queryError{}}
				sortErr.setInvalidField(v.Field)
				return errors.Wrap(sortErr, "")
			}
		} else {
			if err = applyCheck(i, v, fields[v.Field].DBField); err != nil {
				return err
			}
		}
	}

	return nil
}

// ReplaceGroupFields is used to replace query field names and values from slice of groups
// along with verifying that they have right values and applying changes to query
func ReplaceGroupFields(query *string, groups []Group, fields DbFields, isPrependGroup bool) error {
	applyGroup := func(idx int, groupVal Group, dbField string) {
		applyComma := true

		if idx == len(groups)-1 {
			applyComma = false
		}

		groupVal.Field = dbField
		ApplyGroup(query, groupVal, applyComma)
	}

	for i, v := range groups {
		if !isPrependGroup {
			// Check if current sort is within our fields map
			// If it is, check that it is allowed to be grouped
			// by and then check if given parameters are valid
			// If valid, apply sort to query
			// Else throw error
			if conf, ok := fields[v.Field]; ok && conf.OperationConf.CanGroupBy {
				applyGroup(i, v, conf.DBField)
			} else {
				groupErr := &GroupError{queryError: &queryError{}}
				groupErr.setInvalidField(v.Field)
				return errors.Wrap(groupErr, "")
			}
		} else {
			applyGroup(i, v, fields[v.Field].DBField)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////
// APPLY FUNCTIONS
////////////////////////////////////////////////////////////

// ApplyFilter applies the filter passed to the query passed
//
// The applyAnd parameter is used to determine if the query should have
// an "and" added to the end
func ApplyFilter(query *string, filter Filter, applyAnd bool) {
	_, ok := filter.Value.([]interface{})

	if ok {
		*query += " " + filter.Field + " in (?)"
	} else {
		switch filter.Operator {
		case "eq":
			*query += " " + filter.Field + " = ?"
		case "neq":
			*query += " " + filter.Field + " != ?"
		case "startswith":
			*query += " " + filter.Field + " ilike ? || '%'"
		case "endswith":
			*query += " " + filter.Field + " ilike '%' || ?"
		case "contains":
			*query += " " + filter.Field + " ilike '%' || ? || '%'"
		case "doesnotcontain":
			*query += " " + filter.Field + " not ilike '%' || ? || '%'"
		case "isnull":
			*query += " " + filter.Field + " is null"
		case "isnotnull":
			*query += " " + filter.Field + " is not null"
		case "isempty":
			*query += " " + filter.Field + " = ''"
		case "isnotempty":
			*query += " " + filter.Field + " != ''"
		case "lt":
			*query += " " + filter.Field + " < ?"
		case "lte":
			*query += " " + filter.Field + " <= ?"
		case "gt":
			*query += " " + filter.Field + " > ?"
		case "gte":
			*query += " " + filter.Field + " >= ?"
		}
	}

	// If there is more in filter slice, append "and"
	if applyAnd {
		*query += " and"
	}
}

func ApplyFilterText(query *string) {
	filterExp := regexp.MustCompile(`(?i)(\n|\t|\s)where(\n|\t|\s)`)
	if f := filterExp.FindString(*query); f == "" {
		*query += " where "
	} else {
		*query += " and "
	}
}

// ApplySort applies the sort passed to the query passed
//
// The addComma paramter is used to determine if the query should have
// ","(comma) appended to the query
func ApplySort(query *string, sort Sort, addComma bool) {
	*query += " " + sort.Field

	if sort.Dir == "asc" {
		*query += " asc"
	} else {
		*query += " desc"
	}

	if addComma {
		*query += ","
	}
}

func ApplySortText(query *string) {
	sortExp := regexp.MustCompile(`(?i)(\n|\t|\s)order(\n|\t|\s)`)
	if s := sortExp.FindString(*query); s == "" {
		*query += " order by "
	} else {
		*query += ","
	}
}

// ApplyGroup applies the group passed to the query passed
//
// The addComma parameter is used to determine if the query should have
// ","(comma) appended to the query
func ApplyGroup(query *string, group Group, addComma bool) {
	*query += " " + group.Field

	if addComma {
		*query += ","
	}
}

func ApplyGroupText(query *string) {
	groupExp := regexp.MustCompile(`(?i)(\n|\t|\s)group(\n|\t|\s)`)
	if g := groupExp.FindString(*query); g == "" {
		*query += " group by "
	} else {
		*query += ","
	}
}

// ApplyLimit takes given query and applies limit and offset criteria
func ApplyLimit(query *string) {
	*query += " limit ? offset ?"
}

// ApplyOrdering takes given query and applies the given sort criteria
func ApplyOrdering(query *string, sort *Sort) {
	*query += " order by " + snaker.CamelToSnake(sort.Field) + " " + sort.Dir
}

////////////////////////////////////////////////////////////
// CHECK FUNCTIONS
////////////////////////////////////////////////////////////

// SortCheck checks to make sure that the "dir" field either has value "asc" or "desc"
// and if it doesn't, throw error
func SortCheck(s Sort) error {
	if s.Dir != "asc" && s.Dir != "desc" {
		sortErr := &SortError{queryError: &queryError{}}
		sortErr.setInvalidDirError(s.Field, s.Dir)
		return sortErr
	}

	return nil
}

// FilterCheck checks to make sure that the values passed to each filter is valid
// The types passed should be primitive types
// If the filter operator equals "isnull" or "isnotnull" and a value is passed,
// this function will "self correct" and null out the value as those operators
// should not have values attached to them
func FilterCheck(f *Filter) error {
	//var r interface{}

	validTypes := []string{"string", "int", "float", "float64", "int64"}
	hasValidType := false

	if f.Operator != "isnull" && f.Operator != "isnotnull" {
		// First check if value sent is slice
		list, ok := f.Value.([]interface{})

		// If slice, then loop through and make sure all items in list
		// are primitive type, else throw error
		//
		// Else check the value of the single item
		if ok {
			for _, t := range list {
				someType := reflect.TypeOf(t).String()

				for _, v := range validTypes {
					if someType == v {
						hasValidType = true
						break
					}
				}

				if !hasValidType {
					sliceErr := &SliceError{}
					sliceErr.setInvalidSliceError(f.Field, someType)
					return sliceErr
				}
			}

			// r = list
		} else {
			validTypes = append(validTypes, "bool")

			if f.Value == nil {
				filterErr := &FilterError{queryError: &queryError{}}
				filterErr.setInvalidValueError(f.Field, f.Value)
				return filterErr
			}

			someType := reflect.TypeOf(f.Value).String()

			for _, v := range validTypes {
				if someType == v {
					hasValidType = true
					break
				}
			}

			if !hasValidType {
				filterErr := &FilterError{queryError: &queryError{}}
				filterErr.setInvalidValueError(f.Field, someType)
				return filterErr
			}

			// r = f.Value
		}
	} else {
		f.Value = nil
	}

	return nil
}

////////////////////////////////////////////////////////////
// UTIL FUNCTIONS
////////////////////////////////////////////////////////////

// CountSelect take column string and applies count select
func CountSelect(column string) string {
	return fmt.Sprintf("count(%s) as total", column)
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

// HasFilterOrServerError determines if passed error is a filter based error
// or a server type error and writes appropriate response to client
func HasFilterOrServerError(w http.ResponseWriter, r *http.Request, err error, retryDB RetryDB, clientStatus int, conf ServerErrorConfig) bool {
	if err != nil {
		SetHTTPResponseDefaults(&conf.ServerErrorResponse, http.StatusInternalServerError, []byte(serverErrTxt))

		var fe *FilterError
		var se *SortError
		var ge *GroupError

		if errors.As(err, &fe) || errors.As(err, &se) || errors.As(err, &ge) {
			w.WriteHeader(clientStatus)
			w.Write([]byte(errors.Cause(err).Error()))
		} else {
			return dbError(w, r, err, retryDB, conf)
		}

		return true
	}

	return false
}

// QuerySelectItems is utility function for querying against a table and returning
// a list of SelectItem structs to be used with primeng's library components
func QuerySelectItems(db SqlxDB, bindVar int, query string, args ...interface{}) ([]SelectItem, error) {
	var err error

	if query, args, err = InQueryRebind(bindVar, query, args...); err != nil {
		return nil, err
	}

	var items []SelectItem

	if err = db.Select(&items, query, args...); err != nil {
		return []SelectItem{}, err
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
