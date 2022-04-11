package webutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/TravisS25/webutil/webutilcfg"
	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/knq/snaker"
	redistore "gopkg.in/boj/redistore.v1"
)

var _ CacheStore = (*ClientCache)(nil)
var _ SessionStore = (*ClientSession)(nil)

//////////////////////////////////////////////////////////////////
//---------------------- CUSTOM ERRORS ------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrCacheNil is generic err indicating that cache backend came
	// back with nil
	ErrCacheNil = errors.New("webutil: cache is nil")

	// ErrTooManyPlaceHolders returns from SetCacheFromDB function
	// when the length of CacheKey#PlaceHolderPositions is more than
	// the number of columns passed
	ErrTooManyPlaceHolders = errors.New("webutil: there are more place holders then columns")

	// ErrPlaceholderOutOfRange returns from SetCacheFromDB function
	// when an index within the slice is out of range of the number
	// of columns returned from query
	ErrPlaceholderOutOfRange = errors.New("webutil: place holder out of range")
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

	// IgnoreCacheNil will query database for information
	// even if cache returns nil
	IgnoreCacheNil bool

	// Key will be used against Cache to get value based on key
	// This field could be optional depending on use case
	Key string
}

// CacheKey is config struct used to be apart of
// CacheSet to set global cache
type CacheKey struct {
	// Key used in cache
	Key string

	// PlaceHolderPositions should be the indexes to place
	// returned values from database to the formated key
	PlaceHolderPositions []int

	// Expire is Duration in which the key will stay
	// alive in cache
	// This is depended on cache backend and whether your
	// cache backend supports timed caching
	Expire time.Duration
}

// CacheSet is config struct used in CacheSetup config struct
// to set global cache
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

//////////////////////////////////////////////////////////////////
//--------------------------- TYPES --------------------------
//////////////////////////////////////////////////////////////////

// RecoverCache should implement ability to recover from cache error
type RecoverCache func(error) (*ClientCache, error)

// RetryCache should implement ability to take in CacheStore
// and query cache that was tried and recovered by RecoverCache
type RetryCache func(CacheStore) error

//////////////////////////////////////////////////////////////////
//------------------------- STRUCTS --------------------------
//////////////////////////////////////////////////////////////////

// ClientCache is default struct that implements the CacheStore interface
// The underlining implementation is based off of the
// "github.com/go-redis/redis" library
type ClientCache struct {
	*redis.Client
}

// NewClientCache returns pointer of ClientCache
func NewClientCache(client *redis.Client) *ClientCache {
	return &ClientCache{client}
}

// Get gets value based on key passed
// Returns error if key does not exist
func (c *ClientCache) Get(key string) ([]byte, error) {
	var resultsErr error

	results, err := c.Client.Get(key).Bytes()

	if err != nil {
		if err == redis.Nil {
			resultsErr = ErrCacheNil
		} else {
			resultsErr = err
		}
	}

	return results, resultsErr
}

// Set sets value in redis server based on key and value given
// Expiration sets how long the cache will stay in the server
// If 0, key/value will never be deleted
func (c *ClientCache) Set(key string, value interface{}, expiration time.Duration) {
	c.Client.Set(key, value, expiration)
}

// Del deletes given string array of keys from server if exists
func (c *ClientCache) Del(keys ...string) {
	c.Client.Del(keys...)
}

// ClientSession is used for storing session variables
// in a Redis database
type ClientSession struct {
	*redistore.RediStore
}

// NewClientSession returns new instance of *ClientSession
func NewClientSession(r *redistore.RediStore) *ClientSession {
	return &ClientSession{RediStore: r}
}

// Ping verifies that the cache backend is still
// up and running
func (r *ClientSession) Ping() error {
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
		rows, err := db.Queryx(v.Query, v.QueryArgs...)

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
					val = strconv.FormatInt(currentVal.(int64), webutilcfg.IntBase)
				case *int64:
					t := currentVal.(*int64)
					if t != nil {
						val = strconv.FormatInt(*t, webutilcfg.IntBase)
					}
				case []byte:
					t := val.([]byte)
					val, err = strconv.ParseFloat(string(t), webutilcfg.IntBitSize)
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

			if !v.IsSingleKey {
				if colCount < len(v.CacheKey.PlaceHolderPositions) {
					return ErrTooManyPlaceHolders
				}

				keyArgs := make([]interface{}, 0, len(v.CacheKey.PlaceHolderPositions))

				// Check to make sure user is not going out of index
				for i := 0; i < len(v.CacheKey.PlaceHolderPositions); i++ {
					idx := v.CacheKey.PlaceHolderPositions[i]

					if idx >= len(columns) {
						return ErrPlaceholderOutOfRange
					}

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

// HasCacheError determines if cache has error and it tries to recover
// if ServerErrorConfig#RecoverConfig#RecoverCache and retryCache are set
//
// Since cache is usually queried before a database and we can always resort
// to a database, we have the writeIfError parameter which indicates whether
// HasCacheError will write back to the client the server error set in
// ServerErrorConfig#ServerErrorResponse if cache is unable to recover
func HasCacheError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	retryCache RetryCache,
	writeIfError bool,
	conf ServerErrorConfig,
) bool {
	hasError := false

	if err != nil {
		logConf := LogConfig{CauseErr: err}

		if conf.RecoverCache != nil {
			if cache, recoverErr := conf.RecoverCache(err); err == nil {
				if retryCache != nil {
					if retryErr := retryCache(cache); retryErr != nil {
						logConf.RetryCacheErr = retryErr
						hasError = true
					}
				} else {
					hasError = true
				}
			} else {
				logConf.RecoverCacheErr = recoverErr
				hasError = true
			}
		} else {
			hasError = true
		}

		if conf.Logger != nil {
			conf.Logger(r, logConf)
		}

		if writeIfError && hasError {
			SetHTTPResponseDefaults(
				&conf.ServerErrorResponse,
				http.StatusInternalServerError,
				[]byte(serverErrTxt),
			)

			w.WriteHeader(*conf.ServerErrorResponse.HTTPStatus)
			w.Write(conf.ServerErrorResponse.HTTPResponse)
		}
	}

	return hasError
}
