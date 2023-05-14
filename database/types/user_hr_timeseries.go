// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"database/sql"
	"strconv"
	"time"

	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type HeartRateActivities struct {
	types.HeartRateActivities
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
	// Value is a struct containing an array and a field (resting heart rate).
	// The array is a cardio_zone (handled separately), here we ignore value and add
	// Directly the fields (array, to fill with scan queries) here.
	// All the fields are ignored, but we'll handle them manually (in this way igor ignores them).
	Value                int64           `sql:"-"`
	CustomHeartRateZones []HeartRateZone `sql:"-"`
	HeartRateZones       []HeartRateZone `sql:"-"`
	RestingHeartRate     sql.NullInt64
	DateTime             types.FitbitDate `sql:"-"` // It's a Date
	Date                 time.Time
}

func (HeartRateActivities) Headers() []string {
	return []string{
		"RestingHeartRate",

		"OutOfRangeMinutes",
		"OutOfRangeCalories",
		"OutOfRangeMaxBPM",
		"OutOfRangeMinBPM",

		"FatBurnMinutes",
		"FatBurnCalories",
		"FatBurnMaxBPM",
		"FatBurnMinBPM",

		"CardioMinutes",
		"CardioCalories",
		"CardioMaxBPM",
		"CardioMinBPM",

		"PeakMinutes",
		"PeakCalories",
		"PeakMaxBPM",
		"PeakMinBPM",
	}
}

func (f *HeartRateActivities) Values() []string {
	var outOfRangeMinutes, outOfRangeMaxBPM, outOfRangeMinBPM int64
	var fatBurnMinutes, fatBurnMaxBPM, fatBurnMinBPM int64
	var cardioMinutes, cardioMaxBPM, cardioMinBPM int64
	var peakMinutes, peakMaxBPM, peakMinBPM int64

	var outOfRangeCalories, fatBurnCalories, cardioCalories, peakCalories float64

	for _, zone := range f.HeartRateZones {
		switch zone.Name {
		case "Out of Range":
			outOfRangeMinutes = zone.Minutes
			outOfRangeCalories = zone.CaloriesOut
			outOfRangeMaxBPM = zone.Max
			outOfRangeMinBPM = zone.Min
		case "Fat Burn":
			fatBurnMinutes = zone.Minutes
			fatBurnCalories = zone.CaloriesOut
			fatBurnMaxBPM = zone.Max
			fatBurnMinBPM = zone.Min
		case "Cardio":
			cardioMinutes = zone.Minutes
			cardioCalories = zone.CaloriesOut
			cardioMaxBPM = zone.Max
			cardioMinBPM = zone.Min
		case "Peak":
			peakMinutes = zone.Minutes
			peakCalories = zone.CaloriesOut
			peakMaxBPM = zone.Max
			peakMinBPM = zone.Min
		}
	}

	var restingHRstring string
	if f.RestingHeartRate.Valid {
		restingHRstring = strconv.FormatInt(f.RestingHeartRate.Int64, 10)
	}

	return []string{
		restingHRstring,
		strconv.FormatInt(outOfRangeMinutes, 10),
		strconv.FormatFloat(outOfRangeCalories, 'f', 2, 64),
		strconv.FormatInt(outOfRangeMaxBPM, 10),
		strconv.FormatInt(outOfRangeMinBPM, 10),

		strconv.FormatInt(fatBurnMinutes, 10),
		strconv.FormatFloat(fatBurnCalories, 'f', 2, 64),
		strconv.FormatInt(fatBurnMaxBPM, 10),
		strconv.FormatInt(fatBurnMinBPM, 10),

		strconv.FormatInt(cardioMinutes, 10),
		strconv.FormatFloat(cardioCalories, 'f', 2, 64),
		strconv.FormatInt(cardioMaxBPM, 10),
		strconv.FormatInt(cardioMinBPM, 10),

		strconv.FormatInt(peakMinutes, 10),
		strconv.FormatFloat(peakCalories, 'f', 2, 64),
		strconv.FormatInt(peakMaxBPM, 10),
		strconv.FormatInt(peakMinBPM, 10),
	}
}

func (HeartRateActivities) TableName() string {
	return "heart_rate_activities"
}
