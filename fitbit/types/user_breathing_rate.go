// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /br/date/%s.json

type BreathingRate struct {
	Br []BreathingRateTimePoint `json:"br"`
}

type BreathingRateValue struct {
	BreathingRate float64 `json:"breathingRate"`
}

type BreathingRateTimePoint struct {
	DateTime FitbitDateTime     `json:"dateTime"`
	Value    BreathingRateValue `json:"value"`
}
