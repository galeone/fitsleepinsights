package app

import (
	"fmt"
	"net/http"
	"time"

	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/galeone/fitbit/v2"
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

func dashboard(c echo.Context, user *fitbit_pgdb.AuthorizedUser, startDate, endDate time.Time) (err error) {
	var fetcher *fetcher
	if fetcher, err = NewFetcher(user); err != nil {
		return err
	}

	allData := fetcher.FetchByRange(startDate, endDate)

	dailyStepChart := dailyStepCount(user, allData)
	dailyStepChart.Renderer = newChartRenderer(dailyStepChart, dailyStepChart.Validate)
	var dailyStepsCountDescription string
	/*if dailyStepsCountDescription, err = describeChartContent(&dailyStepChart.BaseConfiguration, "calendar heatmap"); err != nil {
		dailyStepsCountDescription = "Failed to generate description: " + err.Error()
	}
	*/

	sleepEfficiencyChart := sleepEfficiencyChart(user, allData)
	sleepEfficiencyChart.Renderer = newChartRenderer(sleepEfficiencyChart, sleepEfficiencyChart.Validate)
	var sleepEfficiencyDescription string
	/*if sleepEfficiencyDescription, err = describeChartContent(&sleepEfficiencyChart.BaseConfiguration, "line chart", "Sleep Efficiency is a value in [0,100] computed as the ratio between the time spent in bed and the time effectively spent asleep"); err != nil {
		sleepEfficiencyDescription = "Failed to generate description: " + err.Error()
	}
	*/

	sleepAggregatedChart := sleepAggregatedStackedBarChart(user, allData)
	sleepAggregatedChart.Renderer = newChartRenderer(sleepAggregatedChart, sleepAggregatedChart.Validate)
	var sleepAggregatedDescription string
	/*if sleepAggregatedDescription, err = describeChartContent(&sleepAggregatedChart.BaseConfiguration, "bar chart",
		"This chart contains 2 series: the time spent asleep, and the sleep time spent awake during the night. The sum of the 2 values along the same axis, give the total time spent in bed"); err != nil {
		sleepAggregatedDescription = "Failed to generate description: " + err.Error()
	}
	*/

	// render without .html = use the master layout
	return c.Render(http.StatusOK, "dashboard/dashboard", echo.Map{
		"title": "Dashboard - FitSleepInsights",

		"sleepEfficiencyChart":       renderChart(sleepEfficiencyChart),
		"sleepEfficiencyDescription": sleepEfficiencyDescription,

		"dailyStepsCountChart":       renderChart(dailyStepChart),
		"dailyStepsCountDescription": dailyStepsCountDescription,

		"sleepAggregatedChart":       renderChart(sleepAggregatedChart),
		"sleepAggregatedDescription": sleepAggregatedDescription,
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

		if c.Param("year") != "" && c.Param("month") != "" && c.Param("day") != "" {
			var dayOfTheWeek, endDate time.Time
			if dayOfTheWeek, err = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", c.Param("year"), c.Param("month"), c.Param("day"))); err != nil {
				return err
			}

			startDate := GetStartDayOfWeek(dayOfTheWeek)
			endDate = startDate.AddDate(0, 0, 7)
			return dashboard(c, user, startDate, endDate)
		}

		return dashboard(c, user, time.Now().AddDate(0, 0, -7), time.Now())
	}
}

func MonthlyDashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var user *fitbit_pgdb.AuthorizedUser
		if user, err = getUser(c); err != nil {
			return err
		}

		if c.Param("year") != "" && c.Param("month") != "" {
			var startDate, endDate time.Time
			if startDate, err = time.Parse("2006-01", fmt.Sprintf("%s-%s", c.Param("year"), c.Param("month"))); err != nil {
				return err
			}
			endDate = startDate.AddDate(0, 1, -1)
			return dashboard(c, user, startDate, endDate)
		}

		return dashboard(c, user, time.Now().AddDate(0, -1, +1), time.Now())
	}
}

func YearlyDashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var user *fitbit_pgdb.AuthorizedUser
		if user, err = getUser(c); err != nil {
			return err
		}

		if c.Param("year") != "" {
			var startDate, endDate time.Time
			if startDate, err = time.Parse("2006", c.Param("year")); err != nil {
				return err
			}
			endDate = startDate.AddDate(1, 0, -1)
			return dashboard(c, user, startDate, endDate)
		}

		return dashboard(c, user, time.Now().AddDate(-1, 0, +1), time.Now())
	}
}
