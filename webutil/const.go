package webutil

//////////////////////////////////////////////////////////////////
//---------------------- VALIDATOR TYPES -----------------------
//////////////////////////////////////////////////////////////////

const (
	validateArgsType = iota + 1
	validateUniquenessType
	validateExistsType
)

//////////////////////////////////////////////////////////////////
//---------------------- DATE LAYOUTS --------------------------
//////////////////////////////////////////////////////////////////

const (
	// DATE_TIME_MILLI_LAYOUT is global format for date time with adding milliseconds for precise calculations
	DATE_TIME_MILLI_LAYOUT = "2006-01-02 15:04:05.00000"

	// DATE_TIME_LAYOUT is global format for date time
	DATE_TIME_LAYOUT = "2006-01-02 15:04:05"

	// DATE_TIME_OFFSET_LAYOUT is global format for date time with time zone offset
	DATE_TIME_OFFSET_LAYOUT = "2006-01-02 15:04:05-07:00"

	// DATE_LAYOUT is global format for date
	DATE_LAYOUT = "2006-01-02"

	// POSTGRES_DATE_LAYOUT is date format used when receiving time from database
	POSTGRES_DATE_LAYOUT = "2006-01-02T15:04:05Z"

	// FORMDATE_TIME_LAYOUT is format that should be received from a form
	FORM_DATE_TIME_LAYOUT = "01/02/2006 3:04 pm"

	// FORM_DATE_LAYOUT is format that should be received from a form
	FORM_DATE_LAYOUT = "01/02/2006"

	// FORM_DATE_TIME_STRF is format used when formatting date in sql query
	// to use a form format
	FORM_DATE_TIME_STRF = "%m/%d/%Y %I:%M %p"

	// FORM_DATE_STRF is format used when formatting date in sql query
	// to use a form format
	FORM_DATE_STRF = "%m/%d/%Y"
)

//////////////////////////////////////////////////////////////////
//---------------------- ERROR TEXT ---------------------------
//////////////////////////////////////////////////////////////////

const (
	// REQUIRED_TXT is string const error when field is required
	REQUIRED_TXT = "required"

	// ALREADY_EXISTS_TXT is string const error when field already exists
	// in database or cache
	ALREADY_EXISTS_TXT = "already exists"

	// DOES_NOT_EXIST_TXT is string const error when field does not exist
	// in database or cache
	DOES_NOT_EXIST_TXT = "does not exist"

	// INVALID_TXT is string const error when field is invalid
	INVALID_TXT = "invalid"

	// INVALID_FORMAT_TXT is string const error when field has invalid format
	INVALID_FORMAT_TXT = "invalid format"

	// INVALID_FUTURE_DATE_TXT is sring const when field is not allowed
	// to be in the future
	INVALID_FUTURE_DATE_TXT = "date can't be after current date/time"

	// INVALID_PAST_DATE_TXT is sring const when field is not allowed
	// to be in the past
	INVALID_PAST_DATE_TXT = "date can't be before current date/time"

	// CANT_BE_NEGATIVE_TXT is sring const when field can't be negative
	CANT_BE_NEGATIVE_TXT = "can't be negative"
)

//////////////////////////////////////////////////////////////////
//---------------------- EMPTY VALUES ---------------------------
//////////////////////////////////////////////////////////////////

const (
	EMPTY_UUID = "00000000-0000-0000-0000-000000000000"

	EMPTY_TIME = "0001-01-01 00:00:00 +0000 UTC"
)

//////////////////////////////////////////////////////////////////
//---------------------- SQL BIND VARS ---------------------------
//////////////////////////////////////////////////////////////////

const (
	UNKNOWN_SQL_BIND_VAR = iota

	// QUESTION_SQL_BIND_VAR is bind var for mysql and other databases
	QUESTION_SQL_BIND_VAR

	// DOLLAR_SQL_BIND_VAR is bind var for mainly for postgres database
	DOLLAR_SQL_BIND_VAR
	NAMED_SQL_BIND_VAR
	AT_SQL_BIND_VAR
)

//////////////////////////////////////////////////////////////////
//------------------------ SSL MODES ---------------------------
//////////////////////////////////////////////////////////////////

const (
	// SSL_DISABLE_MODE represents disable value for "sslmode" query parameter
	SSL_DISABLE_MODE = "disable"

	// SSL_REQUIRE_MODE represents require value for "sslmode" query parameter
	SSL_REQUIRE_MODE = "require"

	// SSL_VERIFY_CA_MODE represents verify-ca value for "sslmode" query parameter
	SSL_VERIFY_CA_MODE = "verify-ca"

	// SSL_VERIFY_FULL_MODE represents verify-full value for "sslmode" query parameter
	SSL_VERIFY_FULL_MODE = "verify-full"
)

//////////////////////////////////////////////////////////////////
//---------------------- DRIVERS ------------------------
//////////////////////////////////////////////////////////////////

const (
	// POSTGRES_DRIVER is protocol string for postgres database
	POSTGRES_DRIVER = "postgres"

	// MYSQL_DRIVER is protocol string for mysql database
	MYSQL_DRIVER = "mysql"

	// SQLITE_DRIVER is protocol string for sqlite database
	SQLITE_DRIVER = "sqlite"
)

//////////////////////////////////////////////////////////////////
//------------------------ HTTP HEADERS ----------------------
//////////////////////////////////////////////////////////////////

const (
	// BINARY_CONTENT_HEADER is key string for content type header "application/octet-stream"
	BINARY_CONTENT_HEADER = "application/octet-stream"

	// FORM_CONTENT_HEADER is key string for content type header "application/x-www-form-urlencoded"
	FORM_CONTENT_HEADER = "application/x-www-form-urlencoded"

	// JSON_CONTENT_HEADER is key string for content type header "application/json"
	JSON_CONTENT_HEADER = "application/json"

	// PDF_CONTENT_HEADER is key string for content type header "application/pdf"
	PDF_CONTENT_HEADER = "application/pdf"

	// ZIP_CONTENT_HEADER is key string for content type header "application/zip"
	ZIP_CONTENT_HEADER = "application/zip"

	// HTML_CONTENT_HEADER is key string for content type header "text/html; charset=utf-8"
	HTML_CONTENT_HEADER = "text/html; charset=utf-8"

	// TEXT_CONTENT_HEADER is key string for content type header "text/plain; charset=utf-8"
	TEXT_CONTENT_HEADER = "text/plain; charset=utf-8"

	// CSV_CONTENT_HEADER is key string for content type header "text/csv; charset=utf-8"
	CSV_CONTENT_HEADER = "text/csv; charset=utf-8"

	// JPG_CONTENT_HEADER is key string for content type header "image/jpeg"
	JPG_CONTENT_HEADER = "image/jpeg"

	// PNG_CONTENT_HEADER is key string for content type header "image/png"
	PNG_CONTENT_HEADER = "image/png"

	// TOKEN_HEADER is key string for "X-CSRF-TOKEN" header
	TOKEN_HEADER = "X-Csrf-Token"

	// COOKIE_HEADER is key string for "Cookie" header
	COOKIE_HEADER = "Cookie"

	// SET_COOKIE_HEADER is key string for "Set-Cookie" header
	SET_COOKIE_HEADER = "Set-Cookie"

	// LOCATION_HEADER is key string for "Location" header
	LOCATION_HEADER = "Location"
)

//////////////////////////////////////////////////////////////////
//-------------------------- REGEX ----------------------------
//////////////////////////////////////////////////////////////////

const (
	ID_REGEX_STR   = "[0-9]+"
	UUID_REGEX_STR = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"

	ID_PARAM   = "{id:" + ID_REGEX_STR + "}"
	UUID_PARAM = "{id:" + UUID_REGEX_STR + "}"
)

//////////////////////////////////////////////////////////////////
//--------------------------- MISC ----------------------------
//////////////////////////////////////////////////////////////////

const (
	// DB_CONN_STR is default format for a connection string to a database
	DB_CONN_STR = "%s://%s:%s@%s:%d/%s?&sslmode=%s&sslrootcert=%s&sslkey=%s&sslcert=%s&search_path=%s"

	// INT_BASE is default base to use for converting string to int64
	INT_BASE = 10

	// INT_BIT_SIZE is default bit size to use for converting string to int64
	INT_BIT_SIZE = 64
)
