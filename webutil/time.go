package webutil

import (
	"time"

	"github.com/pkg/errors"
)

// ConvertToTimezone takes in date string along with timezone and returns
// the same time clock but with given timezone, meaning that the time
// actually changes since we keep the clock the same
func ConvertToTimezone(value time.Time, timezone string, includeTime bool) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "")
	}

	var newTime time.Time

	if includeTime {
		newTime = time.Date(
			value.Year(),
			value.Month(),
			value.Day(),
			value.Hour(),
			value.Minute(),
			value.Second(),
			value.Nanosecond(),
			loc,
		)
	} else {
		newTime = time.Date(
			value.Year(),
			value.Month(),
			value.Day(),
			0,
			0,
			0,
			0,
			loc,
		)
	}

	return newTime, nil
}

// // GetCurrentLocalDateTimeInUTC will return the local date and time based on
// // the time zone passed
// func GetCurrentLocalDateTimeInUTC(timezone string) (time.Time, error) {
// 	return getUTC(timezone, true)
// }

// // GetCurrentLocalDateInUTC will return the local date based on
// // the time zone passed
// func GetCurrentLocalDateInUTC(timezone string) (time.Time, error) {
// 	return getUTC(timezone, false)
// }

// func getUTC(timezone string, includeTime bool) (time.Time, error) {
// 	var utcTime time.Time
// 	location, err := time.LoadLocation(timezone)

// 	if err != nil {
// 		return time.Time{}, errors.Wrap(err, "")
// 	}

// 	utc, err := time.LoadLocation("UTC")

// 	if err != nil {
// 		return time.Time{}, errors.Wrap(err, "")
// 	}

// 	localTime := time.Now().In(location)

// 	if includeTime {
// 		utcTime = time.Date(
// 			localTime.Year(),
// 			localTime.Month(),
// 			localTime.Day(),
// 			localTime.Hour(),
// 			localTime.Minute(),
// 			localTime.Second(),
// 			localTime.Nanosecond(),
// 			utc,
// 		)
// 	} else {
// 		utcTime = time.Date(
// 			localTime.Year(),
// 			localTime.Month(),
// 			localTime.Day(),
// 			0,
// 			0,
// 			0,
// 			0,
// 			utc,
// 		)
// 	}

// 	return utcTime, nil
// }
