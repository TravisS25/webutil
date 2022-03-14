package webutil

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	testifymock "github.com/stretchr/testify/mock"
)

// func TestRedisCacheUnitTest(t *testing.T) {
// 	var err error

// 	mockCacheStore := &MockCacheStore{}
// 	defer mockCacheStore.AssertExpectations(t)

// 	key := "foo"
// 	//r := []byte("hello")

// 	// mockRedis.EXPECT().Get(key).Return(r, nil)
// 	// b, _ := redisCache.HasKey(key)

// 	// if !b {
// 	// 	t.Fatalf("should have key")
// 	// }

// 	mockRedis.EXPECT().Get(key).Return(nil, redis.Nil)
// 	_, err = redisCache.Get(key)

// 	if err != redis.Nil {
// 		t.Fatalf("cache should return nil\n")
// 	}

// 	mockRedis.EXPECT().Get(key).Return(nil, ErrServer)
// 	_, err = redisCache.Get(key)

// 	if err != ErrServer {
// 		t.Fatalf("cache should return server error\n")
// 	}

// 	// mockRedis.EXPECT().Get(key).Return(nil, ErrServer)
// 	// b, _ = redisCache.HasKey(key)

// 	// if b {
// 	// 	t.Fatalf("should NOT have key")
// 	// }
// }

// func TestRedisCacheUnitTest(t *testing.T) {
// 	var err error

// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()

// 	mockRedis := NewMockCacheStore(mockCtrl)
// 	redisCache := NewRedisCache(mockRedis)

// 	key := "foo"
// 	//r := []byte("hello")

// 	// mockRedis.EXPECT().Get(key).Return(r, nil)
// 	// b, _ := redisCache.HasKey(key)

// 	// if !b {
// 	// 	t.Fatalf("should have key")
// 	// }

// 	mockRedis.EXPECT().Get(key).Return(nil, redis.Nil)
// 	_, err = redisCache.Get(key)

// 	if err != redis.Nil {
// 		t.Fatalf("cache should return nil\n")
// 	}

// 	mockRedis.EXPECT().Get(key).Return(nil, ErrServer)
// 	_, err = redisCache.Get(key)

// 	if err != ErrServer {
// 		t.Fatalf("cache should return server error\n")
// 	}

// 	// mockRedis.EXPECT().Get(key).Return(nil, ErrServer)
// 	// b, _ = redisCache.HasKey(key)

// 	// if b {
// 	// 	t.Fatalf("should NOT have key")
// 	// }
// }

// func TestRedisSessionIntegrationTest(t *testing.T) {
// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()

// 	store, err := redistore.NewRediStore(
// 		10,
// 		"tcp",
// 		"localhost:6379",
// 		"",
// 		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
// 		[]byte("fF832S1flhmd6fdl5BgmbkskghmawQP3"),
// 	)

// 	if err != nil {
// 		t.Fatalf("err: %s\n", err.Error())
// 	}

// 	rs := NewClientSession(store)

// 	if err = rs.Ping(); err != nil {
// 		t.Fatalf("err: %s\n", err.Error())
// 	}
// }

// func TestSetCachingUnitTest(t *testing.T) {
// 	var err error

// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()

// 	mockCacheStore := NewMockCacheStore(mockCtrl)
// 	mockCacheStore.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

// 	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

// 	if err != nil {
// 		t.Fatalf("fatal err: %s\n", err.Error())
// 	}

// 	rows := sqlmock.NewRows([]string{"value", "text"}).
// 		AddRow(1, "foo").
// 		AddRow(2, "bar")
// 	mockDB.ExpectQuery("").WillReturnRows(rows)
// 	setup := CacheSetup{
// 		CacheStore: mockCacheStore,
// 		CacheSets: []CacheSet{
// 			{
// 				CacheKey: CacheKey{
// 					Key:       "key-%v",
// 					NumOfArgs: 1,
// 					Expire:    0,
// 				},
// 				Query:       "select",
// 				IsSingleKey: true,
// 			},
// 		},
// 	}

// 	if err = SetCacheFromDB(setup, db); err != nil {
// 		t.Errorf("should not have error\n")
// 		t.Errorf("err: %s\n", err.Error())
// 	}

// 	rows = sqlmock.NewRows([]string{"value", "text"}).
// 		AddRow(1, "foo").
// 		AddRow(2, "bar")
// 	mockDB.ExpectQuery("").WillReturnRows(rows)
// 	setup.CacheSets[0].IsSingleKey = false

// 	if err = SetCacheFromDB(setup, db); err != nil {
// 		t.Errorf("should not have error\n")
// 		t.Errorf("err: %s\n", err.Error())
// 	}

// 	rows = sqlmock.NewRows([]string{"value", "text"}).
// 		AddRow(1, "foo").
// 		AddRow(2, "bar")
// 	mockDB.ExpectQuery("").WillReturnRows(rows)
// 	setup.CacheSets[0].CacheKey.NumOfArgs = 3

// 	if err = SetCacheFromDB(setup, db); err == nil {
// 		t.Errorf("should have error\n")
// 	} else {
// 		if err != errTooManyKeyArgs {
// 			t.Errorf("should have errTooManyKeyArgs error\n")
// 			t.Errorf("err: %s\n", err.Error())
// 		}
// 	}
// }

func TestSetCachingUnitTest(t *testing.T) {
	var err error

	mockCacheStore := &MockCacheStore{}
	defer mockCacheStore.AssertExpectations(t)

	mockCacheStore.On("Set", testifymock.Anything, testifymock.Anything, testifymock.Anything)

	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlAnyMatcher))

	if err != nil {
		t.Fatalf("fatal err: %s\n", err.Error())
	}

	rows := sqlmock.NewRows([]string{"value", "text"}).
		AddRow(1, "foo").
		AddRow(2, "bar")
	mockDB.ExpectQuery("").WillReturnRows(rows)
	setup := CacheSetup{
		CacheStore: mockCacheStore,
		CacheSets: []CacheSet{
			{
				CacheKey: CacheKey{
					Key:                  "key-%v",
					PlaceHolderPositions: []int{1},
					Expire:               0,
				},
				Query:       "select",
				IsSingleKey: true,
			},
		},
	}

	newDB := &sqlx.DB{
		DB: db,
	}

	if err = SetCacheFromDB(setup, newDB); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}

	rows = sqlmock.NewRows([]string{"value", "text"}).
		AddRow(1, "foo").
		AddRow(2, "bar")
	mockDB.ExpectQuery("").WillReturnRows(rows)
	setup.CacheSets[0].IsSingleKey = false

	if err = SetCacheFromDB(setup, newDB); err != nil {
		t.Errorf("should not have error\n")
		t.Errorf("err: %s\n", err.Error())
	}
}

func TestHasCacheErrorUnitTest(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/url", nil)

	if !HasCacheError(rr, req, errors.New("error"), nil, true, ServerErrorConfig{}) {
		t.Errorf("should have error\n")
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(rr.Result().Body)
	rr.Result().Body.Close()

	if rr.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("status code should be 500\n")
	}

	if buf.String() != serverErrTxt {
		t.Errorf("error response should be : %s\n", serverErrTxt)
	}

	// ------------------------------------------------------------------------------

	buf.Reset()
	rr = httptest.NewRecorder()
	sec := ServerErrorConfig{
		Logger: func(r *http.Request, conf LogConfig) {
			if conf.RecoverCacheErr == nil {
				t.Errorf("cache recover error should NOT be nil\n")
			}
		},
		RecoverConfig: RecoverConfig{
			RecoverCache: func(err error) (*ClientCache, error) {
				return nil, errors.New("errors")
			},
		},
	}

	if !HasCacheError(rr, req, errors.New("errors"), nil, true, sec) {
		t.Errorf("should have error\n")
	}

	buf.ReadFrom(rr.Result().Body)
	rr.Result().Body.Close()

	if rr.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("status code should be 500\n")
	}

	if buf.String() != serverErrTxt {
		t.Errorf("error response should be : %s\n", serverErrTxt)
	}

	// ------------------------------------------------------------------------------

	buf.Reset()
	rr = httptest.NewRecorder()
	sec.RecoverCache = func(err error) (*ClientCache, error) {
		return &ClientCache{}, nil
	}
	sec.Logger = func(r *http.Request, conf LogConfig) {
		if conf.RetryCacheErr != nil {
			t.Errorf("retry cache error should NOT be nil\n")
		}
	}

	if !HasCacheError(rr, req, errors.New("errors"), func(cache CacheStore) error {
		return errors.New("errors")
	}, true, sec) {
		t.Errorf("should have error\n")
	}

	buf.ReadFrom(rr.Result().Body)
	rr.Result().Body.Close()

	if rr.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("status code should be 500\n")
	}

	if buf.String() != serverErrTxt {
		t.Errorf("error response should be : %s\n", serverErrTxt)
	}
}
