package app

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"

	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/galeone/fitbit/v2"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/labstack/echo/v4"
)

func getUser(c echo.Context) (*fitbit_pgdb.AuthorizedUser, error) {
	// secure, under middleware
	var err error
	authorizer := c.Get("fitbit").(*fitbit.Authorizer)
	var userID *string
	if userID, err = authorizer.UserID(); err != nil {
		return nil, err
	}

	user := fitbit_pgdb.AuthorizedUser{}
	user.UserID = *userID
	if err = _db.Model(fitbit_pgdb.AuthorizedUser{}).Where(&user).Scan(&user); err != nil {
		return nil, err
	}
	return &user, err
}

func dashboard(c echo.Context, user *fitbit_pgdb.AuthorizedUser, startDate, endDate time.Time, calendarType CalendarType) (err error) {
	var fetcher *fetcher
	if fetcher, err = NewFetcher(user); err != nil {
		return err
	}

	allData := fetcher.FetchByRange(startDate, endDate)
	var activitiesTypes []UserActivityTypes
	if activitiesTypes, err = fetcher.UserActivityTypes(); err != nil {
		return err
	}

	activityCalendars := make(map[string]template.HTML)
	activityStatistics := make(map[string]*ActivityStats)
	activityCalendarsDescriptions := make(map[string]string)

	wg := sync.WaitGroup{}
	mapMux := sync.Mutex{}

	for _, activityType := range activitiesTypes {
		activityList := DailyActivities{}
		for _, dayData := range allData {
			if dayData == nil || dayData.Activities == nil {
				continue
			}
			for _, activity := range *dayData.Activities {
				if activity.ActivityTypeID == activityType.ID {
					activityList = append(activityList, activity)
				}
			}
		}
		if len(activityList) > 0 {
			wg.Add(1)
			go func(activityType UserActivityTypes) {
				defer wg.Done()
				chart := activityCalendar(&activityType, &activityList, calendarType)
				chart.Renderer = newChartRenderer(chart, chart.Validate)
				var activityCalendarDescription string
				if activityCalendarDescription, err = describeChartContent(&chart.BaseConfiguration, "calendar heatmap", fmt.Sprintf("Calendar for %s", activityType.Name)); err != nil {
					activityCalendarDescription = "Failed to generate description: " + err.Error()
				}

				mapMux.Lock()
				activityCalendars[activityType.Name] = renderChart(chart)
				activityStatistics[activityType.Name] = activityStats(&activityList)
				activityCalendarsDescriptions[activityType.Name] = activityCalendarDescription
				mapMux.Unlock()
			}(activityType)
		}
	}

	wg.Wait()

	var dailyStepChart *charts.HeatMap
	var dailyStepsCountDescription string

	var sleepEfficiencyChart *charts.Line
	var sleepEfficiencyDescription string

	var sleepAggregatedChart *charts.Bar
	var sleepAggregatedDescription string
	wg.Add(3)

	go func() {
		defer wg.Done()
		dailyStepChart = dailyStepCount(allData, calendarType)
		dailyStepChart.Renderer = newChartRenderer(dailyStepChart, dailyStepChart.Validate)

		if dailyStepsCountDescription, err = describeChartContent(&dailyStepChart.BaseConfiguration, "calendar heatmap"); err != nil {
			dailyStepsCountDescription = "Failed to generate description: " + err.Error()
		}

	}()

	go func() {
		defer wg.Done()

		sleepEfficiencyChart = sleepEfficiencyLineChart(allData, calendarType)
		sleepEfficiencyChart.Renderer = newChartRenderer(sleepEfficiencyChart, sleepEfficiencyChart.Validate)

		if sleepEfficiencyDescription, err = describeChartContent(&sleepEfficiencyChart.BaseConfiguration, "line chart", "Sleep Efficiency is a value in [0,100] computed as the ratio between the time spent in bed and the time effectively spent asleep"); err != nil {
			sleepEfficiencyDescription = "Failed to generate description: " + err.Error()
		}

	}()

	go func() {
		defer wg.Done()
		sleepAggregatedChart = sleepAggregatedStackedBarChart(allData, calendarType)
		sleepAggregatedChart.Renderer = newChartRenderer(sleepAggregatedChart, sleepAggregatedChart.Validate)

		if sleepAggregatedDescription, err = describeChartContent(&sleepAggregatedChart.BaseConfiguration, "bar chart",
			"This chart contains 2 series: the time spent asleep, and the sleep time spent awake during the night. The sum of the 2 values along the same axis, give the total time spent in bed"); err != nil {
			sleepAggregatedDescription = "Failed to generate description: " + err.Error()
		}

	}()

	wg.Wait()

	// render without .html = use the master layout
	return c.Render(http.StatusOK, "dashboard/dashboard", echo.Map{
		"title": "Dashboard - FitSleepInsights",

		"sleepEfficiencyChart":       renderChart(sleepEfficiencyChart),
		"sleepEfficiencyDescription": sleepEfficiencyDescription,

		"dailyStepsCountChart":       renderChart(dailyStepChart),
		"dailyStepsCountDescription": dailyStepsCountDescription,

		"sleepAggregatedChart":       renderChart(sleepAggregatedChart),
		"sleepAggregatedDescription": sleepAggregatedDescription,

		"activityCalendars":            activityCalendars,
		"activityCalendarsDescription": activityCalendarsDescriptions,
		"activityStatistics":           activityStatistics,

		"isLoggedIn": true,

		"isWeekly":  calendarType == WeeklyCalendar,
		"isMonthly": calendarType == MonthlyCalendar,
		"isYearly":  calendarType == YearlyCalendar,

		"nextWeek":  endDate.AddDate(0, 0, 1).Format(time.DateOnly),
		"prevWeek":  startDate.AddDate(0, 0, -1).Format(time.DateOnly),
		"nextMonth": endDate.AddDate(0, 1, 0).Format("2006-01"),
		"prevMonth": startDate.AddDate(0, -1, 0).Format("2006-01"),
		"nextYear":  endDate.AddDate(1, 0, 0).Format("2006"),
		"prevYear":  startDate.AddDate(-1, 0, 0).Format("2006"),

		"currentWeek":  startDate.Format(time.DateOnly),
		"currentMonth": startDate.Format("2006-01"),
		"currentYear":  startDate.Format("2006"),
	})
}

func GetStartDayOfWeek(tm time.Time) time.Time { //get monday 00:00:00
	weekday := time.Duration(tm.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	year, month, day := tm.Date()
	currentZeroDay := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return currentZeroDay.Add(-1 * (weekday) * 24 * time.Hour)
}

func WeeklyDashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var user *fitbit_pgdb.AuthorizedUser
		if user, err = getUser(c); err != nil {
			return err
		}

		var dayOfTheWeek, endDate time.Time
		if c.Param("year") != "" && c.Param("month") != "" && c.Param("day") != "" {
			if dayOfTheWeek, err = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", c.Param("year"), c.Param("month"), c.Param("day"))); err != nil {
				return err
			}
		} else {
			dayOfTheWeek = time.Now()
		}

		startDate := GetStartDayOfWeek(dayOfTheWeek)
		endDate = startDate.AddDate(0, 0, 7)
		return dashboard(c, user, startDate, endDate, WeeklyCalendar)
	}
}

func MonthlyDashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var user *fitbit_pgdb.AuthorizedUser
		if user, err = getUser(c); err != nil {
			return err
		}
		var startDate, endDate time.Time
		if c.Param("year") != "" && c.Param("month") != "" {
			if startDate, err = time.Parse("2006-01", fmt.Sprintf("%s-%s", c.Param("year"), c.Param("month"))); err != nil {
				return err
			}
		} else {
			startDate = time.Now().AddDate(0, -1, 0)
		}
		endDate = startDate.AddDate(0, 1, -1)
		return dashboard(c, user, startDate, endDate, MonthlyCalendar)
	}
}

func YearlyDashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var user *fitbit_pgdb.AuthorizedUser
		if user, err = getUser(c); err != nil {
			return err
		}

		var startDate, endDate time.Time
		if c.Param("year") != "" {
			if startDate, err = time.Parse("2006", c.Param("year")); err != nil {
				return err
			}
		} else {
			startDate = time.Now().AddDate(-1, 0, 0)
		}
		endDate = startDate.AddDate(1, 0, -1)
		return dashboard(c, user, startDate, endDate, YearlyCalendar)
	}
}
