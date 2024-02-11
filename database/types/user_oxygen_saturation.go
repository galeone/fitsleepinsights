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

func (OxygenSaturation) Headers() []string {
	return []string{
		"AvgOxygenSaturation",
		"MaxOxygenSaturation",
		"MinOxygenSaturation",
	}
}

func (f *OxygenSaturation) Values() []string {
	return []string{
		strconv.FormatFloat(f.Avg, 'f', 2, 64),
		strconv.FormatFloat(f.Max, 'f', 2, 64),
		strconv.FormatFloat(f.Min, 'f', 2, 64),
	}
}

func (OxygenSaturation) TableName() string {
	return "oxygen_saturation"
}
