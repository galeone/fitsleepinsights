package types

// /cardioscore/date/%s.json

type CardioFitnessScore struct {
	CardioScore []CardioScoreTimePoint `json:"cardioScore"`
}

type CardioScoreTimePoint struct {
	DateTime string           `json:"dateTime"`
	Value    CardioScoreValue `json:"value"`
}

type CardioScoreValue struct {
	Vo2Max string `json:"vo2Max"`
}
