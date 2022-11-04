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

func (c *API) userBodyTimeseriesByRange(resource string, startDate, endDate *time.Time) (ret interface{}, err error) {
	var res *http.Response
	hasEndDate := endDate != nil && !endDate.IsZero()

	var path string
	// Same route, but with a period of 1d instead of and end date
	if hasEndDate {
		// GET: /1/user/[user-id]/body/[resource-path]/date/[start-date]/[end-date].json
		path = fmt.Sprintf("/body/%s/date/%s/%s.json", resource, startDate.Format(types.DateLayout), endDate.Format(types.DateLayout))
	} else {
		// GET: /1/user/[user-id]/body/[resource-path]/date/[date]/[period].json
		path = fmt.Sprintf("/body/%s/date/%s/%s.json", resource, startDate.Format(types.DateLayout), types.Period1Day)
	}
	if res, err = c.req.Get(UserV1(path)); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	// https://dev.fitbit.com/build/reference/web-api/body-timeseries/get-body-timeseries-by-date/
	// Supported: bmi | fat | weight
	switch resource {
	case "bmi":
		ret = &types.BMISeries{}
	case "fat":
		ret = &types.BodyFatSeries{}
	case "weight":
		ret = &types.BodyWeightSeries{}
	default:
		panic(fmt.Sprintf("resouce %s not supported", resource))
	}
	err = json.Unmarshal(body, ret)
	return
}

// UserBMITimeSeries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserBMITimeSeries(startDate, endDate *time.Time) (ret *types.BMISeries, err error) {
	var val interface{}
	if val, err = c.userBodyTimeseriesByRange("bmi", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.BMISeries), err
}

// UserBodyWeightTimeSeries retrieves the activity calories over a period of time by specifying a date range.
// The response will include only the daily summary values.
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserBodyWeightTimeSeries(startDate, endDate *time.Time) (ret *types.BodyWeightSeries, err error) {
	var val interface{}
	if val, err = c.userBodyTimeseriesByRange("weight", startDate, endDate); err != nil {
		return nil, err
	}
	return val.(*types.BodyWeightSeries), err
}
