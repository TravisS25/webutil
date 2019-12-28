package webutil

import (
	"testing"

	"github.com/go-redis/redis"
	gomock "github.com/golang/mock/gomock"
	redistore "gopkg.in/boj/redistore.v1"
)

func TestRedisCacheUnitTest(t *testing.T) {
	var err error

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRedis := NewMockCacheStore(mockCtrl)
	redisCache := NewRedisCache(mockRedis)

	key := "foo"
	r := []byte("hello")

	mockRedis.EXPECT().Get(key).Return(r, nil)
	b, _ := redisCache.HasKey(key)

	if !b {
		t.Fatalf("should have key")
	}

	mockRedis.EXPECT().Get(key).Return(nil, redis.Nil)
	_, err = redisCache.Get(key)

	if err != redis.Nil {
		t.Fatalf("cache should return nil\n")
	}

	mockRedis.EXPECT().Get(key).Return(nil, ErrServer)
	_, err = redisCache.Get(key)

	if err != ErrServer {
		t.Fatalf("cache should return server error\n")
	}

	mockRedis.EXPECT().Get(key).Return(nil, ErrServer)
	b, _ = redisCache.HasKey(key)

	if b {
		t.Fatalf("should NOT have key")
	}
}

func TestRedisSessionIntegrationTest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store, err := redistore.NewRediStore(
		10,
		"tcp",
		"localhost:6379",
		"",
		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
	)

	if err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}

	rs := NewRedisSession(store)

	if err = rs.Ping(); err != nil {
		t.Fatalf("err: %s\n", err.Error())
	}
}
