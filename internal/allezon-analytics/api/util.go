package api

import (
	"fmt"
	"strings"
	"time"
)

var datetimeWithZoneFormats = []string{"2006-01-02T15:04:05Z", "2006-01-02T15:04:05.0Z", "2006-01-02T15:04:05.00Z", "2006-01-02T15:04:05.000Z"}
var datetimeFormats = []string{"2006-01-02T15:04:05", "2006-01-02T15:04:05.0", "2006-01-02T15:04:05.00", "2006-01-02T15:04:05.000"}

func ParseDatetimeWithZone(s string) (time.Time, error) {
	var err error
	var res time.Time

	for _, f := range datetimeWithZoneFormats {
		res, err = time.Parse(f, s)
		if err == nil {
			break
		}
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid datetime with zone format: %s", s)
	}

	return res, nil
}

func ParseDatetime(s string) (time.Time, error) {
	var err error
	var res time.Time

	for _, f := range datetimeFormats {
		res, err = time.Parse(f, s)
		if err == nil {
			break
		}
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid datetime format: %s", s)
	}

	return res, nil
}

func ParseTimeRange(s string) (time.Time, time.Time, error) {
	var err error

	trs := strings.Split(s, "_")
	if len(trs) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range format")
	}

	lowerBound, err := ParseDatetime(trs[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid lower bound")
	}

	upperBound, err := ParseDatetime(trs[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid upper bound")
	}

	return lowerBound, upperBound, nil
}
