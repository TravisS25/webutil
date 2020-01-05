#!/bin/bash

mockery -dir webutil -all -inpkg -testonly
mockery -dir webutil -all 

mockery -dir $GOPATH/src/github.com/gorilla/securecookie/ -name Error -output webutil -outpkg webutil -testonly
mockery -dir $GOPATH/src/github.com/gorilla/securecookie/ -name Error

mockery -dir /usr/local/go/src/net/http/ -name Handler -output webutil -outpkg webutil -testonly 
mockery -dir /usr/local/go/src/net/http/ -name Handler

mockery -dir $GOPATH/src/github.com/gorilla/sessions/ -name Store -output webutil -outpkg webutil -testonly
mockery -dir $GOPATH/src/github.com/gorilla/sessions/ -name Store