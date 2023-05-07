// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"time"

	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type HeartRateActivities struct {
	types.HeartRateActivities
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
	// Value is a struct containing an array and a field (resting heart rate).
	// The array is a cardio_zone (handled separately), here we ignore value and add
	// Directly the fields (array, to fill with scan queries) here.
	// All the fields are ignored, but we'll handle them manually (in this way igor ignores them).
	Value                int64           `sql:"-"`
	CustomHeartRateZones []HeartRateZone `sql:"-"`
	HeartRateZones       []HeartRateZone `sql:"-"`
	RestingHeartRate     int64
	DateTime             types.FitbitDate `sql:"-"` // It's a Date
	Date                 time.Time
}

func (HeartRateActivities) TableName() string {
	return "heart_rate_activities"
}
