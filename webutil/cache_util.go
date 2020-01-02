package webutil

//go:generate mockgen -source=cache_util.go -destination=../webutilmock/cache_util_mock.go -package=webutilmock
//go:generate mockgen -source=cache_util.go -destination=cache_util_mock_test.go -package=webutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/sessions"
	"github.com/knq/snaker"
	redistore "gopkg.in/boj/redistore.v1"
)

//////////////////////////////////////////////////////////////////
//---------------------- CUSTOM ERRORS ------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrCacheNil is generic err indicating that cache backend came
	// back with nil
	ErrCacheNil = errors.New("webutil: cache is nil")

	errTooManyKeyArgs = errors.New("webutil: there are more cache args then columns")
)

//////////////////////////////////////////////////////////////////
//------------------------ INTERFACES --------------------------
//////////////////////////////////////////////////////////////////

// CacheStore is interface used to get, set and delete cached values
// from an in-memory database like redis
type CacheStore interface {
	Get(key string) ([]byte, error)
	Set(key string, value interface{}, expiration time.Duration)
	Del(keys ...string)
}

// SessionStore is interface used to implement server side sessions in an
// in-memory database like redis
type SessionStore interface {
	sessions.Store
	Ping() error
}

//////////////////////////////////////////////////////////////////
//---------------------- CONFIG STRUCTS ------------------------
//////////////////////////////////////////////////////////////////

// CacheConfig is used in form validation functions to see
// if given parameters can find cache results based on "Key" passed
type CacheConfig struct {
	// Cache is the cache backend to retrieve information
	Cache CacheStore

	// Key will be used against Cache to get value based on key
	Key string

	// IgnoreCacheNil will query database for information
	// even if cache returns nil
	IgnoreCacheNil bool
}

type CacheKey struct {
	Key string

	NumOfArgs int

	Expire time.Duration
}

type CacheSet struct {
	CacheKey CacheKey

	Query string

	QueryArgs []interface{}

	IsSingleKey bool
}

// CacheSetup is configuration struct used to setup caching database tables
// that generally do not insert/update often
//
// CacheSetup should be used in a map where the key value is the string name of
// the database table to cache and CacheSetup is the value to use for setting up cache
type CacheSetup struct {
	CacheSets []CacheSet

	CacheStore CacheStore
}

// CacheSetup is configuration struct used to setup caching database tables
// that generally do not insert/update often
//
// CacheSetup should be used in a map where the key value is the string name of
// the database table to cache and CacheSetup is the value to use for setting up cache
// type CacheSetup struct {
// 	// StringVal should be the "string" representation of the database table
// 	StringVal string

// 	// CacheIDKey should be the key value you will store the table id in cache
// 	CacheIDKey string

// 	// CacheListKey should be the key value you will store the whole table in cache
// 	CacheListKey string

// 	// OrderByColumn should determine what column to order by if passed
// 	OrderByColumn string

// 	CacheSelectionConf CacheSelectionConfig
// }

//////////////////////////////////////////////////////////////////
//------------------------- STRUCTS --------------------------
//////////////////////////////////////////////////////////////////

// RedisCache is default struct that implements the CacheStore interface
// The underlining implementation is based off of the
// "github.com/go-redis/redis" library
type RedisCache struct {
	CacheStore
}

// NewRedisCache returns pointer of RedisCache
func NewRedisCache(client CacheStore) *RedisCache {
	return &RedisCache{client}
}

// Get takes key value and determines if that key is in cache
func (c *RedisCache) Get(key string) ([]byte, error) {
	bytes, err := c.CacheStore.Get(key)

	if err == redis.ErrNil {
		return nil, ErrCacheNil
	}

	return bytes, err
}

// RedisSession is used for storing session variables
// in a Redis database
type RedisSession struct {
	*redistore.RediStore
}

// NewRedisSession returns new instance of *RedisSession
func NewRedisSession(r *redistore.RediStore) *RedisSession {
	return &RedisSession{RediStore: r}
}

// Ping verifies that the cache backend is still
// up and running
func (r *RedisSession) Ping() error {
	conn := r.RediStore.Pool.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	return err
}

//////////////////////////////////////////////////////////////////
//------------------------- FUNCTIONS --------------------------
//////////////////////////////////////////////////////////////////

// SetCacheFromDB takes the cacheSetup and loops through the
// configurations which queries the database and applies
// the results to cache
func SetCacheFromDB(cacheSetup CacheSetup, db Querier) error {
	for _, v := range cacheSetup.CacheSets {
		rows, err := db.Query(v.Query, v.QueryArgs...)

		if err != nil {
			return err
		}

		columns, err := rows.Columns()

		if err != nil {
			return err
		}

		colCount := len(columns)
		values := make([]interface{}, colCount)
		valuePtrs := make([]interface{}, colCount)
		rowList := make([]interface{}, 0)

		fmt.Printf("high val: %v\n", v.IsSingleKey)

		for rows.Next() {
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			err = rows.Scan(valuePtrs...)

			if err != nil {
				return err
			}

			row := make(map[string]interface{}, 0)

			for i := range columns {
				var val interface{}

				currentVal := values[i]

				switch currentVal.(type) {
				case int64:
					val = strconv.FormatInt(currentVal.(int64), IntBase)
				case *int64:
					t := currentVal.(*int64)
					if t != nil {
						val = strconv.FormatInt(*t, IntBase)
					}
				case []byte:
					t := val.([]byte)
					val, err = strconv.ParseFloat(string(t), IntBitSize)
					if err != nil {
						panic(err)
					}
				default:
					val = currentVal
				}

				columnName := snaker.ForceLowerCamelIdentifier(columns[i])
				row[columnName] = val
			}

			rowBytes, err := json.Marshal(&row)

			if err != nil {
				return err
			}

			fmt.Printf("val: %v\n", v.IsSingleKey)

			if !v.IsSingleKey {
				if colCount < v.CacheKey.NumOfArgs {
					return errTooManyKeyArgs
				}

				keyArgs := make([]interface{}, 0, v.CacheKey.NumOfArgs)

				for i := 0; i < v.CacheKey.NumOfArgs; i++ {
					keyArgs = append(keyArgs, columns[i])
				}

				key := fmt.Sprintf(v.CacheKey.Key, keyArgs...)
				cacheSetup.CacheStore.Set(key, rowBytes, v.CacheKey.Expire)
			} else {
				rowList = append(rowList, rowBytes)
			}
		}

		if v.IsSingleKey {
			rowListBytes, err := json.Marshal(&rowList)

			if err != nil {
				return err
			}

			cacheSetup.CacheStore.Set(v.CacheKey.Key, rowListBytes, v.CacheKey.Expire)
		}
	}

	return nil
}
