// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /body/%s/date/%s/%s.json

type BodyWeightSeries struct {
	TimeSeries []TimeStep `json:"body-weight"`
}

// /body/%s/date/%s/%s.json

type BMISeries struct {
	TimeSeries []TimeStep `json:"body-bmi"`
}

// /body/%s/date/%s/%s.json

type BodyFatSeries struct {
	TimeSeries []TimeStep `json:"body-fat"`
}
