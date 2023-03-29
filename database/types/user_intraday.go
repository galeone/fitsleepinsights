// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"time"

	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type CaloriesSeriesIntraday struct {
	types.CaloriesSeriesIntraday
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (CaloriesSeriesIntraday) TableName() string {
	return "calories_series_intraday"
}

type DistanceSeriesIntraday struct {
	types.DistanceSeriesIntraday
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (DistanceSeriesIntraday) TableName() string {
	return "distance_series_intraday"
}

type ElevationSeriesIntraday struct {
	types.ElevationSeriesIntraday
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (ElevationSeriesIntraday) TableName() string {
	return "elevation_series_intraday"
}

type FloorsSeriesIntraday struct {
	types.FloorsSeriesIntraday
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (FloorsSeriesIntraday) TableName() string {
	return "floors_series_intraday"
}

type StepsSeriesIntraday struct {
	types.StepsSeriesIntraday
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (StepsSeriesIntraday) TableName() string {
	return "steps_series_intraday"
}

type OxygenSaturationIntraday struct {
	types.OxygenSaturationIntraday
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (OxygenSaturationIntraday) TableName() string {
	return "oxygen_saturation_intraday"
}

type HeartRateVariabilityIntradayHRV struct {
	types.HeartRateVariabilityValueIntraday
	ID       int64               `igor:"primary_key"`
	User     pgdb.AuthorizedUser `sql:"-"`
	DateTime time.Time           // required
}

func (HeartRateVariabilityIntradayHRV) TableName() string {
	return "heart_rate_variability_intraday_hrv"
}

type BreathingRateIntraday struct {
	types.BreathingRateIntradaySummary
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (BreathingRateIntraday) TableName() string {
	return "breathing_rate_intraday"
}
