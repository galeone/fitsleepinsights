package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func dailyStepCount(all []*UserData, calendarType CalendarType) *charts.HeatMap {
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
		charts.WithVisualMapOpts(globalVisualMapSettings(maxSteps, "continuous")),
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

func activityCalendar(activityType *UserActivityTypes, activities *DailyActivities, calendarType CalendarType) *charts.HeatMap {
	var maxIndicator float64
	var measurementUnit string
	var defaultActivityIndicator string
	var activityValuePerYear map[int][]opts.HeatMapData = make(map[int][]opts.HeatMapData)
	var coveredMonthsPerYear map[int]map[int]bool = make(map[int]map[int]bool)

	// Depending on the activityType passed we have different default indicators
	// and other values to compute the stats. In the global variable _allActivityCatalog
	// we have the complete list of activities, so we can use it to get the correct values
	activityNameLowerCase := strings.ToLower(activityType.Name)
	for _, activity := range _allActivityCatalog {
		// All the activities IDs refer to the activity_description IDS
		for _, subcategory := range activity.SubCategories {
			for _, activityDescription := range subcategory.Activities {
				if activityDescription.ID == activityType.ID || strings.ToLower(activityDescription.Name) == activityNameLowerCase {
					if activityDescription.HasSpeed {
						defaultActivityIndicator = "Distance"
					}
					break
				}
			}
		}
	}

	if defaultActivityIndicator == "" {
		defaultActivityIndicator = "Time"
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
		value := [2]interface{}{activity.StartTime.Format(time.DateOnly), twoDecimals(activityIndicator)}
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
		charts.WithVisualMapOpts(globalVisualMapSettings(int(maxIndicator), "continuous")),
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

type ActivityStats struct {
	// Totals
	TotalTime              float64
	TotalDistance          float64
	TotalCalories          int64
	TotalSteps             int64
	TotalActiveTime        float64
	TotalActiveZoneMinutes int64
	TotalMinutesInFatBurn  int64
	TotalMinutesInCardio   int64
	TotalMinutesInPeak     int64

	// Averages
	AverageTime      float64
	AverageHeartRate float64
	AveragePace      float64
	AverageSpeed     float64
	AverageDistance  float64
	AverageCalories  float64
	AverageSteps     float64

	// Max
	MaxElevationGain int64
	MaxPace          float64
	MaxSpeed         float64
}

type SleepStats struct {
	TotalSleepTime            float64
	TotalSleepEfficiency      float64
	TotalSleepMinutes         float64
	TotalRemSleepMinutes      float64
	TotalLightSleepMinutes    float64
	TotalDeepSleepMinutes     float64
	TotalWakeSleepMinutes     float64
	TotalAwakeSleepMinutes    float64
	TotalRestlessSleepMinutes float64
	TotalTimeInBed            float64
	TotalTimeAsleep           float64
	TotalTimeAwake            float64
	TotalTimeToFallAsleep     float64
	TotalTimeAfterWakeup      float64
	TotalTimeInBedAwake       float64
	TotalTimeInBedAsleep      float64
	TotalTimeInBedRestless    float64
}

func activityStats(activities *DailyActivities) *ActivityStats {
	var stats ActivityStats

	for _, activity := range *activities {
		stats.TotalTime += float64(activity.Duration) * msToMin
		stats.TotalDistance += activity.Distance
		stats.TotalCalories += activity.Calories
		stats.TotalSteps += activity.Steps
		stats.TotalActiveTime += float64(activity.ActiveDuration) * msToMin
		stats.TotalActiveZoneMinutes += activity.ActiveZoneMinutes.TotalMinutes

		for _, zone := range activity.ActiveZoneMinutes.MinutesInHeartRateZones {
			switch zone.ZoneName {
			case "Fat Burn":
				stats.TotalMinutesInFatBurn += zone.Minutes
			case "Cardio":
				stats.TotalMinutesInCardio += zone.Minutes
			case "Peak":
				stats.TotalMinutesInPeak += zone.Minutes
			}
		}

		if activity.ElevationGain > stats.MaxElevationGain {
			stats.MaxElevationGain = activity.ElevationGain
		}
		if activity.Pace > stats.MaxPace {
			stats.MaxPace = activity.Pace
		}
		if activity.Speed > stats.MaxSpeed {
			stats.MaxSpeed = activity.Speed
		}

		stats.AverageHeartRate += float64(activity.AverageHeartRate)
		stats.AveragePace += activity.Pace
		stats.AverageSpeed += activity.Speed

	}
	// Average
	tot := float64(len(*activities))
	stats.AverageHeartRate /= tot
	stats.AveragePace /= tot
	stats.AverageSpeed /= tot
	stats.AverageTime = stats.TotalTime / tot
	stats.AverageDistance = float64(stats.TotalDistance) / tot
	stats.AverageCalories = float64(stats.TotalCalories) / tot
	stats.AverageSteps = float64(stats.TotalSteps) / tot

	// Two decimals for all the float values
	stats.TotalTime = twoDecimals(stats.TotalTime)
	stats.TotalDistance = twoDecimals(stats.TotalDistance)
	stats.TotalActiveTime = twoDecimals(stats.TotalActiveTime)
	stats.AverageHeartRate = twoDecimals(stats.AverageHeartRate)
	stats.AveragePace = twoDecimals(stats.AveragePace)
	stats.AverageSpeed = twoDecimals(stats.AverageSpeed)
	stats.AverageDistance = twoDecimals(stats.AverageDistance)
	stats.AverageCalories = twoDecimals(stats.AverageCalories)
	stats.AverageTime = twoDecimals(stats.AverageTime)
	stats.AverageSteps = twoDecimals(stats.AverageSteps)
	stats.MaxPace = twoDecimals(stats.MaxPace)
	stats.MaxSpeed = twoDecimals(stats.MaxSpeed)

	return &stats

}
