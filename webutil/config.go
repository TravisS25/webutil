package webutil

const (
	// DateTimeMilliLayout is global format for date time with adding milliseconds for precise calculations
	DateTimeMilliLayout = "2006-01-02 15:04:05.00000"

	// DateTimeLayout is global format for date time
	DateTimeLayout = "2006-01-02 15:04:05"

	// DateTimeOffsetLayout is global format for date time with time zone offset
	DateTimeOffsetLayout = "2006-01-02 15:04:05-07:00"

	// DateLayout is global format for date
	DateLayout = "2006-01-02"

	// PostgresDateLayout is date format used when receiving time from database
	PostgresDateLayout = "2006-01-02T15:04:05Z"

	// FormDateTimeLayout is format that should be received from a form
	FormDateTimeLayout = "01/02/2006 3:04 pm"

	// FormDateLayout is format that should be received from a form
	FormDateLayout = "01/02/2006"
)

const (
	// IntBase is default base to use for converting string to int64
	IntBase = 10

	// IntBitSize is default bit size to use for converting string to int64
	IntBitSize = 64
)

var (
	// True contains true value to use as reference variable
	// for bool pointers
	True = true

	// False contains false value to use as reference variable
	// for bool pointers
	False = false
)
