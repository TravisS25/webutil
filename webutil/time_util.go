package webutil

import (
	"fmt"
	"time"

	pkgerrors "github.com/pkg/errors"
)

// ConvertTimeToLocalDateTime is used to convert the date string passed
// to the local time zone passed and returns a time instance
func ConvertTimeToLocalDateTime(dateString, timezone string) (time.Time, error) {
	location, err := time.LoadLocation(timezone)

	if err != nil {
		return time.Time{}, pkgerrors.Wrap(err, "")
	}

	parsedTime, err := time.Parse(PostgresDateLayout, dateString)

	if err != nil {
		return time.Time{}, pkgerrors.Wrap(err, "")
	}

	return parsedTime.In(location), nil
}

// func GetCurrentDateTimeInUTC() time.Time {
// 	currentDate := time.Now()
// 	year := strconv.Itoa(currentDate.Year())
// 	month := fmt.Sprintf("%02d", currentDate.Month())
// 	day := fmt.Sprintf("%02d", currentDate.Day())
// 	currentDateString := year + "-" + month + "-" + day
// 	currentUTCDate, _ := time.Parse(DateLayout, currentDateString)
// 	return currentUTCDate
// }

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
		fmt.Printf("foo err: %s\n", err.Error())
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
