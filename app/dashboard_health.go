package app

import (
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type HealthStats struct {
	AverageStartTime time.Time
	AverageEndTime   time.Time
	AverageDuration  float64
	MaxDuration      float64
	MinDuration      float64
}

type HealthDashboard struct {
	BreathingRate        *charts.Line
	HeartRateVariability *charts.Line
	SkinTemperature      *charts.Bar
	OxygenSaturation     *charts.Line
	RestingHeartRate     *charts.Line
	Weight               *charts.Line
	BMI                  *charts.Line
	Stats                *HealthStats
}

func healthDashboard(all []*UserData, calendarType CalendarType) *HealthDashboard {
	var dates []string

	var skinTemperature []opts.BarData
	var breathingRate, heartRateVariability, restingHeartRate []opts.LineData
	oxygenSaturation := map[string][]opts.LineData{
		"average": {},
		"min":     {},
		"max":     {},
	}

	var bmi, weight []opts.LineData

	counters := map[string]int{
		"skinTemperature":      0,
		"breathingRate":        0,
		"heartRateVariability": 0,
		"oxygenSaturation":     0,
		"restingHeartRate":     0,
		"weight":               0,
		"bmi":                  0,
	}
	for _, dayData := range all {
		if dayData == nil {
			continue
		}
		// format date to YYYY-MM-DD
		dates = append(dates, dayData.Date.Format(time.DateOnly))

		if dayData.SkinTemperature != nil {
			counters["skinTemperature"]++
			skinTemperature = append(skinTemperature, opts.BarData{Value: dayData.SkinTemperature.Value})
		}

		if dayData.BreathingRate != nil {
			counters["breathingRate"]++
			breathingRate = append(breathingRate, opts.LineData{Value: dayData.BreathingRate.BreathingRateTimePoint.Value})
		}

		if dayData.HeartRateVariability != nil {
			counters["heartRateVariability"]++
			heartRateVariability = append(heartRateVariability, opts.LineData{Value: dayData.HeartRateVariability.DailyRmssd})
		}

		if dayData.OxygenSaturation != nil {
			counters["oxygenSaturation"]++
			oxygenSaturation["min"] = append(oxygenSaturation["min"], opts.LineData{Value: dayData.OxygenSaturation.Min})
			oxygenSaturation["max"] = append(oxygenSaturation["max"], opts.LineData{Value: dayData.OxygenSaturation.Max})
			oxygenSaturation["average"] = append(oxygenSaturation["average"], opts.LineData{Value: dayData.OxygenSaturation.Avg})
		}

		if dayData.HeartRate != nil && dayData.HeartRate.RestingHeartRate.Valid {
			counters["restingHeartRate"]++
			restingHeartRate = append(restingHeartRate, opts.LineData{Value: dayData.HeartRate.RestingHeartRate.Int64})
		}

		if dayData.BodyWeight != nil {
			counters["weight"]++
			weight = append(weight, opts.LineData{Value: dayData.BodyWeight.Value})
		}
		if dayData.BMI != nil {
			counters["bmi"]++
			bmi = append(bmi, opts.LineData{Value: dayData.BMI.Value})
		}
	}

	skinTemperatureBarChart := charts.NewBar()
	skinTemperatureBarChart.SetGlobalOptions(
		globalChartSettings(calendarType, 1),
		charts.WithTitleOpts(globalTitleSettings("Skin Temperature")),
		charts.WithLegendOpts(globalLegendSettings()),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
	)
	skinTemperatureBarChart.SetXAxis(dates)
	skinTemperatureBarChart.AddSeries("Nightly Skin Temperature", skinTemperature, charts.WithLineChartOpts(opts.LineChart{
		Color: "#1976FF",
	}))

	breathingRateLineChart := charts.NewLine()
	breathingRateLineChart.SetGlobalOptions(
		charts.WithTitleOpts(globalTitleSettings("Breathing Rate")),
		globalChartSettings(calendarType, 1),
		charts.WithLegendOpts(globalLegendSettings()),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
	)
	breathingRateLineChart.SetXAxis(dates)
	breathingRateLineChart.AddSeries("Actual", breathingRate, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

	hrvLineChart := charts.NewLine()
	hrvLineChart.SetGlobalOptions(
		charts.WithTitleOpts(globalTitleSettings("Heart Rate Variability")),
		globalChartSettings(calendarType, 1),
		charts.WithLegendOpts(globalLegendSettings()),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
	)
	hrvLineChart.SetXAxis(dates)
	hrvLineChart.AddSeries("Actual", heartRateVariability, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

	sp02lineChart := charts.NewLine()
	sp02lineChart.SetGlobalOptions(
		charts.WithTitleOpts(globalTitleSettings("Oxygen Saturation")),
		globalChartSettings(calendarType, 1),
		charts.WithLegendOpts(globalLegendSettings()),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Min: 80,
			Max: 105,
		}),
	)
	sp02lineChart.SetXAxis(dates)
	sp02lineChart.AddSeries("Min", oxygenSaturation["min"], charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))
	sp02lineChart.AddSeries("Max", oxygenSaturation["max"], charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))
	sp02lineChart.AddSeries("Average", oxygenSaturation["average"], charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

	restingHeartRateLineChart := charts.NewLine()
	restingHeartRateLineChart.SetGlobalOptions(
		charts.WithTitleOpts(globalTitleSettings("Resting Heart Rate")),
		globalChartSettings(calendarType, 1),
		charts.WithLegendOpts(globalLegendSettings()),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
	)
	restingHeartRateLineChart.SetXAxis(dates)
	restingHeartRateLineChart.AddSeries("Actual", restingHeartRate, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

	weightLineChart := charts.NewLine()
	weightLineChart.SetGlobalOptions(
		charts.WithTitleOpts(globalTitleSettings("Weight")),
		globalChartSettings(calendarType, 1),
		charts.WithLegendOpts(globalLegendSettings()),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
	)
	weightLineChart.SetXAxis(dates)
	weightLineChart.AddSeries("Actual", weight, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

	bmiLineChart := charts.NewLine()
	bmiLineChart.SetGlobalOptions(
		charts.WithTitleOpts(globalTitleSettings("BMI")),
		globalChartSettings(calendarType, 1),
		charts.WithLegendOpts(globalLegendSettings()),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
	)
	bmiLineChart.SetXAxis(dates)
	bmiLineChart.AddSeries("Actual", bmi, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

	return &HealthDashboard{
		BreathingRate:        breathingRateLineChart,
		HeartRateVariability: hrvLineChart,
		SkinTemperature:      skinTemperatureBarChart,
		OxygenSaturation:     sp02lineChart,
		RestingHeartRate:     restingHeartRateLineChart,
		Weight:               weightLineChart,
		BMI:                  bmiLineChart,
		Stats:                &HealthStats{},
	}
}
