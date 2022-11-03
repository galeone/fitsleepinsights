// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /activities/heart/date/%s/%s.json

type HeartRateSeries struct {
	Heart []HeartRateActivities `json:"activities-heart"`
}

type HeartRateZones struct {
	CaloriesOut float64 `json:"caloriesOut"`
	Max         int64   `json:"max"`
	Min         int64   `json:"min"`
	Minutes     int64   `json:"minutes"`
	Name        string  `json:"name"`
}

type HeartRateTimePointValue struct {
	CustomHeartRateZones []HeartRateZones `json:"customHeartRateZones"`
	HeartRateZones       []HeartRateZones `json:"heartRateZones"`
	RestingHeartRate     int64            `json:"restingHeartRate"`
}

type HeartRateActivities struct {
	DateTime FitbitDate              `json:"dateTime"`
	Value    HeartRateTimePointValue `json:"value"`
}
