// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /hrv/date/%s.json

type HeartRateVariability struct {
	Hrv []HeartRateVTimeStep `json:"hrv"`
}

type HeartRateVariabilityValue struct {
	DailyRmssd float64 `json:"dailyRmssd"`
	DeepRmssd  float64 `json:"deepRmssd"`
}

type HeartRateVTimeStep struct {
	DateTime FitbitDateTime            `json:"dateTime"`
	Value    HeartRateVariabilityValue `json:"value"`
}
