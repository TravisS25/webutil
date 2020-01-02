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
//----------------------- INTERFACES --------------------------
//////////////////////////////////////////////////////////////////

// FormRequest is used to get form values from url string
// Will mostly come from http.Request
type FormRequest interface {
	FormValue(string) string
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
	isSelectQuery bool,
	req FormRequest,
	paramConf ParamConfig,
	queryConf QueryConfig,
	fields DbFields,
) ([]interface{}, error) {
	var allValues []interface{}
	var sorts []Sort
	var groups []Group
	var limitOffset *LimitOffset
	var err error

	f := "filters"
	sk := "skip"
	so := "sorts"
	t := "take"
	g := "groups"

	sql := sqlx.QUESTION
	limit := 100

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

	// Get filters and append to query
	if _, allValues, err = GetFilterReplacements(
		req,
		query,
		*paramConf.Filter,
		queryConf,
		fields,
	); err != nil {
		return nil, errors.Wrap(err, "")
	}

	// Get groups and append to query
	if groups, err = GetGroupReplacements(
		req,
		query,
		*paramConf.Group,
		queryConf,
		fields,
	); err != nil {
		return nil, errors.Wrap(err, "")
	}

	// If currently executing on a select query, apply sort
	// Else skip as sort doesn't matter on count query
	if isSelectQuery {
		if sorts, err = DecodeSorts(req, *paramConf.Sort); err != nil {
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

						// If there is group field that matches current
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

		if sorts, err = GetSortReplacements(
			req,
			query,
			*paramConf.Sort,
			queryConf,
			fields,
		); err != nil {
			return nil, errors.Wrap(err, "")
		}

		if limitOffset, err = GetLimitWithOffsetValues(
			req,
			query,
			*paramConf.Take,
			*paramConf.Skip,
			*queryConf.TakeLimit,
			queryConf.ExcludeLimitWithOffset,
		); err != nil {
			return nil, errors.Wrap(err, "")
		}

		if limitOffset != nil {
			allValues = append(allValues, limitOffset.Take, limitOffset.Skip)
		}
	}

	if *query, allValues, err = InQueryRebind(
		*queryConf.SQLBindVar,
		*query,
		allValues...,
	); err != nil {
		return nil, errors.Wrap(err, "")
	}

	return allValues, nil
}

// GetQueriedAndCountResults is a wrapper function for GetQueriedResults()
// and GetCountResults() functions and simply returns the values for both
func GetQueriedAndCountResults(
	query *string,
	countQuery *string,
	fields DbFields,
	req FormRequest,
	db Querier,
	paramConf ParamConfig,
	queryConf QueryConfig,
) (*sql.Rows, int, error) {
	rower, err := GetQueriedResults(
		query,
		fields,
		req,
		db,
		paramConf,
		queryConf,
	)

	if err != nil {
		return nil, 0, errors.Wrap(err, "")
	}

	count, err := GetCountResults(
		countQuery,
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
	countQuery *string,
	fields DbFields,
	req FormRequest,
	db Querier,
	paramConf ParamConfig,
	queryConf QueryConfig,
) (int, error) {
	var results []interface{}
	var err error

	if results, err = getValueResults(
		countQuery,
		false,
		req,
		paramConf,
		queryConf,
		fields,
	); err != nil {
		return 0, errors.Wrap(err, "")
	}

	rower, err := db.Query(*countQuery, results...)

	if err != nil {
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
	fields DbFields,
	req FormRequest,
	paramConf ParamConfig,
	queryConf QueryConfig,
) ([]interface{}, error) {
	var results []interface{}
	var err error

	if results, err = getValueResults(
		query,
		true,
		req,
		paramConf,
		queryConf,
		fields,
	); err != nil {
		return nil, errors.Wrap(err, "")
	}

	return results, nil
}

// GetQueriedResults dynamically adds filters, sorts and groups to query
// based on query params passed given from url and returns queried results
//
// This is a wrapper function for GetPreQueryResults() function that
// executes the query and returns results
func GetQueriedResults(
	query *string,
	fields DbFields,
	req FormRequest,
	db Querier,
	paramConf ParamConfig,
	queryConf QueryConfig,
) (*sql.Rows, error) {
	values, err := GetPreQueryResults(
		query,
		fields,
		req,
		paramConf,
		queryConf,
	)

	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return db.Query(*query, values...)
}

////////////////////////////////////////////////////////////
// GET REPLACEMENT FUNCTIONS
////////////////////////////////////////////////////////////

// GetFilterReplacements will decode passed paramName parameter from FormRequest into []Filter
// It will then apply these filters to passed query and return extracted values
// Applies "where" or "and" to query string depending on whether the query string
// already contains a where clause
//
// Throws FilterError{} or json.SyntaxError{} error type if error occurs
func GetFilterReplacements(
	req FormRequest,
	query *string,
	paramName string,
	queryConf QueryConfig,
	fields DbFields,
) ([]Filter, []interface{}, error) {
	var err error
	var allFilters, filters []Filter
	var replacements, prependReplacements, allReplacements []interface{}

	filterExp := regexp.MustCompile(`(?i)(\n|\t|\s)where(\n|\t|\s)`)

	if len(queryConf.PrependFilterFields) > 0 {
		if f := filterExp.FindString(*query); f == "" {
			*query += " where"
		} else {
			*query += " and"
		}

		if prependReplacements, err = ReplaceFilterFields(
			query,
			queryConf.PrependFilterFields,
			fields,
		); err != nil {
			return nil, nil, errors.Wrap(err, "")
		}
	} else {
		queryConf.PrependFilterFields = make([]Filter, 0)
	}

	if !queryConf.ExcludeFilters {
		if filters, err = DecodeFilters(req, paramName); err != nil {
			return nil, nil, errors.Wrap(err, "")
		}

		if len(filters) > 0 {
			if f := filterExp.FindString(*query); f == "" {
				*query += " where"
			} else {
				*query += " and"
			}

			if replacements, err = ReplaceFilterFields(query, filters, fields); err != nil {
				return nil, nil, errors.Wrap(err, "")
			}
		}
	} else {
		filters = make([]Filter, 0)
	}

	allFilters = make([]Filter, 0, len(queryConf.PrependFilterFields)+len(filters))
	allReplacements = make([]interface{}, 0, len(prependReplacements)+len(replacements))

	for _, v := range queryConf.PrependFilterFields {
		allFilters = append(allFilters, v)
	}
	for _, v := range filters {
		allFilters = append(allFilters, v)
	}

	for _, v := range prependReplacements {
		allReplacements = append(allReplacements, v)
	}
	for _, v := range replacements {
		allReplacements = append(allReplacements, v)
	}

	return allFilters, allReplacements, nil
}

// GetSortReplacements will decode passed paramName parameter from FormRequest into []Sort
// It will then apply these sorts to passed query and return extracted values
// Will apply "order by" text to query if not found
//
// Throws SortError{} or json.UnmarshalTypeError{} error type if error occurs
func GetSortReplacements(
	r FormRequest,
	query *string,
	paramName string,
	queryConf QueryConfig,
	fields DbFields,
) ([]Sort, error) {
	var allSorts, sortSlice []Sort
	var err error

	orderExp := regexp.MustCompile(`(?i)(\n|\t|\s)order(\n|\t|\s)`)

	if len(queryConf.PrependSortFields) > 0 {
		if s := orderExp.FindString(*query); s == "" {
			*query += " order by "
		} else {
			*query += ","
		}

		if err = ReplaceSortFields(
			query,
			queryConf.PrependSortFields,
			fields,
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
			if s := orderExp.FindString(*query); s == "" {
				*query += " order by "
			} else {
				*query += ","
			}

			if err = ReplaceSortFields(query, sortSlice, fields); err != nil {
				return nil, errors.Wrap(err, "")
			}
		}
	} else {
		sortSlice = make([]Sort, 0)
	}

	allSorts = make([]Sort, 0, len(queryConf.PrependSortFields)+len(sortSlice))

	for _, v := range queryConf.PrependSortFields {
		allSorts = append(allSorts, v)
	}
	for _, v := range sortSlice {
		allSorts = append(allSorts, v)
	}

	return allSorts, nil
}

// GetGroupReplacements will decode passed paramName parameter from FormRequest into []Group
// It will then apply these groups to passed query and return extracted values
// Will apply "group by" text to query if not found
//
// Throws GroupError{} or json.UnmarshalTypeError{} error type if error occurs
func GetGroupReplacements(
	req FormRequest,
	query *string,
	paramName string,
	queryConf QueryConfig,
	fields DbFields,
) ([]Group, error) {
	var allGroups, groupSlice []Group
	var err error

	groupExp := regexp.MustCompile(`(?i)(\n|\t|\s)group(\n|\t|\s)`)

	if len(queryConf.PrependGroupFields) > 0 {
		if g := groupExp.FindString(*query); g == "" {
			*query += " group by "
		} else {
			*query += ","
		}

		if err = ReplaceGroupFields(
			query,
			queryConf.PrependGroupFields,
			fields,
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
			if g := groupExp.FindString(*query); g == "" {
				*query += " group by "
			} else {
				*query += ","
			}

			if err = ReplaceGroupFields(query, groupSlice, fields); err != nil {
				return nil, errors.Wrap(err, "")
			}
		}
	} else {
		groupSlice = make([]Group, 0)
	}

	allGroups = make([]Group, 0, len(queryConf.PrependGroupFields)+len(groupSlice))

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
	req FormRequest,
	query *string,
	takeParam,
	skipParam string,
	takeLimit int,
	excludeLimitOffset bool,
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

	//replacements := []interface{}{takeInt, skipInt}
	ApplyLimit(query)
	return &LimitOffset{Take: takeInt, Skip: skipInt}, nil
}

////////////////////////////////////////////////////////////
// DECODE FUNCTIONS
////////////////////////////////////////////////////////////

// DecodeFilters will use passed paramName parameter to extract json encoded
// filter from passed FormRequest and decode into Filter
// If paramName is not found in FormRequest, error will be thrown
// Will also throw error if can't properly decode
func DecodeFilters(req FormRequest, paramName string) ([]Filter, error) {
	var filterSlice []Filter
	var err error

	if err = decodeQueryParams(req, paramName, &filterSlice); err != nil {
		return nil, errors.Wrap(err, "")
	}

	return filterSlice, nil
}

// DecodeSorts will use passed paramName parameter to extract json encoded
// sort from passed FormRequest and decode into Sort
// If paramName is not found in FormRequest, error will be thrown
// Will also throw error if can't properly decode
func DecodeSorts(req FormRequest, paramName string) ([]Sort, error) {
	var sortSlice []Sort
	var err error

	if err = decodeQueryParams(req, paramName, &sortSlice); err != nil {
		return nil, errors.Wrap(err, "")
	}

	return sortSlice, nil
}

// DecodeGroups will use passed paramName parameter to extract json encoded
// group from passed FormRequest and decode into Group
// If paramName is not found in FormRequest, error will be thrown
// Will also throw error if can't properly decode
func DecodeGroups(req FormRequest, paramName string) ([]Group, error) {
	var groupSlice []Group
	var err error

	if err = decodeQueryParams(req, paramName, &groupSlice); err != nil {
		return nil, err
	}

	return groupSlice, nil
}

func decodeQueryParams(req FormRequest, paramName string, val interface{}) error {
	formVal := req.FormValue(paramName)

	if formVal != "" {
		param, err := url.QueryUnescape(formVal)

		if err != nil {
			return err
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
func ReplaceFilterFields(query *string, filters []Filter, fields DbFields) ([]interface{}, error) {
	var err error
	replacements := make([]interface{}, 0, len(filters))

	for i, v := range filters {
		var r interface{}

		// Check if current filter is within our fields map
		// If it is, check that it is allowed to be filtered
		// by and then check if given parameters are valid
		// If valid, apply filter to query
		// Else throw error
		if conf, ok := fields[v.Field]; ok {
			if !conf.OperationConf.CanFilterBy {
				filterErr := &FilterError{queryError: &queryError{}}
				filterErr.setInvalidOperation(v.Field, filter)
				return nil, errors.Wrap(filterErr, "")
			}

			if r, err = FilterCheck(v); err != nil {
				return nil, errors.Wrap(err, "")
			}

			replacements = append(replacements, r)
			applyAnd := true

			if i == len(filters)-1 {
				applyAnd = false
			}

			v.Field = conf.DBField
			ApplyFilter(query, v, applyAnd)
		} else {
			filterErr := &FilterError{queryError: &queryError{}}
			filterErr.setInvalidField(v.Field)
			return nil, errors.Wrap(filterErr, "")
		}
	}

	return replacements, nil
}

// ReplaceSortFields is used to replace query field names and values from slice of sorts
// along with verifying that they have right values and applying changes to query
// This function does not apply "order by" string for query so one must do it before
// passing query
func ReplaceSortFields(query *string, sorts []Sort, fields DbFields) error {
	var err error

	for i, v := range sorts {
		// Check if current sort is within our fields map
		// If it is, check that it is allowed to be sorted
		// by and then check if given parameters are valid
		// If valid, apply sort to query
		// Else throw error
		if conf, ok := fields[v.Field]; ok {
			if !conf.OperationConf.CanSortBy {
				sortErr := &SortError{queryError: &queryError{}}
				sortErr.setInvalidOperation(v.Field, sort)
				return errors.Wrap(sortErr, "")
			}

			if err = SortCheck(v); err != nil {
				return err
			}

			addComma := true

			if i == len(sorts)-1 {
				addComma = false
			}

			v.Field = conf.DBField
			ApplySort(query, v, addComma)
		} else {
			sortErr := &SortError{queryError: &queryError{}}
			sortErr.setInvalidField(v.Field)
			return errors.Wrap(sortErr, "")
		}
	}

	return nil
}

// ReplaceGroupFields is used to replace query field names and values from slice of groups
// along with verifying that they have right values and applying changes to query
func ReplaceGroupFields(query *string, groups []Group, fields DbFields) error {
	for i, v := range groups {
		// Check if current sort is within our fields map
		// If it is, check that it is allowed to be grouped
		// by and then check if given parameters are valid
		// If valid, apply sort to query
		// Else throw error
		if conf, ok := fields[v.Field]; ok {
			if !conf.OperationConf.CanGroupBy {
				groupErr := &GroupError{queryError: &queryError{}}
				groupErr.setInvalidOperation(v.Field, group)
				return errors.Wrap(groupErr, "")
			}

			addComma := true

			if i == len(groups)-1 {
				addComma = false
			}

			v.Field = conf.DBField
			ApplyGroup(query, v, addComma)
		} else {
			groupErr := &GroupError{queryError: &queryError{}}
			groupErr.setInvalidField(v.Field)
			return errors.Wrap(groupErr, "")
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
// Returns the "Value" property of Filter type if all checks pass
func FilterCheck(f Filter) (interface{}, error) {
	var r interface{}

	validTypes := []string{"string", "int", "float", "float64", "int64"}
	hasValidType := false

	if f.Value != "" && f.Operator != "isnull" && f.Operator != "isnotnull" {
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
					return nil, sliceErr
				}
			}

			r = list
		} else {
			validTypes = append(validTypes, "bool")

			if f.Value == nil {
				filterErr := &FilterError{queryError: &queryError{}}
				filterErr.setInvalidValueError(f.Field, f.Value)
				return nil, filterErr
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
				return nil, filterErr
			}

			r = f.Value
		}
	}

	return r, nil
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

// SetRowerResults gathers the results within rower and applies
// it to the cache store
// func SetRowerResults(
// 	rower Rower,
// 	cache CacheStore,
// 	cacheSetup CacheSetup,
// ) error {
// 	var err error
// 	columns, err := rower.Columns()

// 	if err != nil {
// 		return err
// 	}

// 	count := len(columns)
// 	values := make([]interface{}, count)
// 	valuePtrs := make([]interface{}, count)
// 	rows := make([]interface{}, 0)
// 	forms := make([]FormSelection, 0)

// 	for rower.Next() {
// 		form := FormSelection{}

// 		for i := range columns {
// 			valuePtrs[i] = &values[i]
// 		}

// 		err = rower.Scan(valuePtrs...)

// 		if err != nil {
// 			return err
// 		}

// 		row := make(map[string]interface{}, 0)
// 		var idVal interface{}

// 		for i, k := range columns {
// 			var v interface{}
// 			//var formVal string

// 			val := values[i]

// 			if k == "id" {
// 				idVal = val
// 			}

// 			switch val.(type) {
// 			case int64:
// 				v = strconv.FormatInt(val.(int64), IntBase)
// 			case *int64:
// 				t := val.(*int64)
// 				if t != nil {
// 					v = strconv.FormatInt(*t, IntBase)
// 				}
// 			case []byte:
// 				t := val.([]byte)
// 				v, err = strconv.ParseFloat(string(t), IntBitSize)
// 				if err != nil {
// 					panic(err)
// 				}
// 			default:
// 				v = val
// 			}

// 			var columnName string

// 			if snaker.IsInitialism(columns[i]) {
// 				columnName = strings.ToLower(columns[i])
// 			} else {
// 				camelCaseJSON := snaker.ForceLowerCamelIdentifier(columns[i])
// 				firstLetter := strings.ToLower(string(camelCaseJSON[0]))
// 				columnName = firstLetter + camelCaseJSON[1:]
// 			}

// 			row[columnName] = v

// 			if cacheSetup.CacheSelectionConf.ValueColumn == columnName {
// 				form.Value = v
// 			}

// 			if cacheSetup.CacheSelectionConf.TextColumn == columnName {
// 				form.Text = v
// 			}
// 		}

// 		rowBytes, err := json.Marshal(&row)

// 		if err != nil {
// 			return err
// 		}

// 		var cacheID string

// 		switch idVal.(type) {
// 		case int64:
// 			cacheID = strconv.FormatInt(idVal.(int64), IntBase)
// 		case int:
// 			cacheID = strconv.Itoa(idVal.(int))
// 		default:
// 			return errors.New("Invalid id type")
// 		}

// 		cache.Set(
// 			fmt.Sprintf(cacheSetup.CacheIDKey, cacheID),
// 			rowBytes,
// 			0,
// 		)

// 		rows = append(rows, row)
// 		forms = append(forms, form)
// 	}

// 	rowsBytes, err := json.Marshal(&rows)

// 	if err != nil {
// 		return err
// 	}

// 	formBytes, err := json.Marshal(&forms)

// 	if err != nil {
// 		return err
// 	}

// 	cache.Set(cacheSetup.CacheListKey, rowsBytes, 0)
// 	cache.Set(cacheSetup.CacheSelectionConf.FormSelectionKey, formBytes, 0)
// 	return nil
// }

// HasFilterOrServerError determines if passed error is a filter based error
// or a server type error and writes appropriate response to client
func HasFilterOrServerError(w http.ResponseWriter, err error, errResp ServerAndClientErrorConfig) bool {
	if err != nil {
		SetHTTPResponseDefaults(&errResp.ClientErrorResponse, http.StatusNotAcceptable, []byte(err.Error()))
		SetHTTPResponseDefaults(&errResp.ServerErrorResponse, http.StatusInternalServerError, []byte(ErrServer.Error()))

		serverResp := func() {
			w.WriteHeader(*errResp.ServerErrorResponse.HTTPStatus)
			w.Write(errResp.ServerErrorResponse.HTTPResponse)
		}

		switch err.(type) {
		case *FilterError, *SortError, *GroupError:
			w.WriteHeader(*errResp.ClientErrorResponse.HTTPStatus)
			w.Write(errResp.ClientErrorResponse.HTTPResponse)
			return true
		default:
			if errResp.RecoverDB != nil {
				if err = errResp.RecoverDB(err); err != nil {
					serverResp()
					return true
				}
			} else {
				serverResp()
				return true
			}
		}
	}

	return false
}
