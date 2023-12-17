package app

import (
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/galeone/fitbit"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v2"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/labstack/echo/v4"
)

func sleepEfficiencyChart(user *fitbit_pgdb.AuthorizedUser, all []*UserData) *charts.Line {
	var dates []string
	var sleepEfficiency []opts.LineData
	for _, dayData := range all {
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
	)
	chart.SetXAxis(dates)

	chart.AddSeries("Actual", sleepEfficiency, charts.WithLineChartOpts(opts.LineChart{
		Smooth: true,
	}))

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
	return chart
}

func Dashboard() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// secure, under middleware
		authorizer := c.Get("fitbit").(*fitbit.Authorizer)
		var userID *string
		if userID, err = authorizer.UserID(); err != nil {
			return err
		}

		user := fitbit_pgdb.AuthorizedUser{}
		user.UserID = *userID
		if err = _db.Model(fitbit_pgdb.AuthorizedUser{}).Where(&user).Scan(&user); err != nil {
			return err
		}

		var fetcher *fetcher
		if fetcher, err = NewFetcher(&user); err != nil {
			return err
		}

		var all []*UserData
		if all, err = fetcher.FetchAll(FetchAllWithSleepLog); err != nil {
			return err
		}

		slices.Reverse(all)

		chart := sleepEfficiencyChart(&user, all)
		chart.Renderer = newChartRenderer(chart, chart.Validate)

		// render without .html = use the master layout
		return c.Render(http.StatusOK, "dashboard", echo.Map{
			"title":                "Dashboard - FitSleepInsights",
			"sleepEfficiencyChart": renderChart(chart),
		})
	}
}
