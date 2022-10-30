package types

// hrv/date/%s.json

type HRVSummary struct {
	Hrv []HRVTimeStep `json:"hrv"`
}

type HRVSummaryValue struct {
	DailyRmssd float64 `json:"dailyRmssd"`
	DeepRmssd  float64 `json:"deepRmssd"`
}

type HRVTimeStep struct {
	DateTime string          `json:"dateTime"`
	Value    HRVSummaryValue `json:"value"`
}
