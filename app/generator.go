// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package client handles the client subdomain.
// The client subdomain is the subdomain that interacts
// with the Fitbit API

package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/galeone/fitbit"
	"github.com/galeone/rts"
	"github.com/labstack/echo/v4"

	_ "github.com/joho/godotenv/autoload"
)

const APIURL string = "https://api.fitbit.com/1"

func UserAPI(userID, endpoint string) string {
	return fmt.Sprintf("%s/user/%s/%s", APIURL, userID, strings.TrimLeft(endpoint, "/"))
}

func UserAPIDot2(userID, endpoint string) string {
	return fmt.Sprintf("%s.2/user/%s/%s", APIURL, userID, strings.TrimLeft(endpoint, "/"))
}

func FitbitAPI(endpoint string) string {
	return fmt.Sprintf("%s/%s", APIURL, strings.TrimLeft(endpoint, "/"))
}

// GenerateTypes is an internal use endpoint
// That allows to have a Go representation of the JSON responses
// for all the GET endpoints of the Fitbit API.
// It doesn't use the authorizer methods, because it has been used
// to create the types used by the authorizer itself.
func GenerateTypes() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// Safe because of middleware
		authorizer := c.Get("fitbit").(*fitbit.Authorizer)
		user, _ := authorizer.UserID()
		//activityID := "90009" // from "activityTypeId"

		var req *http.Client
		req, err = authorizer.HTTP()
		if err != nil {
			return err
		}

		yesterday := time.Now().Add(time.Hour * -24).Format("2006-01-02")
		today := time.Now().Format("2006-01-02")

		// https://dev.fitbit.com/build/reference/web-api/activity/
		paths := []string{
			/*
				"/activities/goals/%s.json",
				"/activities/goals/%s.json",
				fmt.Sprintf("/activities/list.json?afterDate=%s&sort=asc&offset=0&limit=2", yesterday),
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
			// TODO: nutrition & nutrition time series. I have no data for generating meaningful responses
			/*
					"/spo2/date/%s.json",
					"/temp/core/date/%s.json",
					"/temp/skin/date/%s.json",

				"/sleep/date/%s.json",
				"/sleep/date/%s/%s.json",
				fmt.Sprintf("/sleep/list.json?afterDate=%s&sort=asc&offset=0&limit=2", yesterday),
				"/sleep/goal.json",
			*/

			// Intraday

			// /1/user/[user-id]/activities/[resource]/date/[start-date][end-date]/[detail-level]/time/[start-time]/[end-time].json
			// resource in calories | distance | elevation | floors | steps
			// datail level in 1min | 1min (using 1min, hope 1min has the same response)
			"/activities/%s/date/%s/%s/%s/time/%s/%s.json",
			"/activities/%s/date/%s/%s/%s/time/%s/%s.json",
			"/activities/%s/date/%s/%s/%s/time/%s/%s.json",
			"/activities/%s/date/%s/%s/%s/time/%s/%s.json",
			"/activities/%s/date/%s/%s/%s/time/%s/%s.json",

			"/br/date/%s/%s/all.json",
			// /1/user/[user-id]/activities/heart/date/[date]/1d/[detail-level]/time/[start-time]/[end-time].json
			// /1/user/[user-id]/activities/heart/date/[start-date]/[end-date]/[detail-level]/time/[start-time]/[end-time].json
			"/activities/heart/date/%s/%s/%s/time/%s/%s.json",
			// /1/user/[user-id]/hrv/date/[startDate]/[endDate]/all.json
			"/hrv/date/%s/%s/all.json",
			// /1/user/[user-id]/spo2/date/[start-date]/[end-date]/all.json
			"/spo2/date/%s/%s/all.json",
		}
		uriArgs := [][]any{
			/*
				{"daily"},
				{"weekly"},
				{},
				{},
				{activityID},
				{today},
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
			/*
					{yesterday},
					{yesterday},
					{yesterday},

				{yesterday},
				{yesterday, today},
				{},
				{},
			*/

			{"calories", today, today, "1min", "8:00", "8:30"},
			{"distance", today, today, "1min", "8:00", "8:30"},
			{"elevation", today, today, "1min", "8:00", "8:30"},
			{"floors", today, today, "1min", "8:00", "8:30"},
			{"steps", today, today, "1min", "8:00", "8:30"},
			{yesterday, today},
			// /1/user/[user-id]/activities/heart/date/[start-date]/[end-date]/[detail-level]/time/[start-time]/[end-time].json
			{yesterday, yesterday, "1min", "8:00", "8:30"},
			{yesterday, today},
			{yesterday, today},
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
			/*
					true,
					true,
					true,
				true,
				true,
				true,
				true,
			*/

			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
			true,
		}

		if len(paths) != len(isUser) || len(paths) != len(uriArgs) {
			log.Println(len(paths), len(isUser), len(uriArgs))
			panic("check the config")
		}
		if len(paths) == 0 {
			return nil
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
				if strings.Contains(path, "/sleep/") {
					dest = UserAPIDot2(*user, fmt.Sprintf(path, args...))
				} else {
					dest = UserAPI(*user, fmt.Sprintf(path, args...))
				}
			} else {
				dest = FitbitAPI(fmt.Sprintf(path, args...))
			}
			log.Println(dest)
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
			log.Println(bodyString)
			time.Sleep(time.Second * 1)
		}

		if _, err = userFile.WriteString(userSb.String()); err != nil {
			log.Println(err.Error())
		}
		if _, err = genericFile.WriteString(genericSb.String()); err != nil {
			log.Println(err.Error())
		}
		return
	}
}
