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
	// Value changed type, who cares, we need to ignore it, it's a filed useful only when decoding json
	Value    int64            `sql:"-"`
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
}

func (HeartRateActivities) TableName() string {
	return "heart_rate_activities"
}
