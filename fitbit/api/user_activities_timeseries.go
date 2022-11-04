// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/galeone/sleepbit/fitbit/types"
)

func (c *API) userActivityTimeseriesByRange(resource string, startDate, endDate *time.Time) (ret interface{}, err error) {
	var res *http.Response
	hasEndDate := endDate != nil && !endDate.IsZero()

	var path string
	// Same route, but with a period of 1d instead of and end date
	if hasEndDate {
		// GET: /1/user/[user-id]/activities/[resource-path]/date/[start-date]/[end-date].json
		path = fmt.Sprintf("/activities/%s/date/%s/%s.json", resource, startDate.Format(types.DateLayout), endDate.Format(types.DateLayout))
	} else {
		// GET: /1/user/[user-id]/activities/[resource-path]/date/[date]/[period].json
		path = fmt.Sprintf("/activities/%s/date/%s/%s.json", resource, startDate.Format(types.DateLayout), types.Period1Day)
	}
	if res, err = c.req.Get(UserV1(path)); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	// https://dev.fitbit.com/build/reference/web-api/activity-timeseries/get-activity-timeseries-by-date-range/#Resource-Options
	switch resource {
	case "activityCalories":
		ret = &types.ActivityCaloriesSeries{}
	case "calories":
		ret = &types.CaloriesSeries{}
	case "caloriesBMR":
		ret = &types.CaloriesBMRSeries{}
	case "distance":
		ret = &types.DistanceSeries{}
	case "elevation":
		ret = &types.ElevationSeries{}
	case "floors":
		ret = &types.FloorsSeries{}
	case "minutesSedentary":
		ret = &types.MinutesSedentarySeries{}
	case "minutesLightlyActive":
		ret = &types.MinutesLightlyActiveSeries{}
	case "minutesFairlyActive":
		ret = &types.MinutesFailryActiveSeries{}
	case "minutesVeryActive":
		ret = &types.MinutesVeryActiveSeries{}
	case "steps":
		ret = &types.StepsSeries{}
	default:
		panic(fmt.Sprintf("resouce %s not supported", resource))
	}
	err = json.Unmarshal(body, ret)
	return
}

// UserActivityCaloriesTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserActivityCaloriesTimeseries(startDate, endDate *time.Time) (ret *types.ActivityCaloriesSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("activityCalories", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.ActivityCaloriesSeries), err
}

// UserCaloriesTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserCaloriesTimeseries(startDate, endDate *time.Time) (ret *types.CaloriesSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("calories", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.CaloriesSeries), err
}

// UserCaloriesBMRTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserCaloriesBMRTimeseries(startDate, endDate *time.Time) (ret *types.CaloriesBMRSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("caloriesBMR", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.CaloriesBMRSeries), err
}

// UserDistanceTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserDistanceTimeseries(startDate, endDate *time.Time) (ret *types.DistanceSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("distance", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.DistanceSeries), err
}

// UserElevationTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserElevationTimeseries(startDate, endDate *time.Time) (ret *types.ElevationSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("elevation", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.ElevationSeries), err
}

// UserFloorsTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserFloorsTimeseries(startDate, endDate *time.Time) (ret *types.FloorsSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("floors", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.FloorsSeries), err
}

// UserMinutesSedentaryTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserMinutesSedentaryTimeseries(startDate, endDate *time.Time) (ret *types.MinutesSedentarySeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("minutesSedentary", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.MinutesSedentarySeries), err
}

// UserMinutesLightlyActiveTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserMinutesLightlyActiveTimeseries(startDate, endDate *time.Time) (ret *types.MinutesLightlyActiveSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("minutesLightlyActive", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.MinutesLightlyActiveSeries), err
}

// UserMinutesFailryActiveTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserMinutesFailryActiveTimeseries(startDate, endDate *time.Time) (ret *types.MinutesFailryActiveSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("minutesFairlyActive", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.MinutesFailryActiveSeries), err
}

// UserMinutesVeryActiveTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserMinutesVeryActiveTimeseries(startDate, endDate *time.Time) (ret *types.MinutesVeryActiveSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("minutesVeryActive", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.MinutesVeryActiveSeries), err
}

// UserStepsTimeseries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserStepsTimeseries(startDate, endDate *time.Time) (ret *types.StepsSeries, err error) {
	var val interface{}
	if val, err = c.userActivityTimeseriesByRange("steps", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.StepsSeries), err
}
