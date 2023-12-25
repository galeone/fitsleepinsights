// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"strconv"
	"time"

	pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/galeone/fitbit/v2/types"
)

type HeartRateVariabilityTimeSeries struct {
	types.HeartRateVariabilityTimeStep
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
	// Value is a nested struct. Ignore and add fields
	Value      float64 `sql:"-"`
	DailyRmssd float64
	DeepRmssd  float64
	DateTime   time.Time `sql:"-"` // it's a date
	Date       time.Time
}

func (HeartRateVariabilityTimeSeries) Headers() []string {
	return []string{
		"DailyRmssd",
		"DeepRmssd",
	}
}

func (f *HeartRateVariabilityTimeSeries) Values() []string {
	return []string{
		strconv.FormatFloat(f.DailyRmssd, 'f', 2, 64),
		strconv.FormatFloat(f.DeepRmssd, 'f', 2, 64),
	}
}

func (HeartRateVariabilityTimeSeries) TableName() string {
	return "heart_rate_variability_time_series"
}
