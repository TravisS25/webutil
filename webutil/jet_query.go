package webutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

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
	CustomFunc func(map[string]interface{}) error
}

type CountInputParams struct {
	QueryCfg QueryConfig
}

type QueryInputParams struct {
	QueryCfg            QueryConfig
	OrderByID           string
	CanMultiColumnOrder bool
	CanMultiColumnGroup bool
	IsCountQuery        bool
	CustomFunc          func(map[string]interface{}) error
}

func Query(
	ctx context.Context,
	decoderFunc func(dest interface{}) *mapstructure.DecoderConfig,
	db qrm.Queryable,
	bindType int,
	query string,
	args []interface{},
	destPtr interface{},
) error {
	destType := reflect.TypeOf(destPtr)

	if destType.Kind() != reflect.Ptr {
		return fmt.Errorf("webutil: destPtr parameter must be a pointer")
	}

	newQuery, newArgs, err := JetInQueryRebind(bindType, query, args...)
	if err != nil {
		return errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), query, args))
	}

	rows, err := db.QueryContext(ctx, newQuery, newArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	decoder, err := mapstructure.NewDecoder(decoderFunc(destPtr))
	if err != nil {
		return fmt.Errorf("webutil: error trying to create decoder: %s", err)
	}

	if destType.Elem().Kind() == reflect.Slice {
		var list []map[string]interface{}

		if err = QueryMapList(ctx, db, bindType, query, args, &list); err != nil {
			return err
		}

		if err = decoder.Decode(list); err != nil {
			return fmt.Errorf("webutil: error trying to decode into slice: %s", err)
		}
	} else if destType.Elem().Kind() == reflect.Struct {
		dest := map[string]interface{}{}
		found := false

		for rows.Next() {
			found = true

			if err = MapScanner(rows, dest); err != nil {
				return fmt.Errorf("webutil: error trying to scan row %s", err)
			}

			break
		}

		if !found {
			return sql.ErrNoRows
		}

		if err = decoder.Decode(dest); err != nil {
			return fmt.Errorf("webutil: error trying to decode into struct: %s", err)
		}
	} else {
		return fmt.Errorf("webutil: destination has to be a pointer to slice or pointer to struct")
	}

	return nil
}

func QueryRows(ctx context.Context, db qrm.Queryable, bindType int, query string, args []interface{}) (*sql.Rows, error) {
	newQuery, newArgs, err := JetInQueryRebind(bindType, query, args...)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), query, args))
	}

	return db.QueryContext(ctx, newQuery, newArgs...)
}

func QueryJetSingleColumn(ctx context.Context, db qrm.Queryable, bindType int, query string, args []interface{}, destPtr interface{}) error {
	data, ok := destPtr.(*[]interface{})
	if !ok {
		return fmt.Errorf("destPtr parameter must be pointer of []interface{}")
	}

	newQuery, newArgs, err := JetInQueryRebind(bindType, query, args...)
	if err != nil {
		return errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), query, args))
	}

	rows, err := db.QueryContext(ctx, newQuery, newArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id interface{}

		if rows.Scan(&id); err != nil {
			return err
		}

		*data = append(*data, id)
	}

	return nil
}

func QueryMapList(ctx context.Context, db qrm.Queryable, bindType int, query string, args []interface{}, destPtr interface{}) error {
	data, ok := destPtr.(*[]map[string]interface{})
	if !ok {
		return fmt.Errorf("destPtr parameter must be pointer of []map[string]interface{}")
	}

	newQuery, newArgs, err := JetInQueryRebind(bindType, query, args...)
	if err != nil {
		return errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), query, args))
	}

	rows, err := db.QueryContext(ctx, newQuery, newArgs...)
	if err != nil {
		return errors.WithStack(err)
	}
	defer rows.Close()

	for rows.Next() {
		row := map[string]interface{}{}

		if err = MapScanner(rows, row); err != nil {
			return errors.WithStack(err)
		}

		*data = append(*data, row)
	}

	return nil
}

func QueryJetCount(ctx context.Context, db qrm.Queryable, bindType int, query string, args []interface{}, dest *int64) error {
	newQuery, newArgs, err := JetInQueryRebind(bindType, query, args...)
	if err != nil {
		return errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), query, args))
	}

	rows, err := db.QueryContext(ctx, newQuery, newArgs...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if rows.Scan(dest); err != nil {
			return err
		}
	}

	return nil
}

func JetInQueryRebind(bindType int, query string, args ...interface{}) (string, []interface{}, error) {
	query, args, err := In(query, args...)
	if err != nil {
		return query, args, err
	}

	query = Rebind(bindType, query)
	return query, args, nil
}

func QueryDataResult(
	ctx context.Context,
	req *http.Request,
	builder sq.SelectBuilder,
	dbFields DbFields,
	db qrm.Queryable,
	bindVar int,
	dest interface{},
	params DataInputParams,
) error {
	data, ok := dest.(*[]map[string]interface{})
	if !ok {
		return fmt.Errorf("dest parameter must pointer to []map[string]interface{}; got %s", reflect.TypeOf(dest))
	}

	var err error

	if builder, err = GetNewQueryBuilder(
		req,
		dbFields,
		builder,
		params.QueryCfg,
	); err != nil {
		return errors.WithStack(err)
	}

	rows, err := getRowsFromBuilder(ctx, builder, db, bindVar)
	if err != nil {
		return err
	}

	for rows.Next() {
		row := map[string]interface{}{}

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
	dest interface{},
	params CountInputParams,
) error {
	count, ok := dest.(*uint64)
	if !ok {
		return fmt.Errorf("dest parameter must pointer to uint64; got %s", reflect.TypeOf(dest))
	}

	var err error

	if builder, err = GetNewQueryBuilder(
		req,
		dbFields,
		builder,
		params.QueryCfg,
	); err != nil {
		return errors.WithStack(err)
	}

	rows, err := getRowsFromBuilder(ctx, builder, db, bindVar)
	if err != nil {
		return err
	}

	for rows.Next() {
		for rows.Next() {
			if err = rows.Scan(count); err != nil {
				return errors.WithStack(err)
			}
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
	dataDest interface{},
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

func GetNewInnerBuilderResults(
	r *http.Request,
	builder sq.SelectBuilder,
	dbFields DbFields,
	defaultOrderBy string,
	queryCfg QueryConfig,
) (string, []interface{}, error) {
	if queryCfg.OrderParam == "" && defaultOrderBy != "" {
		builder = builder.OrderBy(defaultOrderBy)
	}

	innerDataBuilder, err := GetNewQueryBuilder(
		r,
		dbFields,
		builder,
		queryCfg,
	)
	if err != nil {
		return "", nil, errors.WithStack(err)
	}

	return innerDataBuilder.ToSql()

	// innerCountBuilder, err := GetNewQueryBuilder(
	// 	r,
	// 	dbFields,
	// 	builder,
	// 	countParams.QueryCfg,
	// )
	// if err != nil {
	// 	return nil, errors.WithStack(err)
	// }

	// innerCountQuery, _, err := innerCountBuilder.ToSql()
	// if err != nil {
	// 	return nil, errors.WithStack(err)
	// }

	// return &InnerBuilderResult{
	// 	DataQuery:  innerDataQuery,
	// 	CountQuery: innerCountQuery,
	// 	Args:       innerArgs,
	// }, nil
}

func GetNewQueryBuilder(
	r *http.Request,
	dbFields DbFields,
	builder sq.SelectBuilder,
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

	if limitParam != "" {
		var limit uint64

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

// ------------------------------------------------------------------------------------------

func GetJetQueryBuilder(
	r *http.Request,
	dbFields DbFields,
	builder sq.SelectBuilder,
	params QueryInputParams,
) (sq.SelectBuilder, error) {
	var err error
	var ok bool

	filterParam := r.FormValue(params.QueryCfg.FilterParam)
	orderParam := r.FormValue(params.QueryCfg.OrderParam)
	limitParam := r.FormValue(params.QueryCfg.LimitParam)
	offsetParam := r.FormValue(params.QueryCfg.OffsetParam)

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

			if !params.CanMultiColumnOrder {
				break
			}
		}
	}

	if limitParam != "" {
		var limit uint64

		if limit, err = strconv.ParseUint(
			limitParam,
			IntBase,
			IntBitSize,
		); err != nil {
			return sq.SelectBuilder{}, errors.WithStack(
				&QueryBuilderError{errorMsg: "invalid limit"},
			)
		}

		if params.QueryCfg.Limit > 0 && limit > params.QueryCfg.Limit {
			limit = params.QueryCfg.Limit
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

		if params.QueryCfg.OffSet > 0 && offset > params.QueryCfg.OffSet {
			offset = params.QueryCfg.OffSet
		}

		builder = builder.Offset(offset)
	}

	return builder, nil
}

func QueryBuilderResult(
	ctx context.Context,
	req *http.Request,
	builder sq.SelectBuilder,
	dbFields DbFields,
	db qrm.Queryable,
	bindVar int,
	dest interface{},
	params QueryInputParams,
) error {
	if params.IsCountQuery {
		if _, ok := dest.(*uint64); !ok {
			return fmt.Errorf("dest parameter must pointer to unit64 when 'IsCountQuery' is set")
		}
	} else {
		if _, ok := dest.(*[]map[string]interface{}); !ok {
			return fmt.Errorf("dest parameter must pointer to []map[string]interface{}; got %s", reflect.TypeOf(dest))
		}
	}

	var err error

	if builder, err = GetJetQueryBuilder(
		req,
		dbFields,
		builder,
		params,
	); err != nil {
		return errors.WithStack(err)
	}

	var query string
	var args []interface{}

	if query, args, err = builder.ToSql(); err != nil {
		return errors.WithStack(err)
	}

	resQuery := query
	resArgs := args

	if query, args, err = JetInQueryRebind(bindVar, query, args...); err != nil {
		return errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), resQuery, resArgs))
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	if params.IsCountQuery {
		for rows.Next() {
			if err = rows.Scan(dest); err != nil {
				return errors.WithStack(err)
			}
		}
	} else {
		data := dest.(*[]map[string]interface{})

		for rows.Next() {
			row := map[string]interface{}{}

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

		//fmt.Printf("final data: %+v", data)
	}

	return nil
}

func QueryDataAndCountBuilderResult(
	ctx context.Context,
	req *http.Request,
	dataBuilder sq.SelectBuilder,
	countBuilder sq.SelectBuilder,
	dbFields DbFields,
	db qrm.Queryable,
	bindVar int,
	dataDest interface{},
	countDest *uint64,
	dataParams QueryInputParams,
	countParams QueryInputParams,
) error {
	err := QueryBuilderResult(
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

	err = QueryBuilderResult(
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

func GetJetInnerBuilderResults(
	r *http.Request,
	builder sq.SelectBuilder,
	dbFields DbFields,
	dataParams QueryInputParams,
	countParams QueryInputParams,
) (*InnerBuilderResult, error) {
	innerDataBuilder, err := GetJetQueryBuilder(
		r,
		dbFields,
		builder,
		dataParams,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	innerDataQuery, innerArgs, err := innerDataBuilder.ToSql()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	innerCountBuilder, err := GetJetQueryBuilder(
		r,
		dbFields,
		builder,
		countParams,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	innerCountQuery, _, err := innerCountBuilder.ToSql()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &InnerBuilderResult{
		DataQuery:  innerDataQuery,
		CountQuery: innerCountQuery,
		Args:       innerArgs,
	}, nil
}

// --------------------------------------------------------------------------------
//
// The below functions are helper functions for the "InQueryRebind" function

func appendReflectSlice(args []interface{}, v reflect.Value, vlen int) []interface{} {
	switch val := v.Interface().(type) {
	case []interface{}:
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

func asSliceForIn(i interface{}) (v reflect.Value, ok bool) {
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
	var args []interface{}

	if query, args, err = builder.ToSql(); err != nil {
		return nil, errors.WithStack(err)
	}

	resQuery := query
	resArgs := args

	if query, args, err = JetInQueryRebind(bindVar, query, args...); err != nil {
		return nil, errors.WithStack(fmt.Errorf("\n err: %s\n\n query: %s\n\n args: %v\n", err.Error(), resQuery, resArgs))
	}

	return db.QueryContext(ctx, query, args...)
}
