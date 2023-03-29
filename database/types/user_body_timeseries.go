// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type BodyWeightSeries struct {
	types.BodyWeightSeries
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (BodyWeightSeries) TableName() string {
	return "body_weight_series"
}

type BMISeries struct {
	types.BMISeries
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (BMISeries) TableName() string {
	return "bmi_series"
}

type BodyFatSeries struct {
	types.BodyFatSeries
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (BodyFatSeries) TableName() string {
	return "body_fat_series"
}
