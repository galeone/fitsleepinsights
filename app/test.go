// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"fmt"
	"time"

	"github.com/galeone/fitbit"
	"github.com/galeone/fitbit/client"
	"github.com/galeone/fitbit/types"
	"github.com/galeone/tcx"
	"github.com/labstack/echo/v4"
)

func TestGET() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		// secure, under middelware
		authorizer := c.Get("fitbit").(*fitbit.Authorizer)

		var fb *client.Client
		if fb, err = client.NewClient(authorizer); err != nil {
			fmt.Println(1)
			return
		}
		var logs *types.ActivityLogList
		if logs, err = fb.UserActivityLogList(&types.Pagination{
			Offset:     0,
			BeforeDate: types.FitbitDateTime{Time: time.Now()},
			Limit:      10,
			Sort:       "desc",
		}); err != nil {
			return
		}

		for _, activity := range logs.Activities {
			if activity.TcxLink != "" {
				fmt.Println("GPS tracked activity")
				var tcxDB *tcx.TCXDB
				if tcxDB, err = fb.UserActivityTCX(activity.LogID); err != nil {
					return
				}
				if tcxDB.Acts != nil {
					fmt.Println("activities found")

					for _, v := range tcxDB.Acts.Act {
						for i, lap := range v.Laps {
							fmt.Printf("Lap: %d.\n%v\n", i, lap)
						}
					}
				}
			}
		}

		var series *types.ActivityCaloriesSeries
		if series, err = fb.UserActivityCaloriesTimeseries(&yesterday, &now); err != nil {
			return
		}
		fmt.Println("yesterday - today")
		for day := range series.TimeSeries {
			fmt.Println(series.TimeSeries[day])
		}

		fmt.Println("only today")
		if series, err = fb.UserActivityCaloriesTimeseries(&now, nil); err != nil {
			return
		}
		for day := range series.TimeSeries {
			fmt.Println(series.TimeSeries[day])
		}

		var intradayCalories *types.CaloriesSeriesIntraday
		twentyMinAgo := now.Add(-20 * time.Minute)
		if intradayCalories, err = fb.UserCaloriesIntraday(&twentyMinAgo, &now); err != nil { //, &now); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(intradayCalories)

		var brIntraday *types.BreathingRateIntraday
		if brIntraday, err = fb.UserBreathingRateIntraday(&now, nil); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(brIntraday)

		var hrIntraday *types.HeartRateIntraday
		if hrIntraday, err = fb.UserHeartRateIntraday(&now, nil); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(hrIntraday)

		var hrvIntraday *types.HeartRateVariabilityIntraday
		if hrvIntraday, err = fb.UserHeartRateVariabilityIntraday(&now, nil); err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println(hrvIntraday)

		var spo2Intraday *types.OxygenSaturationIntraday
		if spo2Intraday, err = fb.UserOxygenSaturationIntraday(&now, nil); err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println(spo2Intraday)

		return nil
	}

}
