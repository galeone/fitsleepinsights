// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import "time"

// hrv/date/%s.json

type HRVSummary struct {
	Hrv []HRVTimeStep `json:"hrv"`
}

type HRVSummaryValue struct {
	DailyRmssd float64 `json:"dailyRmssd"`
	DeepRmssd  float64 `json:"deepRmssd"`
}

type HRVTimeStep struct {
	DateTime time.Time       `json:"dateTime"`
	Value    HRVSummaryValue `json:"value"`
}
