package app

import (
	"fmt"
	"math"
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
	verticalOffset   int = 120
	cellSize         int = 15
	marginLeft       int = 60
	marginRight      int = 30
	minCalendarWidth int = cellSize * 6 * 3
	maxCalendarWidth int = cellSize * 52

	msToMin float64 = 1.0 / (1000.0 * 60.0)
)

type CalendarType int

const (
	WeeklyCalendar CalendarType = iota
	MonthlyCalendar
	YearlyCalendar
)

// globalChartSettings returns the global settings for the calendar charts
// these are the settings that all the calendars should have in common
func globalChartSettings(calendarType CalendarType, numYears int) charts.GlobalOpts {
	var contentWidth int
	switch calendarType {
	case WeeklyCalendar, MonthlyCalendar:
		contentWidth = minCalendarWidth
	case YearlyCalendar:
		contentWidth = maxCalendarWidth
	}
	return charts.WithInitializationOpts(opts.Initialization{
		Theme:  "dark",
		Height: fmt.Sprintf("%dpx", verticalOffset+numYears*(verticalOffset+30)),
		Width:  fmt.Sprintf("%dpx", contentWidth+marginLeft+marginRight),
	})
}

func globalCalendarSettings(calendarType CalendarType, id, year int, coveredMonthsPerYear map[int]map[int]bool, firstActivityDate time.Time) *opts.Calendar {
	var calendarRange []string = make([]string, 0, 2)
	var orient string = "horizontal"

	// Depending on the number of months covered by the data, we define a different range
	// in order to create a calendar without too many empty cells
	months := make([]int, 0, len(coveredMonthsPerYear[year]))
	for k := range coveredMonthsPerYear[year] {
		months = append(months, k)
	}
	sort.Ints(months)

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
		weekStartDay := GetStartDayOfWeek(firstActivityDate)
		calendarRange = append(calendarRange, weekStartDay.Format(time.DateOnly))
		calendarRange = append(calendarRange, weekStartDay.AddDate(0, 0, 7).Format(time.DateOnly))
	}

	return &opts.Calendar{
		Orient: orient,
		Silent: false,
		Range:  calendarRange,
		Top:    fmt.Sprintf("%d", verticalOffset+id*(verticalOffset+30)),
		Left:   "center", //fmt.Sprintf("%d", marginLeft),
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
	}
}

func globalVisualMapSettings(maxValue int) opts.VisualMap {
	return opts.VisualMap{
		//Type:   "piecewise",
		Calculable: true,
		Max:        float32(nextMultipleOfTen(maxValue)),
		Show:       true,
		Orient:     "horizontal",
		Left:       "center",
		Top:        fmt.Sprintf("%d", 50),
		TextStyle: &opts.TextStyle{
			Color: "white",
		},
	}
}

func globalTitleSettings(title string) opts.Title {
	return opts.Title{
		Title: title,
		Top:   "15",
		Left:  "center",
	}
}

func dailyStepCount(user *fitbit_pgdb.AuthorizedUser, all []*UserData, calendarType CalendarType) *charts.HeatMap {
	var dailyStepsPerYear map[int][]opts.HeatMapData = make(map[int][]opts.HeatMapData)
	var maxSteps int = 0
	var coveredMonthsPerYear map[int]map[int]bool = make(map[int]map[int]bool)
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
		month := int(dayData.Date.Month())
		dailyStepsPerYear[year] = append(dailyStepsPerYear[year], opts.HeatMapData{Value: value, Name: value[0].(string)})

		if _, ok := coveredMonthsPerYear[year]; !ok {
			coveredMonthsPerYear[year] = make(map[int]bool)
			coveredMonthsPerYear[year][month] = true
		}
	}

	years := make([]int, 0, len(dailyStepsPerYear))
	for k := range dailyStepsPerYear {
		years = append(years, k)
	}
	sort.Ints(years)

	chart := charts.NewHeatMap()
	chart.SetGlobalOptions(
		globalChartSettings(calendarType, len(years)),
		charts.WithTitleOpts(globalTitleSettings("Daily Steps")),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "item",
			Show:    true,
		}),
		charts.WithVisualMapOpts(globalVisualMapSettings(maxSteps)),
		charts.WithLegendOpts(opts.Legend{
			Show: false,
		}),
	)

	for id, year := range years {
		chart.AddSeries("Daily Steps", dailyStepsPerYear[year],
			charts.WithCoordinateSystem("calendar"),
			charts.WithCalendarIndex(id),
		)

		chart.AddCalendar(globalCalendarSettings(calendarType, id, year, coveredMonthsPerYear, all[0].Date))
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
		for _, subcategory := range activity.SubCategories {
			for _, activityDescription := range subcategory.Activities {
				if activityDescription.ID == activityType.ID {
					// TODO questo é rotto
					if activityDescription.HasSpeed {
						defaultActivityIndicator = "Distance"
					} else {
						defaultActivityIndicator = "Time"
					}
					break
				}
			}
		}
		fmt.Println("not found for ", activity.Name, activityType.ID)

	}

	for _, activity := range *activities {
		// Depending on the activity type, we use different default indicators for the heatmap
		// Distance is a good default indicator for Walking, Running, Treadmill, Swimming, Biking, Aerobic Workout...
		var activityIndicator float64
		if defaultActivityIndicator == "Distance" {
			activityIndicator = activity.Distance
		} else {
			activityIndicator = float64(activity.ActiveDuration) * msToMin
		}
		if measurementUnit == "" {
			if activity.DistanceUnit == "nd" {
				measurementUnit = "min"
			} else {
				measurementUnit = activity.DistanceUnit
			}
		}

		if activityIndicator > maxIndicator {
			maxIndicator = activityIndicator
		}
		// format date to YYYY-MM-DD
		value := [2]interface{}{activity.StartTime.Format(time.DateOnly), math.Round(activityIndicator*100) / 100}
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

	chart := charts.NewHeatMap()
	chart.SetGlobalOptions(
		charts.WithTitleOpts(
			globalTitleSettings(fmt.Sprintf("Daily %s %s [%s]", activityType.Name, defaultActivityIndicator, measurementUnit)),
		),
		globalChartSettings(calendarType, len(years)),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "item",
			Show:    true,
		}),
		charts.WithVisualMapOpts(globalVisualMapSettings(int(maxIndicator))),
		charts.WithLegendOpts(opts.Legend{
			Show: false,
		}),
	)

	for id, year := range years {
		chart.AddSeries(defaultActivityIndicator, activityValuePerYear[year],
			charts.WithCoordinateSystem("calendar"),
			charts.WithCalendarIndex(id),
		)
		chart.AddCalendar(globalCalendarSettings(calendarType, id, year, coveredMonthsPerYear, (*activities)[0].StartTime))
	}

	return chart
}
