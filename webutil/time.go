package webutil

import (
	"time"

	pkgerrors "github.com/pkg/errors"
)

func convertTimeToLocalDateTime(dateString, dateFormat, timezone string) (time.Time, error) {
	location, err := time.LoadLocation(timezone)

	if err != nil {
		return time.Time{}, pkgerrors.Wrap(err, "")
	}

	parsedTime, err := time.Parse(dateFormat, dateString)

	if err != nil {
		return time.Time{}, pkgerrors.Wrap(err, "")
	}

	return parsedTime.In(location), nil
}

// ConvertTimeToLocalDateTime is used to convert the date string passed
// to the local time zone passed and returns a time instance
func ConvertTimeToLocalDateTime(dateString, timezone string) (time.Time, error) {
	location, err := time.LoadLocation(timezone)

	if err != nil {
		return time.Time{}, pkgerrors.Wrap(err, "")
	}

	parsedTime, err := time.Parse(POSTGRES_DATE_LAYOUT, dateString)

	if err != nil {
		return time.Time{}, pkgerrors.Wrap(err, "")
	}

	return parsedTime.In(location), nil
}

// ConvertTimeToLocalDateTimeF is used to convert the date string passed
// to the local time zone passed along with date format that will be retrieved
// from datasource and returns a time instance
func ConvertTimeToLocalDateTimeF(dateString, dateFormat, timezone string) (time.Time, error) {
	return convertTimeToLocalDateTime(dateString, dateFormat, timezone)
}

// GetCurrentLocalDateTimeInUTC will return the local date and time based on
// the time zone passed
func GetCurrentLocalDateTimeInUTC(timezone string) (time.Time, error) {
	return getUTC(timezone, true)
}

// GetCurrentLocalDateInUTC will return the local date based on
// the time zone passed
func GetCurrentLocalDateInUTC(timezone string) (time.Time, error) {
	return getUTC(timezone, false)
}

func getUTC(timezone string, includeTime bool) (time.Time, error) {
	var utcTime time.Time
	location, err := time.LoadLocation(timezone)

	if err != nil {
		return time.Time{}, pkgerrors.Wrap(err, "")
	}

	utc, err := time.LoadLocation("UTC")

	if err != nil {
		return time.Time{}, pkgerrors.Wrap(err, "")
	}

	localTime := time.Now().In(location)

	if includeTime {
		utcTime = time.Date(
			localTime.Year(),
			localTime.Month(),
			localTime.Day(),
			localTime.Hour(),
			localTime.Minute(),
			localTime.Second(),
			localTime.Nanosecond(),
			utc,
		)
	} else {
		utcTime = time.Date(
			localTime.Year(),
			localTime.Month(),
			localTime.Day(),
			0,
			0,
			0,
			0,
			utc,
		)
	}

	return utcTime, nil
}
