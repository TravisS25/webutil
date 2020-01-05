// Code generated by mockery v1.0.0. DO NOT EDIT.

package webutil

import (
	io "io"

	minio "github.com/minio/minio-go"
	mock "github.com/stretchr/testify/mock"

	time "time"

	url "net/url"
)

// MockStorageReaderWriter is an autogenerated mock type for the StorageReaderWriter type
type MockStorageReaderWriter struct {
	mock.Mock
}

// GetObject provides a mock function with given fields: bucketName, objectName, opts
func (_m *MockStorageReaderWriter) GetObject(bucketName string, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	ret := _m.Called(bucketName, objectName, opts)

	var r0 *minio.Object
	if rf, ok := ret.Get(0).(func(string, string, minio.GetObjectOptions) *minio.Object); ok {
		r0 = rf(bucketName, objectName, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*minio.Object)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, minio.GetObjectOptions) error); ok {
		r1 = rf(bucketName, objectName, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PresignedGetObject provides a mock function with given fields: bucketName, objectName, expiry, reqParams
func (_m *MockStorageReaderWriter) PresignedGetObject(bucketName string, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
	ret := _m.Called(bucketName, objectName, expiry, reqParams)

	var r0 *url.URL
	if rf, ok := ret.Get(0).(func(string, string, time.Duration, url.Values) *url.URL); ok {
		r0 = rf(bucketName, objectName, expiry, reqParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*url.URL)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, time.Duration, url.Values) error); ok {
		r1 = rf(bucketName, objectName, expiry, reqParams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutObject provides a mock function with given fields: bucketName, objectName, reader, objectSize, opts
func (_m *MockStorageReaderWriter) PutObject(bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (int64, error) {
	ret := _m.Called(bucketName, objectName, reader, objectSize, opts)

	var r0 int64
	if rf, ok := ret.Get(0).(func(string, string, io.Reader, int64, minio.PutObjectOptions) int64); ok {
		r0 = rf(bucketName, objectName, reader, objectSize, opts)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, io.Reader, int64, minio.PutObjectOptions) error); ok {
		r1 = rf(bucketName, objectName, reader, objectSize, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveObject provides a mock function with given fields: bucketName, objectName
func (_m *MockStorageReaderWriter) RemoveObject(bucketName string, objectName string) error {
	ret := _m.Called(bucketName, objectName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(bucketName, objectName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
