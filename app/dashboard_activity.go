package app

import (
	"fmt"
	"sort"
	"time"

	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func nextMultipleOfTen(n int) int {
	// https://stackoverflow.com/a/2403917/2891324
	return ((n + 9) / 10) * 10
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
				Max:    float32(nextMultipleOfTen(maxSteps)),
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

func activityCalendar(user *fitbit_pgdb.AuthorizedUser, activities *DailyActivities) *charts.HeatMap {
	var activityName string
	var maxIndicator float64
	var measurementUnit string
	var defaultActivityIndicator string
	var activityValuePerYear map[int][]opts.HeatMapData = make(map[int][]opts.HeatMapData)

	for _, activity := range *activities {
		if activityName == "" {
			activityName = activity.ActivityName
		}
		// Depending on the activity type, we use different default indicators for the heatmap
		// Distance is a good default indicator for Walking, Running, Treadmill, Swimming, Biking, Aerobic Workout...
		activityIndicator := activity.Distance
		defaultActivityIndicator = "Distance"
		measurementUnit = activity.DistanceUnit

		// Sport is very generic and often automatically added by Fitbit, so we use the ActiveDuration instead
		// We do the same for Weights
		if activity.ActivityName == "Sport" {
			activityIndicator = float64(activity.ActiveDuration)
			defaultActivityIndicator = "ActiveDuration"
			measurementUnit = "minutes"
		} else if activity.ActivityName == "Weights" {
			activityIndicator = float64(activity.ActiveDuration)
			defaultActivityIndicator = "ActiveDuration"
			measurementUnit = "minutes"
		}

		if activityIndicator > maxIndicator {
			maxIndicator = activityIndicator
		}
		// format date to YYYY-MM-DD
		value := [2]interface{}{activity.StartTime.Format(time.DateOnly), activityIndicator}
		year := activity.StartTime.Year()
		activityValuePerYear[year] = append(activityValuePerYear[year], opts.HeatMapData{Value: value, Name: value[0].(string)})
	}

	years := make([]int, 0, len(activityValuePerYear))
	for k := range activityValuePerYear {
		years = append(years, k)
	}
	sort.Ints(years)

	const verticalOffset int = 120

	chart := charts.NewHeatMap()
	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: fmt.Sprintf("Daily %s %s [%s]", activityName, defaultActivityIndicator, measurementUnit),
			Top:   "30",
			Left:  "center",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  "dark",
			Height: fmt.Sprintf("%dpx", verticalOffset+len(years)*(verticalOffset+30)),
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "item",
			Show:    true,
		}),
		charts.WithVisualMapOpts(
			opts.VisualMap{
				Type:   "piecewise",
				Min:    0,
				Max:    float32(nextMultipleOfTen(int(maxIndicator))),
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
		chart.AddSeries(defaultActivityIndicator, activityValuePerYear[year],
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
