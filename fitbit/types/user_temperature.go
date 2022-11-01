// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /temp/core/date/%s.json

type CoreTemperature struct {
	TempCore []CoreTemperatureTimePoint `json:"tempCore"`
}

type CoreTemperatureTimePoint struct {
	DateTime FitbitDateTime `json:"dateTime"`
	Value    float64        `json:"value"`
}

// /temp/skin/date/%s.json

type SkinTemperature struct {
	TempSkin []SkinTemperatureTimePoint `json:"tempSkin"`
}

type SkinTemperatureTimePoint struct {
	DateTime FitbitDateTime       `json:"dateTime"`
	LogType  string               `json:"logType"`
	Value    SkinTemperatureValue `json:"value"`
}

type SkinTemperatureValue struct {
	NightlyRelative float64 `json:"nightlyRelative"`
}
