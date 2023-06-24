// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"strconv"
	"time"

	pgdb "github.com/galeone/fitbit-pgdb/v2"
	"github.com/galeone/fitbit/types"
)

type BodyWeightSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (BodyWeightSeries) TableName() string {
	return "body_weight_series"
}

func (BodyWeightSeries) Headers() []string {
	return []string{
		"BodyWeight",
	}
}

func (f *BodyWeightSeries) Values() []string {
	return []string{
		strconv.FormatFloat(f.Value, 'f', 2, 64),
	}
}

type BMISeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (BMISeries) Headers() []string {
	return []string{
		"BMI",
	}
}

func (f *BMISeries) Values() []string {
	return []string{
		strconv.FormatFloat(f.Value, 'f', 2, 64),
	}
}

func (BMISeries) TableName() string {
	return "bmi_series"
}

type BodyFatSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (BodyFatSeries) Headers() []string {
	return []string{
		"BodyFat",
	}
}

func (f *BodyFatSeries) Values() []string {
	return []string{
		strconv.FormatFloat(f.Value, 'f', 2, 64),
	}
}

func (BodyFatSeries) TableName() string {
	return "body_fat_series"
}
