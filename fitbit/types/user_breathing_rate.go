package types

// /br/date/%s.json

type BreathingRate struct {
	Br []BreathingRateTimePoint `json:"br"`
}

type BreathingRateValue struct {
	BreathingRate float64 `json:"breathingRate"`
}

type BreathingRateTimePoint struct {
	DateTime string             `json:"dateTime"`
	Value    BreathingRateValue `json:"value"`
}
