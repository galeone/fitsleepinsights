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

const (
	verticalOffset int = 120
	cellSize       int = 15
	marginLeft     int = 60
	marginRight    int = 30
)

type CalendarType int

const (
	WeeklyCalendar CalendarType = iota
	MonthlyCalendar
	YearlyCalendar
)

const msToMin float64 = 1.0 / (1000.0 * 60.0)

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
			Range:    []string{fmt.Sprintf("%d", year)},
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

func activityCalendar(user *fitbit_pgdb.AuthorizedUser, activityType *UserActivityTypes, activities *DailyActivities, calendarType CalendarType) *charts.HeatMap {
	var maxIndicator float64
	var measurementUnit string
	var defaultActivityIndicator string
	var activityValuePerYear map[int][]opts.HeatMapData = make(map[int][]opts.HeatMapData)
	var coveredMonthsPerYear map[int]map[int]bool = make(map[int]map[int]bool)

	// Depending on the activityType passed we have different default indicators
	// and other values to compute the stats. In the global variable _allActivityCatalog
	// we have the complete list of activities, so we can use it to get the correct values
	// TODO: handle
	for _, activity := range _allActivityCatalog {
		// All the activities IDs refer to the activity_description IDS

		for _, description := range activity.Activities {
			if description.ID == activityType.ID {
				activityType.Name = activity.Name
				break
			}
		}

		for _, subcategory := range activity.SubCategories {
			for _, activityDescription := range subcategory.Activities {
				if activityDescription.ID == activityType.ID {
					activityType.Name = activity.Name
					break
				}
			}
		}

	}

	switch activityType.Name {
	case "Sport", "Weights":
		defaultActivityIndicator = "ActiveDuration"
		measurementUnit = "minutes"
	default:
		defaultActivityIndicator = "Distance"
		measurementUnit = "km"
	}

	if (*activities)[0].DistanceUnit != "nd" {
		measurementUnit = (*activities)[0].DistanceUnit
	}

	for _, activity := range *activities {
		// Depending on the activity type, we use different default indicators for the heatmap
		// Distance is a good default indicator for Walking, Running, Treadmill, Swimming, Biking, Aerobic Workout...
		activityIndicator := activity.Distance
		defaultActivityIndicator = "Distance"

		measurementUnit = activity.DistanceUnit

		// Sport is very generic and often automatically added by Fitbit, so we use the ActiveDuration instead
		// We do the same for Weights
		switch activity.ActivityName {
		case "Sport", "Weights":
			activityIndicator = float64(activity.ActiveDuration) * msToMin
		}

		if activityIndicator > maxIndicator {
			maxIndicator = activityIndicator
		}
		// format date to YYYY-MM-DD
		value := [2]interface{}{activity.StartTime.Format(time.DateOnly), activityIndicator}
		year := activity.StartTime.Year()
		month := int(activity.StartTime.Month())
		if _, ok := coveredMonthsPerYear[year]; !ok {
			coveredMonthsPerYear[year] = make(map[int]bool)
			coveredMonthsPerYear[year][month] = true
		}
		activityValuePerYear[year] = append(activityValuePerYear[year], opts.HeatMapData{Value: value, Name: value[0].(string)})
	}

	years := make([]int, 0, len(activityValuePerYear))
	for k := range activityValuePerYear {
		years = append(years, k)
	}
	sort.Ints(years)

	var contentWidth int
	switch calendarType {
	case WeeklyCalendar, MonthlyCalendar:
		contentWidth = cellSize * 25
	case YearlyCalendar:
		contentWidth = cellSize * 52
	}

	chart := charts.NewHeatMap()
	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: fmt.Sprintf("Daily %s %s [%s]", activityType.Name, defaultActivityIndicator, measurementUnit),
			Top:   "30",
			Left:  "center",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  "dark",
			Height: fmt.Sprintf("%dpx", verticalOffset+len(years)*(verticalOffset+30)),
			Width:  fmt.Sprintf("%dpx", contentWidth+marginLeft+marginRight),
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

		// Depending on the number of months covered by the data, we define a different range
		// in order to create a calendar without too many empty cells

		months := make([]int, 0, len(coveredMonthsPerYear[year]))
		for k := range coveredMonthsPerYear[year] {
			months = append(months, k)
		}
		sort.Ints(months)

		var calendarRange []string = make([]string, 0, 2)
		var orient string = "horizontal"
		if calendarType == MonthlyCalendar {
			calendarRange = append(calendarRange, fmt.Sprintf("%d-%02d", year, months[0]))
			if months[0] == 12 {
				calendarRange = append(calendarRange, fmt.Sprintf("%d-%02d", year+1, 1))
			} else {
				calendarRange = append(calendarRange, fmt.Sprintf("%d-%02d", year, months[0]+1))
			}
		} else if calendarType == YearlyCalendar {
			calendarRange = append(calendarRange, fmt.Sprintf("%d", year))
		} else if calendarType == WeeklyCalendar {
			// Weekly calendar: get an activity date, extract the first day of the week and use it as the starting point
			weekStartDay := GetStartDayOfWeek((*activities)[0].StartTime)
			calendarRange = append(calendarRange, weekStartDay.Format(time.DateOnly))
			calendarRange = append(calendarRange, weekStartDay.AddDate(0, 0, 7).Format(time.DateOnly))
			orient = "vertical"
		}

		chart.AddCalendar(&opts.Calendar{
			Orient: orient,
			Silent: false,
			Range:  calendarRange,
			Top:    fmt.Sprintf("%d", verticalOffset+id*(verticalOffset+30)),
			Left:   "60",
			// Right:    "30", keeping this commented allows us to have cell of the same sizes
			CellSize: fmt.Sprintf("%d", cellSize),
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
