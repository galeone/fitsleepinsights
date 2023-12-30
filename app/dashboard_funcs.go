package app

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/vertexai/genai"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"google.golang.org/api/option"
)

func describeChartContent(chart *charts.BaseConfiguration, chartType string, additionalPrompts ...string) (string, error) {
	var description string = "This is the data used to generate the chart titled " + chart.Title.Title + "\n"
	description += "The data is in the format of a series of points.\n"
	description += "The data is in the healthcare domain.\n"
	description += "The data is in the format of a " + chartType + ".\n"

	description += "Here's the data in JSON format.\n"
	for _, series := range chart.MultiSeries {
		seriesInfo := map[string]interface{}{
			"name": series.Name,
			"data": series.Data,
		}
		if jsonData, err := json.Marshal(seriesInfo); err != nil {
			return "", err
		} else {
			description += fmt.Sprintf("%s\n", jsonData)
		}
	}

	for _, additionalPrompt := range additionalPrompts {
		description += additionalPrompt + "\n"
	}

	description += "Generate the chart description.\n"
	description += "Add to the description hints and insights about the data.\n"
	description += "The description must be in Markdown.\n"
	description += "The user that generated the data is reading your description. Talk directly to the user.\n"

	ctx := context.Background()
	// Access your API key as an environment variable (see "Set up your API key" above)
	var client *genai.Client
	var err error
	const region = "us-central1"
	if client, err = genai.NewClient(ctx, _vaiProjectID, region, option.WithCredentialsFile(_vaiServiceAccountKey)); err != nil {
		return "", err
	}
	defer client.Close()

	// For text-only input, use the gemini-pro model
	model := client.GenerativeModel("gemini-pro")
	var resp *genai.GenerateContentResponse
	if resp, err = model.GenerateContent(ctx, genai.Text(description)); err != nil {
		return "", err
	}
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned")
	}

	return fmt.Sprint(resp.Candidates[0].Content.Parts[0]), nil
}

func sleepAggregatedLineChart(user *fitbit_pgdb.AuthorizedUser, all []*UserData) *charts.Line {
	var dates []string
	var sleepDuration []opts.LineData
	//var minutesToFallAsleep []opts.LineData is always 0 (at least for me)
	var minutesAwake []opts.LineData
	var minutesAsleep []opts.LineData

	for _, dayData := range all {
		if dayData == nil || dayData.SleepLog == nil {
			continue
		}
		// format date to YYYY-MM-DD
		dates = append(dates, dayData.Date.Format(time.DateOnly))
		// Duration is in milliseconds, convert it to minutes
		sleepDuration = append(sleepDuration, opts.LineData{Value: dayData.SleepLog.Duration / 1000 / 60})
		//minutesToFallAsleep = append(minutesToFallAsleep, opts.LineData{Value: dayData.SleepLog.MinutesToFallAsleep})
		minutesAwake = append(minutesAwake, opts.LineData{Value: dayData.SleepLog.MinutesAwake})
		minutesAsleep = append(minutesAsleep, opts.LineData{Value: dayData.SleepLog.MinutesAsleep})
	}
	chart := charts.NewLine()

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

	chart.AddSeries("Duration", sleepDuration, charts.WithLineChartOpts(opts.LineChart{
		Smooth: false,
	}))
	//chart.AddSeries("Minutes To Fall Asleep", minutesToFallAsleep, charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	chart.AddSeries("Minutes Awake", minutesAwake, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))
	chart.AddSeries("Minutes Asleep", minutesAsleep, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

	return chart
}

func sleepAggregatedStackedBarChart(user *fitbit_pgdb.AuthorizedUser, all []*UserData) *charts.Bar {
	var dates []string
	// var sleepDuration []opts.BarData
	// var minutesToFallAsleep []opts.BarData
	var minutesAwake []opts.BarData
	var minutesAsleep []opts.BarData

	for _, dayData := range all {
		if dayData == nil || dayData.SleepLog == nil {
			continue
		}
		// format date to YYYY-MM-DD
		dates = append(dates, dayData.Date.Format(time.DateOnly))
		// Duration is in milliseconds, convert it to minutes
		// sleepDuration = append(sleepDuration, opts.BarData{Value: dayData.SleepLog.Duration / 1000 / 60})
		// minutesToFallAsleep = append(minutesToFallAsleep, opts.BarData{Value: dayData.SleepLog.MinutesToFallAsleep})
		minutesAwake = append(minutesAwake, opts.BarData{Value: dayData.SleepLog.MinutesAwake})
		minutesAsleep = append(minutesAsleep, opts.BarData{Value: dayData.SleepLog.MinutesAsleep})
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
	chart.AddSeries("Minutes Awake", minutesAwake, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
		Stack:  "minutes",
	}))
	chart.AddSeries("Minutes Asleep", minutesAsleep, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
		Stack:  "minutes",
	}))

	return chart
}

func dailyStepCount(user *fitbit_pgdb.AuthorizedUser, all []*UserData) *charts.HeatMap {
	var dailyStepsPerYear map[int][]opts.HeatMapData = make(map[int][]opts.HeatMapData)
	var maxSteps int = 0
	for _, dayData := range all {
		if dayData == nil || dayData.Steps == nil {
			continue
		}
		steps := int(dayData.Steps.Value)
		if steps > maxSteps {
			maxSteps = steps
		}
		// format date to YYYY-MM-DD
		value := [2]interface{}{dayData.Date.Format(time.DateOnly), steps}
		year := dayData.Date.Year()
		dailyStepsPerYear[year] = append(dailyStepsPerYear[year], opts.HeatMapData{Value: value, Name: value[0].(string)})
	}

	years := make([]int, 0, len(dailyStepsPerYear))
	for k := range dailyStepsPerYear {
		years = append(years, k)
	}
	sort.Ints(years)

	const verticalOffset int = 120

	chart := charts.NewHeatMap()
	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Daily Steps Count",
			Top:   "30",
			Left:  "center",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  "dark",
			Height: fmt.Sprintf("%dpx", verticalOffset+len(years)*(verticalOffset+30)),
			//Width:  fmt.Sprintf("%dpx", 15*52+60),
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "item",
			Show:    true,
		}),
		charts.WithVisualMapOpts(
			opts.VisualMap{
				Type:   "piecewise",
				Min:    0,
				Max:    float32(maxSteps),
				Show:   true,
				Orient: "horizontal",
				Left:   "center",
				Top:    "65",
				TextStyle: &opts.TextStyle{
					Color: "white",
				},
			}),
		charts.WithLegendOpts(opts.Legend{
			Show: false,
		}),
	)

	for id, year := range years {
		chart.AddSeries("Daily Steps", dailyStepsPerYear[year],
			charts.WithCoordinateSystem("calendar"),
			charts.WithCalendarIndex(id),
		)

		chart.AddCalendar(&opts.Calendar{
			Orient:   "horizontal",
			Silent:   false,
			Range:    []float32{float32(year)},
			Top:      fmt.Sprintf("%d", verticalOffset+id*(verticalOffset+30)),
			Left:     "60",
			Right:    "30",
			CellSize: "15",
			ItemStyle: &opts.ItemStyle{
				BorderWidth: 0.5,
			},
			YearLabel: &opts.CalendarLabel{
				Show: true,
			},
			DayLabel: &opts.CalendarLabel{
				Show:  true,
				Color: "white",
			},
			MonthLabel: &opts.CalendarLabel{
				Show:  true,
				Color: "white",
			},
		})
	}

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
