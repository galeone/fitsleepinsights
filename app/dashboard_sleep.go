package app

import (
	"time"

	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func sleepAggregatedStackedBarChart(user *fitbit_pgdb.AuthorizedUser, all []*UserData) *charts.Bar {
	var dates []string
	// var sleepDuration []opts.BarData
	// var minutesToFallAsleep []opts.BarData
	var minutesAsleep []opts.BarData

	var deepSleepMinutes []opts.BarData
	var lightSleepMinutes []opts.BarData
	var remSleepMinutes []opts.BarData
	var wakeSleepMinutes []opts.BarData

	for _, dayData := range all {
		if dayData == nil || dayData.SleepLog == nil {
			continue
		}
		// format date to YYYY-MM-DD
		dates = append(dates, dayData.Date.Format(time.DateOnly))
		// Duration is in milliseconds, convert it to minutes
		// sleepDuration = append(sleepDuration, opts.BarData{Value: dayData.SleepLog.Duration / 1000 / 60})
		// minutesToFallAsleep = append(minutesToFallAsleep, opts.BarData{Value: dayData.SleepLog.MinutesToFallAsleep})
		minutesAsleep = append(minutesAsleep, opts.BarData{Value: dayData.SleepLog.MinutesAsleep})

		deepSleepMinutes = append(deepSleepMinutes, opts.BarData{Value: dayData.SleepLog.Levels.Summary.Deep.Minutes})
		lightSleepMinutes = append(lightSleepMinutes, opts.BarData{Value: dayData.SleepLog.Levels.Summary.Light.Minutes})
		remSleepMinutes = append(remSleepMinutes, opts.BarData{Value: dayData.SleepLog.Levels.Summary.Rem.Minutes})
		wakeSleepMinutes = append(wakeSleepMinutes, opts.BarData{Value: dayData.SleepLog.Levels.Summary.Wake.Minutes})
	}
	chart := charts.NewBar()

	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Sleep Data",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: "dark",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
	)
	chart.SetXAxis(dates)

	// chart.AddSeries("Duration", sleepDuration, charts.WithLineChartOpts(opts.LineChart{Smooth: true,Stack:  "minutes"}))
	// chart.AddSeries("Minutes To Fall Asleep", minutesToFallAsleep, charts.WithLineChartOpts(opts.LineChart{Smooth: true,Stack:  "minutes"}))
	chart.AddSeries("Asleep", minutesAsleep, charts.WithLineChartOpts(opts.LineChart{
		Stack: "total",
		Color: "#1976FF",
	}))

	chart.AddSeries("Deep", deepSleepMinutes, charts.WithLineChartOpts(opts.LineChart{
		Stack: "sleepPhases",
		Color: "#1976D2",
	}))
	chart.AddSeries("Light", lightSleepMinutes, charts.WithLineChartOpts(opts.LineChart{
		Stack: "sleepPhases",
		Color: "#C59972",
	}))
	chart.AddSeries("Rem", remSleepMinutes, charts.WithLineChartOpts(opts.LineChart{
		Stack: "sleepPhases",
		Color: "#FAEBD7",
	}))
	chart.AddSeries("Wake", wakeSleepMinutes, charts.WithLineChartOpts(opts.LineChart{
		Stack: "sleepPhases",
		Color: "#AA0000",
	}))

	return chart
}

func sleepEfficiencyChart(user *fitbit_pgdb.AuthorizedUser, all []*UserData) *charts.Line {
	var dates []string
	var sleepEfficiency []opts.LineData
	for _, dayData := range all {
		if dayData == nil || dayData.SleepLog == nil {
			continue
		}
		// format date to YYYY-MM-DD
		dates = append(dates, dayData.Date.Format(time.DateOnly))
		sleepEfficiency = append(sleepEfficiency, opts.LineData{Value: dayData.SleepLog.Efficiency})
	}
	chart := charts.NewLine()

	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Sleep Efficiency",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: "dark",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
	)
	chart.SetXAxis(dates)

	chart.AddSeries("Actual", sleepEfficiency, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

	/*
		predictions, err := PredictSleepEfficiency(user, all)
		if err != nil {
			log.Println(err)
		} else {
			var predictedSleepEfficiency []opts.LineData
			for _, prediction := range predictions {
				predictedSleepEfficiency = append(predictedSleepEfficiency, opts.LineData{Value: prediction})
			}
			chart.AddSeries("Predicted", predictedSleepEfficiency, charts.WithLineChartOpts(opts.LineChart{
				Smooth: true,
			}))
		}
	*/
	return chart
}
