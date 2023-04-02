// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type HeartRateTimePointValue struct {
	types.HeartRateTimePointValue
	ID            int64         `igor:"primary_key"`
	HeartRateZone HeartRateZone `sql:"-"`
}

func (HeartRateTimePointValue) TableName() string {
	return "heart_rate_time_point_values"
}

type HeartRateActivities struct {
	types.HeartRateActivities
	ID                      int64               `igor:"primary_key"`
	User                    pgdb.AuthorizedUser `sql:"-"`
	UserID                  int64
	HeartRateTimePointValue HeartRateTimePointValue `sql:"-"`
}

func (HeartRateActivities) TableName() string {
	return "heart_rate_activities"
}
