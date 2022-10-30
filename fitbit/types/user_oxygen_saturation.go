// /spo2/date/%s.json
package types

type OxygenSaturationSummary struct {
	DateTime string                `json:"dateTime"`
	Value    OxygenSaturationValue `json:"value"`
}

type OxygenSaturationValue struct {
	Avg float64 `json:"avg"`
	Max float64 `json:"max"`
	Min float64 `json:"min"`
}
