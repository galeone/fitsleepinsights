package app

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/galeone/fitbit/v2"
	"github.com/galeone/fitsleepinsights/database/types"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func getUser(c echo.Context) (*types.User, error) {
	// secure, under middleware
	var err error
	authorizer := c.Get("fitbit").(*fitbit.Authorizer)
	var userID *string
	if userID, err = authorizer.UserID(); err != nil {
		return nil, err
	}

	user := types.User{}
	user.UserID = *userID
	if err = _db.Model(types.User{}).Where(&user).Scan(&user); err != nil {
		return nil, err
	}
	return &user, err
}

func dashboard(c echo.Context, user *types.User, startDate, endDate time.Time, calendarType CalendarType) (err error) {
	var fetcher *fetcher
	if fetcher, err = NewFetcher(user); err != nil {
		log.Error(err)
		return err
	}

	var allData []*UserData
	if allData, err = fetcher.FetchByRange(startDate, endDate); err != nil {
		var fetcherError *FetcherError
		if errors.As(err, &fetcherError) {
			return c.Render(http.StatusOK, "dashboard/dashboard", echo.Map{
				"title":      "Dashboard - FitSleepInsights",
				"isLoggedIn": true,
				"startDate":  startDate.Format(time.DateOnly),
				"endDate":    endDate.Format(time.DateOnly),
				"dumping":    true,
			})
		} else {
			log.Error(err)
			return err
		}
	}
	var activitiesTypes []UserActivityTypes
	if activitiesTypes, err = fetcher.UserActivityTypes(); err != nil {
		log.Error(err)
		return err
	}

	activityCalendars := make(map[string]template.HTML)
	activityStatistics := make(map[string]*ActivityStats)

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

				mapMux.Lock()
				activityCalendars[activityType.Name] = renderChart(chart)
				activityStatistics[activityType.Name] = activityStats(&activityList)
				mapMux.Unlock()
			}(activityType)
		}
	}

	wg.Wait()

	var dailyStepChart *charts.HeatMap
	var dailyStepsStatistics *DailyStepsStats
	var sleepBoard *SleepDashboard
	var healthBoard *HealthDashboard
	wg.Add(3)

	go func() {
		defer wg.Done()
		dailyStepChart, dailyStepsStatistics = dailyStepCount(allData, calendarType)
		dailyStepChart.Renderer = newChartRenderer(dailyStepChart, dailyStepChart.Validate)
	}()

	go func() {
		defer wg.Done()
		sleepBoard = sleepDashboard(allData, calendarType)
		sleepBoard.AggregatedStages.Renderer = newChartRenderer(sleepBoard.AggregatedStages, sleepBoard.AggregatedStages.Validate)
		sleepBoard.Efficiency.Renderer = newChartRenderer(sleepBoard.Efficiency, sleepBoard.Efficiency.Validate)
		sleepBoard.HeartRateVariabilityDeepSleep.Renderer = newChartRenderer(sleepBoard.HeartRateVariabilityDeepSleep, sleepBoard.HeartRateVariabilityDeepSleep.Validate)
	}()

	go func() {
		defer wg.Done()
		healthBoard = healthDashboard(allData, calendarType)
		//healthBoard.BreathingRate.Renderer = newChartRenderer(healthBoard.BreathingRate, healthBoard.BreathingRate.Validate)
		healthBoard.HeartRateVariability.Renderer = newChartRenderer(healthBoard.HeartRateVariability, healthBoard.HeartRateVariability.Validate)
		healthBoard.OxygenSaturation.Renderer = newChartRenderer(healthBoard.OxygenSaturation, healthBoard.OxygenSaturation.Validate)
		healthBoard.RestingHeartRate.Renderer = newChartRenderer(healthBoard.RestingHeartRate, healthBoard.RestingHeartRate.Validate)
		healthBoard.SkinTemperature.Renderer = newChartRenderer(healthBoard.SkinTemperature, healthBoard.SkinTemperature.Validate)
		healthBoard.BMI.Renderer = newChartRenderer(healthBoard.BMI, healthBoard.BMI.Validate)
		healthBoard.Weight.Renderer = newChartRenderer(healthBoard.Weight, healthBoard.Weight.Validate)
	}()

	wg.Wait()

	// render without .html = use the master layout
	return c.Render(http.StatusOK, "dashboard/dashboard", echo.Map{
		"title":      "Dashboard - FitSleepInsights",
		"isLoggedIn": true,
		"startDate":  startDate.Format(time.DateOnly),
		"endDate":    endDate.Format(time.DateOnly),
		"dumping":    false,

		"sleepEfficiencyChart": renderChart(sleepBoard.Efficiency),
		"sleepAggregatedChart": renderChart(sleepBoard.AggregatedStages),
		"sleepHrvChart":        renderChart(sleepBoard.HeartRateVariabilityDeepSleep),
		"sleepStatistics":      sleepBoard.Stats,

		"dailyStepsCountChart": renderChart(dailyStepChart),
		"dailyStepsStatistics": dailyStepsStatistics,

		"activityCalendars":  activityCalendars,
		"activityStatistics": activityStatistics,

		//"breathingRateChart":        renderChart(healthBoard.BreathingRate),
		"heartRateVariabilityChart": renderChart(healthBoard.HeartRateVariability),
		"oxygenSaturationChart":     renderChart(healthBoard.OxygenSaturation),
		"restingHeartRateChart":     renderChart(healthBoard.RestingHeartRate),
		"skinTemperatureChart":      renderChart(healthBoard.SkinTemperature),
		"bmiChart":                  renderChart(healthBoard.BMI),
		"weightChart":               renderChart(healthBoard.Weight),
		"healthStatistics":          healthBoard.Stats,
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

func startDateEndDateFromParams(c echo.Context) (startDate, endDate time.Time, err error) {
	if c.Param("year") != "" && c.Param("month") != "" && c.Param("day") != "" {
		var dayOfTheWeek time.Time
		dayOfTheWeek, err = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", c.Param("year"), c.Param("month"), c.Param("day")))
		startDate = GetStartDayOfWeek(dayOfTheWeek)
		endDate = startDate.AddDate(0, 0, 7)
	} else if c.Param("year") != "" && c.Param("month") != "" {
		startDate, err = time.Parse("2006-01", fmt.Sprintf("%s-%s", c.Param("year"), c.Param("month")))
		endDate = startDate.AddDate(0, 1, -1)
	} else if c.Param("year") != "" {
		startDate, err = time.Parse("2006", c.Param("year"))
		endDate = startDate.AddDate(1, 0, -1)
	}
	return startDate, endDate, err
}

func WeeklyDashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var user *types.User
		if user, err = getUser(c); err != nil {
			log.Error(err)
			return err
		}

		var startDate, endDate time.Time
		if c.Param("year") != "" && c.Param("month") != "" && c.Param("day") != "" {
			if startDate, endDate, err = startDateEndDateFromParams(c); err != nil {
				log.Error(err)
				return err
			}
		} else {
			startDate = GetStartDayOfWeek(time.Now())
			endDate = startDate.AddDate(0, 0, 7)

		}
		return dashboard(c, user, startDate, endDate, WeeklyCalendar)
	}
}

func MonthlyDashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {

		var user *types.User
		if user, err = getUser(c); err != nil {
			log.Error(err)
			return err
		}
		var startDate, endDate time.Time
		if c.Param("year") != "" && c.Param("month") != "" {
			if startDate, endDate, err = startDateEndDateFromParams(c); err != nil {
				log.Error(err)
				return err
			}
		} else {
			startDate = time.Now().AddDate(0, -1, 0)
			endDate = startDate.AddDate(0, 1, -1)
		}
		return dashboard(c, user, startDate, endDate, MonthlyCalendar)
	}
}

func YearlyDashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var user *types.User
		if user, err = getUser(c); err != nil {
			log.Error(err)
			return err
		}

		var startDate, endDate time.Time

		if c.Param("year") != "" {
			if startDate, endDate, err = startDateEndDateFromParams(c); err != nil {
				log.Error(err)
				return err
			}
		} else {
			startDate = time.Now().AddDate(-1, 0, 0)
			endDate = startDate.AddDate(1, 0, -1)
		}

		return dashboard(c, user, startDate, endDate, YearlyCalendar)
	}
}

func CustomDashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var user *types.User
		if user, err = getUser(c); err != nil {
			log.Error(err)
			return err
		}

		var startDate, endDate time.Time
		if startDate, err = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", c.Param("startYear"), c.Param("startMonth"), c.Param("startDay"))); err != nil {
			log.Error(err)
			return err
		}
		if endDate, err = time.Parse(time.DateOnly, fmt.Sprintf("%s-%s-%s", c.Param("endYear"), c.Param("endMonth"), c.Param("endDay"))); err != nil {
			log.Error(err)
			return err
		}

		calendarType := MonthlyCalendar

		if endDate.Sub(startDate) < 7*24*time.Hour {
			calendarType = WeeklyCalendar
		} else if endDate.Sub(startDate) < 31*24*time.Hour {
			calendarType = MonthlyCalendar
		} else if endDate.Sub(startDate) < 62*24*time.Hour {
			calendarType = BiMonthlyCalendar
		} else if endDate.Sub(startDate) < 93*24*time.Hour {
			calendarType = TriMonthlyCalendar
		} else if endDate.Sub(startDate) < 124*24*time.Hour {
			calendarType = QuadriMonthlyCalendar
		} else if endDate.Sub(startDate) < 155*24*time.Hour {
			calendarType = PentaMonthlyCalendar
		} else if endDate.Sub(startDate) < 186*24*time.Hour {
			calendarType = HexaMonthlyCalendar
		} else if endDate.Sub(startDate) < 365*24*time.Hour {
			calendarType = YearlyCalendar
		}

		return dashboard(c, user, startDate, endDate, calendarType)
	}
}
