module github.com/TravisS25/webutil

go 1.13

require (
	github.com/DATA-DOG/go-sqlmock v1.4.0
	github.com/Masterminds/squirrel v1.5.3
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/go-jet/jet/v2 v2.10.1
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/google/uuid v1.4.0
	github.com/gorilla/csrf v1.6.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.10.8
	github.com/mitchellh/mapstructure v1.5.0
	github.com/nqd/flat v0.2.0
	github.com/pkg/errors v0.9.0
	github.com/sanity-io/litter v1.5.5
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/viper v1.6.1
	github.com/stretchr/objx v0.5.0
	github.com/stretchr/testify v1.8.2
)

replace github.com/jmoiron/sqlx => /home/travis/programming/open-source/go/sqlx
