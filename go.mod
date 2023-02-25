module github.com/TravisS25/webutil

go 1.13

require (
	github.com/DATA-DOG/go-sqlmock v1.4.0
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/gorilla/csrf v1.6.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/knq/snaker v0.0.0-20181215144011-2bc8a4db4687
	github.com/lib/pq v1.3.0
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/pkg/errors v0.9.0
	github.com/shopspring/decimal v1.2.0
	github.com/spf13/viper v1.6.1
	github.com/stretchr/objx v0.3.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	gopkg.in/boj/redistore.v1 v1.0.0-20160128113310-fc113767cd6b
	gopkg.in/yaml.v2 v2.2.7 // indirect
)

replace github.com/jmoiron/sqlx => /home/travis/programming/go/src/github.com/jmoiron/sqlx
