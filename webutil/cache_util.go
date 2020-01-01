package webutil

//go:generate mockgen -source=cache_util.go -destination=../webutilmock/cache_util_mock.go -package=webutilmock
//go:generate mockgen -source=cache_util.go -destination=cache_util_mock_test.go -package=webutil

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/sessions"
	redistore "gopkg.in/boj/redistore.v1"
)

//////////////////////////////////////////////////////////////////
//---------------------- CUSTOM ERRORS ------------------------
//////////////////////////////////////////////////////////////////

var (
	// ErrCacheNil is generic err indicating that cache backend came
	// back with nil
	ErrCacheNil = errors.New("webutil: cache is nil")
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
	//HasKey(key string) (bool, error)
	//HasKey(key string) error
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

// CacheSelectionConfig is config struct used for setting queried
// results from database to cache
type CacheSelectionConfig struct {
	TextColumn       string
	ValueColumn      string
	FormSelectionKey string
}

type CacheSet struct {
	Key string

	KeyValues []interface{}

	Expire time.Duration
}

// // CacheSetup is configuration struct used to setup caching database tables
// // that generally do not insert/update often
// //
// // CacheSetup should be used in a map where the key value is the string name of
// // the database table to cache and CacheSetup is the value to use for setting up cache
// type CacheSetup struct {
// 	// StringVal should be the "string" representation of the database table
// 	StringVal string

// 	// OrderByColumn should determine what column to order by if passed
// 	OrderByColumn string

// 	CacheSelectionConf CacheSelectionConfig
// }

// CacheSetup is configuration struct used to setup caching database tables
// that generally do not insert/update often
//
// CacheSetup should be used in a map where the key value is the string name of
// the database table to cache and CacheSetup is the value to use for setting up cache
type CacheSetup struct {
	// StringVal should be the "string" representation of the database table
	StringVal string

	// CacheIDKey should be the key value you will store the table id in cache
	CacheIDKey string

	// CacheListKey should be the key value you will store the whole table in cache
	CacheListKey string

	// OrderByColumn should determine what column to order by if passed
	OrderByColumn string

	CacheSelectionConf CacheSelectionConfig
}

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
