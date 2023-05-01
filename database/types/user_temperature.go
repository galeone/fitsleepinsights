// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"time"

	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type CoreTemperature struct {
	types.CoreTemperatureTimePoint
	ID       int64               `igor:"primary_key"`
	User     pgdb.AuthorizedUser `sql:"-"`
	UserID   int64
	DateTime types.FitbitDate `sql:"-"` // it's a date
	Date     time.Time
}

func (CoreTemperature) TableName() string {
	return "core_temperatures"
}

type SkinTemperature struct {
	types.SkinTemperatureTimePoint
	ID       int64               `igor:"primary_key"`
	User     pgdb.AuthorizedUser `sql:"-"`
	UserID   int64
	DateTime types.FitbitDate `sql:"-"` // it's a date
	Date     time.Time
	Value    float64 // for some reason in the API this value is wrapped in a structure
}

func (SkinTemperature) TableName() string {
	return "skin_temperatures"
}
