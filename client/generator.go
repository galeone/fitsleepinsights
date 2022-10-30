// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package client handles the client subdomain.
// The client subdomain is the subdomain that interacts
// with the Fitbit API

package client

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/galeone/rts"
	"github.com/galeone/sleepbit/fitbit"
	"github.com/labstack/echo/v4"

	_ "github.com/joho/godotenv/autoload"
)

const APIURL string = "https://api.fitbit.com/1"

func UserAPI(userID, endpoint string) string {
	return fmt.Sprintf("%s/user/%s/%s", APIURL, userID, strings.TrimLeft(endpoint, "/"))
}

func FitbitAPI(endpoint string) string {
	return fmt.Sprintf("%s/%s", APIURL, strings.TrimLeft(endpoint, "/"))
}

// GenerateTypes is an internal use endpoint
// That allows to have a Go representation of the JSON responses
// for all the GET endpoints of the Fitbit API.
// It doesn't use the fitbitClient methods, because it has been used
// to create the types used by the fitbitClient itself.
func GenerateTypes() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Safe because of middleware
		fitbitClient := c.Get("fitbit").(*fitbit.FitbitClient)
		user, _ := fitbitClient.UserID()
		//activityID := "90009" // from "activityTypeId"

		req, err := fitbitClient.HTTP()
		if err != nil {
			return err
		}

		yesterday := time.Now().Add(time.Hour * -24).Format("2006-01-02")
		//today := time.Now().Format("2006-01-02")

		// https://dev.fitbit.com/build/reference/web-api/activity/
		paths := []string{
			/*
				"/activities/goals/%s.json",
				"/activities/goals/%s.json",
				fmt.Sprintf("/activities/list.json?afterDate=%s&sort=asc&offset=0&limit=2", yesterday),
				// TODO: /1/user/[user-id]/activities/[log-id].tcx <-- IS XML!
				"/activities.json",
				"/activities/%s.json",
				"/activities/date/%s.json",
				"/activities/favorite.json",
				"/activities/frequent.json",
				// repeat for all resources: https://dev.fitbit.com/build/reference/web-api/activity-timeseries/get-activity-timeseries-by-date/#Resource-Options
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				"/activities/%s/date/%s/%s.json",
				// There's also the same thing ^ but with both start and end date. The answer does not change
				"/body/log/%s/goal.json",
				"/body/log/%s/goal.json",
				"/body/log/fat/date/%s.json",
				"/body/log/weight/date/%s.json",
			*/
			//"activities.json",
			//"/activities/recent.json",
			// manually added responses for body time series
			/*
				"/br/date/%s.json",
				"/cardioscore/date/%s.json",
				"/activities/heart/date/%s/%s.json",
				"hrv/date/%s.json",
			*/
			// https://dev.fitbit.com/build/reference/web-api/intraday/
			// Intraday requires a dedicated request TODO

			// TODO: nutrition & nutrition time series. I have no data for generating meaningful responses

			"/spo2/date/%s.json",
			"/temp/core/date/%s.json",
			"/temp/skin/date/%s.json",
		}
		uriArgs := [][]any{
			/*
				{"daily"},
				{"weekly"},
				{},
				{},
				{activityID},
				{yesterday},
				{},
				{},
				// Set as period 1 day, ending yesterday
				// The available periods are 1d | 7d | 30d | 1w | 1m | 3m | 6m | 1y
				// but I hope the response won't change
				{"activityCalories", yesterday, "1d"},
				{"calories", yesterday, "1d"},
				{"caloriesBMR", yesterday, "1d"},
				{"distance", yesterday, "1d"},
				{"elevation", yesterday, "1d"},
				{"floors", yesterday, "1d"},
				{"minutesSedentary", yesterday, "1d"},
				{"minutesLightlyActive", yesterday, "1d"},
				{"minutesFairlyActive", yesterday, "1d"},
				{"minutesVeryActive", yesterday, "1d"},
				{"steps", yesterday, "1d"},
				// end of the 11

				// body goal (2)
				{"weight"},
				{"fat"},
				{today},
				{today},
			*/
			//{},
			//{},
			// manually added responses for body time series
			/*
				{today},
				{today},
				{today, "1d"},
				{yesterday},
			*/

			{yesterday},

			{yesterday},
			{yesterday},
		}
		isUser := []bool{
			/*
				true,
				true,
				true,
				false, // it's not user -> it's the description of an activity given an id (e.g 90009 -> running in general)
				false,
				true,
				true,
				true,
				true,
				// repeated 11 times
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				// end of the 11
				true,
				true,
				true,
			*/
			//true,
			//true,
			// manually added responses for body time series
			/*
				true,
				true,
				true,
				true,
			*/

			true,
			true,
			true,
		}

		if len(paths) != len(isUser) || len(paths) != len(uriArgs) {
			fmt.Println(len(paths), len(isUser), len(uriArgs))
			panic("check the config")
		}

		var userFile *os.File
		if userFile, err = os.Create("user_generated.go"); err != nil {
			return err
		}
		defer userFile.Close()

		var genericFile *os.File
		if genericFile, err = os.Create("generic_generated.go"); err != nil {
			return err
		}
		defer genericFile.Close()

		var userSb, genericSb strings.Builder
		for i := range paths {
			path := paths[i]
			args := uriArgs[i]
			userReq := isUser[i]
			var res *http.Response
			var dest string
			if userReq {
				dest = UserAPI(*user, fmt.Sprintf(path, args...))
			} else {
				dest = FitbitAPI(fmt.Sprintf(path, args...))
			}
			fmt.Println(dest)
			if res, err = req.Get(dest); err != nil {
				return err
			}
			var body []byte

			if body, err = io.ReadAll(res.Body); err != nil {
				return err
			}
			var pack []byte
			bodyString := string(body)
			if pack, err = rts.DoRaw("types", bodyString); err != nil {
				return err
			}
			if userReq {
				userSb.WriteString(fmt.Sprintf("// %s\n", path))
				userSb.Write(pack)
			} else {
				genericSb.WriteString(fmt.Sprintf("// %s\n", path))
				genericSb.Write(pack)
			}
			fmt.Println(bodyString)
			time.Sleep(time.Second * 1)
		}

		userFile.WriteString(userSb.String())
		genericFile.WriteString(genericSb.String())
		return nil
	}
}
