#!/bin/bash

# This script is used to generate mocks so that the webutil package can use mocks to test
# internal functions but also generate a mocks package itself so a user could import
# the mocks library to test in their own application

mockery --dir webutil --all --testonly --inpackage
mockery --dir webutil --all

mockery --dir=/usr/local/go/src/net/http/ --name Handler --output webutil --outpkg webutil
mockery --dir=/usr/local/go/src/net/http/ --name Handler

mockery --dir $GOPATH/pkg/mod/github.com/gorilla/sessions@v1.2.0 --name Error --output webutil --outpkg webutil
mockery --dir $GOPATH/pkg/mod/github.com/gorilla/sessions@v1.2.0 --name Error

mockery --dir=$GOPATH/pkg/mod/github.com/gorilla/sessions@v1.2.0 --name Store --output webutil --outpkg webutil
mockery --dir=$GOPATH/pkg/mod/github.com/gorilla/sessions@v1.2.0 --name Store

mockery --dir webutiltest --all --testonly --inpackage