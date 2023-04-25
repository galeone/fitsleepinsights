// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"time"

	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type OxygenSaturation struct {
	types.OxygenSaturation
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
	// Value nested in the api response
	Value string `sql:"-"`
	Avg   float64
	Max   float64
	Min   float64
	// DateTime
	DateTime time.Time `sql:"-"` // it's a date
	Date     time.Time
}

func (OxygenSaturation) TableName() string {
	return "oxygen_saturation"
}
