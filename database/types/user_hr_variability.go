// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"time"

	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
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

func (HeartRateVariabilityTimeSeries) TableName() string {
	return "heart_rate_variability_time_series"
}
