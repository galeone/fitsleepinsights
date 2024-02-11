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

type BreathingRate struct {
	types.BreathingRateTimePoint
	ID       int64               `igor:"primary_key"`
	User     pgdb.AuthorizedUser `sql:"-"`
	UserID   int64
	DateTime time.Time
	// BreathingRate is nested in a struct in the API,
	// We expose its value as the column BreathingRate
	Value         types.BreathingRateValue `sql:"-"`
	BreathingRate float64
}

func (BreathingRate) Headers() []string {
	return []string{
		"BreathingRate",
	}
}

func (f *BreathingRate) Values() []string {
	return []string{
		strconv.FormatFloat(f.Value.BreathingRate, 'f', 2, 64),
	}
}

func (BreathingRate) TableName() string {
	return "breathing_rate"
}

type BreathingRateTimePoint struct {
	types.BreathingRateTimePoint
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (BreathingRateTimePoint) TableName() string {
	return "breathing_rate_series"
}
