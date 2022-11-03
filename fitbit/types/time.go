// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"fmt"
	"strings"
	"time"
)

type (
	// Custom type for the DateTime fields returned by the Fitbit API
	FitbitDateTime struct{ time.Time }

	// Custom type for the Date (without Time) fields returned by the Fitbit API
	FitbitDate struct{ time.Time }

	// Custom type for the Time (without Date) fields returned by the Fitbit API
	FitbitTime struct{ time.Time }

	Period string
)

const (
	DateTimeLayout = "2006-01-02T15:04"
	DateLayout     = "2006-01-02"
	TimeLayout     = "15:04"

	// 1d | 7d | 30d | 1w | 1m | 3m | 6m | 1y
	Period1Day    Period = "1d"
	Period7Days   Period = "7d"
	Period30Days  Period = "30d"
	Period1Week   Period = "1w"
	Period1Month  Period = "1m"
	Period3Months Period = "3m"
	Period6Months Period = "6m"
	Period1Year   Period = "1y"
)

func (d *FitbitDateTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		d.Time = time.Time{}
		return
	}
	// First try with the custom layout
	if d.Time, err = time.Parse(DateTimeLayout, s); err != nil {
		// In case of error, try with the standard parsing of format
		// time.RFC3339
		d.Time, err = time.Parse(time.RFC3339, s)
	}
	return
}

func (d *FitbitDateTime) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Format(DateTimeLayout))), nil
}

func (d *FitbitDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		d.Time = time.Time{}
		return
	}
	d.Time, err = time.Parse(DateLayout, s)
	return
}

func (d *FitbitDate) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Format(DateLayout))), nil
}

func (d *FitbitTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		d.Time = time.Time{}
		return
	}
	d.Time, err = time.Parse(TimeLayout, s)
	return
}

func (d *FitbitTime) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Format(TimeLayout))), nil
}
