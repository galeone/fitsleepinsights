// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	pgdb "github.com/galeone/fitbit-pgdb/v2"
	"github.com/galeone/fitbit/types"
)

type BreathingRateTimePoint struct {
	types.BreathingRateTimePoint
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (BreathingRateTimePoint) TableName() string {
	return "breathing_rate_series"
}
