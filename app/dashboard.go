package app

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"cloud.google.com/go/vertexai/genai"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/galeone/fitbit/v2"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/labstack/echo/v4"
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
	var activitiesTypes []UserActivityTypes
	if activitiesTypes, err = fetcher.UserActivityTypes(); err != nil {
		return err
	}

	activityCalendars := make(map[string]template.HTML)
	activityCalendarsDescriptions := make(map[string]string)
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
			chart := activityCalendar(user, &activityList)
			chart.Renderer = newChartRenderer(chart, chart.Validate)
			activityCalendars[activityType.Name] = renderChart(chart)
			var activityCalendarDescription string
			/*
				if activityCalendarDescription, err = describeChartContent(&activityCalendars[activityType.Name].BaseConfiguration, "calendar heatmap", fmt.Sprintf("Calendar for %s", activityType.Name)); err != nil {
					activityCalendarDescription = "Failed to generate description: " + err.Error()
				}
			*/
			activityCalendarsDescriptions[activityType.Name] = activityCalendarDescription
		}
	}

	dailyStepChart := dailyStepCount(user, allData)
	dailyStepChart.Renderer = newChartRenderer(dailyStepChart, dailyStepChart.Validate)
	var dailyStepsCountDescription string
	/*
		if dailyStepsCountDescription, err = describeChartContent(&dailyStepChart.BaseConfiguration, "calendar heatmap"); err != nil {
			dailyStepsCountDescription = "Failed to generate description: " + err.Error()
		}
	*/

	sleepEfficiencyChart := sleepEfficiencyChart(user, allData)
	sleepEfficiencyChart.Renderer = newChartRenderer(sleepEfficiencyChart, sleepEfficiencyChart.Validate)
	var sleepEfficiencyDescription string
	/*
		if sleepEfficiencyDescription, err = describeChartContent(&sleepEfficiencyChart.BaseConfiguration, "line chart", "Sleep Efficiency is a value in [0,100] computed as the ratio between the time spent in bed and the time effectively spent asleep"); err != nil {
			sleepEfficiencyDescription = "Failed to generate description: " + err.Error()
		}
	*/

	sleepAggregatedChart := sleepAggregatedStackedBarChart(user, allData)
	sleepAggregatedChart.Renderer = newChartRenderer(sleepAggregatedChart, sleepAggregatedChart.Validate)
	var sleepAggregatedDescription string
	/*
		if sleepAggregatedDescription, err = describeChartContent(&sleepAggregatedChart.BaseConfiguration, "bar chart",
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

		"activityCalendars":            activityCalendars,
		"activityCalendarsDescription": activityCalendarsDescriptions,
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
		return dashboard(c, user, startDate, endDate)
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
