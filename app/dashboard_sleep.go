package app

import (
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type SleepStats struct {
	AverageStartTime time.Time
	AverageEndTime   time.Time
	AverageDuration  float64
	MaxDuration      float64
	MinDuration      float64
}

type SleepDashboard struct {
	AggregatedStackedBarChart *charts.Bar
	EfficiencyLineChart       *charts.Line
	Stats                     *SleepStats
}

func sleepDashboard(all []*UserData, calendarType CalendarType) *SleepDashboard {
	var dates []string

	var minutesAsleep []opts.BarData
	var deepSleepMinutes []opts.BarData
	var lightSleepMinutes []opts.BarData
	var remSleepMinutes []opts.BarData
	var wakeSleepMinutes []opts.BarData

	var sleepEfficiency []opts.LineData

	var stats SleepStats
	var counter int64

	var unixStartTime int64
	var unixEndTime int64
	var location *time.Location

	for _, dayData := range all {
		if dayData == nil || dayData.SleepLog == nil {
			continue
		}
		counter++
		// format date to YYYY-MM-DD
		dates = append(dates, dayData.Date.Format(time.DateOnly))

		minutesAsleep = append(minutesAsleep, opts.BarData{Value: dayData.SleepLog.MinutesAsleep})
		deepSleepMinutes = append(deepSleepMinutes, opts.BarData{Value: dayData.SleepLog.Levels.Summary.Deep.Minutes})
		lightSleepMinutes = append(lightSleepMinutes, opts.BarData{Value: dayData.SleepLog.Levels.Summary.Light.Minutes})
		remSleepMinutes = append(remSleepMinutes, opts.BarData{Value: dayData.SleepLog.Levels.Summary.Rem.Minutes})
		wakeSleepMinutes = append(wakeSleepMinutes, opts.BarData{Value: dayData.SleepLog.Levels.Summary.Wake.Minutes})

		sleepEfficiency = append(sleepEfficiency, opts.LineData{Value: dayData.SleepLog.Efficiency})

		realDurationInMinutes := float64(dayData.SleepLog.Duration)*msToMin - float64(dayData.SleepLog.MinutesAwake)

		stats.AverageDuration += realDurationInMinutes

		if location == nil {
			location = dayData.SleepLog.StartTime.Location()
		}

		unixEndTime += int64(dayData.SleepLog.EndTime.Hour())*3600 + int64(dayData.SleepLog.EndTime.Minute())*60 + int64(dayData.SleepLog.EndTime.Second())
		unixStartTime += int64(dayData.SleepLog.StartTime.Hour())*3600 + int64(dayData.SleepLog.StartTime.Minute())*60 + int64(dayData.SleepLog.StartTime.Second())

		if realDurationInMinutes > stats.MaxDuration {
			stats.MaxDuration = realDurationInMinutes
		}
		if realDurationInMinutes < stats.MinDuration || stats.MinDuration == 0 {
			stats.MinDuration = realDurationInMinutes
		}

	}
	if counter > 0 {
		stats.AverageDuration = stats.AverageDuration / float64(counter)
		stats.AverageStartTime = time.Unix(unixStartTime/counter, 0).In(location)
		stats.AverageEndTime = time.Unix(unixEndTime/counter, 0).In(location)
	}

	aggregatedStackedBarChart := charts.NewBar()

	aggregatedStackedBarChart.SetGlobalOptions(
		globalChartSettings(calendarType, 1),
		charts.WithTitleOpts(globalTitleSettings("Sleep Data")),
		charts.WithLegendOpts(globalLegendSettings()),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside",
		}),
	)
	aggregatedStackedBarChart.SetXAxis(dates)

	aggregatedStackedBarChart.AddSeries("Asleep", minutesAsleep, charts.WithLineChartOpts(opts.LineChart{
		Stack: "total",
		Color: "#1976FF",
	}))

	aggregatedStackedBarChart.AddSeries("Deep", deepSleepMinutes, charts.WithLineChartOpts(opts.LineChart{
		Stack: "sleepPhases",
		Color: "#1976D2",
	}))
	aggregatedStackedBarChart.AddSeries("Light", lightSleepMinutes, charts.WithLineChartOpts(opts.LineChart{
		Stack: "sleepPhases",
		Color: "#C59972",
	}))
	aggregatedStackedBarChart.AddSeries("Rem", remSleepMinutes, charts.WithLineChartOpts(opts.LineChart{
		Stack: "sleepPhases",
		Color: "#FAEBD7",
	}))
	aggregatedStackedBarChart.AddSeries("Wake", wakeSleepMinutes, charts.WithLineChartOpts(opts.LineChart{
		Stack: "sleepPhases",
		Color: "#AA0000",
	}))

	sleepEfficiencyLineChart := charts.NewLine()

	sleepEfficiencyLineChart.SetGlobalOptions(
		charts.WithTitleOpts(globalTitleSettings("Sleep Efficiency")),
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
	sleepEfficiencyLineChart.SetXAxis(dates)

	sleepEfficiencyLineChart.AddSeries("Actual", sleepEfficiency, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

	return &SleepDashboard{
		AggregatedStackedBarChart: aggregatedStackedBarChart,
		EfficiencyLineChart:       sleepEfficiencyLineChart,
		Stats:                     &stats,
	}
}

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
