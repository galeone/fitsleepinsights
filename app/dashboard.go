package app

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/galeone/fitbit"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v2"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/labstack/echo/v4"
)

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

		// 2. Create the dashboard
		// 2.1. Create the sleep efficiency chart

		var dates []time.Time
		var sleepEfficiency []opts.LineData
		for _, dayData := range all {
			dates = append(dates, dayData.Date)
			sleepEfficiency = append(sleepEfficiency, opts.LineData{Value: dayData.SleepLog.Efficiency})
		}
		fmt.Println(dates[0])
		fmt.Println(len(dates))
		fmt.Println(dates[len(dates)-1])

		chart := charts.NewLine()
		chart.Renderer = newChartRenderer(chart, chart.Validate)

		chart.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title:    "Sleep Efficiency",
				Subtitle: "Day by day",
			}),
			charts.WithInitializationOpts(opts.Initialization{
				Theme: "dark",
			}),
		)
		chart.SetXAxis(dates)

		chart.AddSeries("Actual", sleepEfficiency, charts.WithLineChartOpts(opts.LineChart{
			Smooth: true,
		}))
		// TODO add predicted

		// render without .html = use the master layout
		return c.Render(http.StatusOK, "dashboard", echo.Map{
			"title":                "Dashboard - FitSleepInsights",
			"sleepEfficiencyChart": renderChart(chart),
		})
	}
}
