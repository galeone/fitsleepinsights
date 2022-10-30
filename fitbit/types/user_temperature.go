package types

// /temp/core/date/%s.json

type CoreTemperature struct {
	TempCore []CoreTemperatureTimePoint `json:"tempCore"`
}

type CoreTemperatureTimePoint struct {
	DateTime string  `json:"dateTime"`
	Value    float64 `json:"value"`
}

// /temp/skin/date/%s.json

type SkinTemperature struct {
	TempSkin []SkinTemperatureTimePoint `json:"tempSkin"`
}

type SkinTemperatureTimePoint struct {
	DateTime string               `json:"dateTime"`
	LogType  string               `json:"logType"`
	Value    SkinTemperatureValue `json:"value"`
}

type SkinTemperatureValue struct {
	NightlyRelative float64 `json:"nightlyRelative"`
}
