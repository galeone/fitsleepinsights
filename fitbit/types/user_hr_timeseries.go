// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /activities/heart/date/%s/%s.json

type HRSeries struct {
	Heart []HRActivities `json:"activities-heart"`
}

type HRZones struct {
	CaloriesOut float64 `json:"caloriesOut"`
	Max         int64   `json:"max"`
	Min         int64   `json:"min"`
	Minutes     int64   `json:"minutes"`
	Name        string  `json:"name"`
}

type HRTimePointValue struct {
	CustomHeartRateZones []HRZones `json:"customHeartRateZones"`
	HeartRateZones       []HRZones `json:"heartRateZones"`
	RestingHeartRate     int64     `json:"restingHeartRate"`
}

type HRActivities struct {
	DateTime FitbitDateTime   `json:"dateTime"`
	Value    HRTimePointValue `json:"value"`
}
