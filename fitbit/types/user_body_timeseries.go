package types

// /body/%s/date/%s/%s.json

type BodyWeightSeries struct {
	TimeSeries []TimeStep `json:"body-weight"`
}

// /body/%s/date/%s/%s.json

type BodyBMISeries struct {
	TimeSeries []TimeStep `json:"body-bmi"`
}

// /body/%s/date/%s/%s.json

type BodyFatSeries struct {
	TimeSeries []TimeStep `json:"body-fat"`
}
