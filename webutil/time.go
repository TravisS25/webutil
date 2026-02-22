package webutil

import (
	"time"

	"github.com/pkg/errors"
)

// ConvertToTimezone takes in date string along with timezone and returns
// the same time clock but with given timezone, meaning that the time
// actually changes since we keep the clock the same
func ConvertToTimezone(value time.Time, timezone string, includeTime bool) (time.Time, error) {
	if timezone == "" {
		timezone = "UTC"
	}

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
