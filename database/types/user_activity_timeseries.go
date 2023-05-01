// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"time"

	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type ActivityCaloriesSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (ActivityCaloriesSeries) TableName() string {
	return "activity_calories_series"
}

type CaloriesSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (CaloriesSeries) TableName() string {
	return "calories_series"
}

type CaloriesBMRSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (CaloriesBMRSeries) TableName() string {
	return "calories_bmr_series"
}

type DistanceSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (DistanceSeries) TableName() string {
	return "distance_series"
}

type ElevationSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (ElevationSeries) TableName() string {
	return "elevation_series"
}

type FloorsSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (FloorsSeries) TableName() string {
	return "floors_series"
}

type MinutesSedentarySeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (MinutesSedentarySeries) TableName() string {
	return "minutes_sedentary_series"
}

type MinutesLightlyActiveSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (MinutesLightlyActiveSeries) TableName() string {
	return "minutes_lightly_active_series"
}

type MinutesFairlyActiveSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (MinutesFairlyActiveSeries) TableName() string {
	return "minutes_fairly_active_series"
}

type MinutesVeryActiveSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (MinutesVeryActiveSeries) TableName() string {
	return "minutes_very_active_series"
}

type StepsSeries struct {
	types.TimeStep
	DateTime types.FitbitDate `sql:"-"` // It's a Date
	Date     time.Time
	// Overwrite Value type. In the API it's returned as a string
	Value  float64
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (StepsSeries) TableName() string {
	return "steps_series"
}
