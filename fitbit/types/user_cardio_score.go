// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /cardioscore/date/%s.json

type CardioFitnessScore struct {
	CardioScore []CardioScoreTimePoint `json:"cardioScore"`
}

type CardioScoreTimePoint struct {
	DateTime FitbitDateTime   `json:"dateTime"`
	Value    CardioScoreValue `json:"value"`
}

type CardioScoreValue struct {
	Vo2Max string `json:"vo2Max"`
}
