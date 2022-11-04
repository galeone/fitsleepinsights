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

func (c *API) userBodyGoals(goalType string) (ret interface{}, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1(fmt.Sprintf("/body/log/%s/goal.json", goalType))); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	// https://dev.fitbit.com/build/reference/web-api/body/get-body-goals/
	switch goalType {
	case "weight":
		ret = &types.UserWeightGoal{}
	case "fat":
		ret = &types.UserFatGoal{}
	default:
		panic(fmt.Sprintf("goalType %s not supported", goalType))
	}
	err = json.Unmarshal(body, ret)
	return
}

// UserWeightGoal retrieves the user weight goal
func (c *API) UserWeightGoal() (ret *types.UserWeightGoal, err error) {
	var val interface{}
	if val, err = c.userBodyGoals("weight"); err != nil {
		return nil, err
	}
	return val.(*types.UserWeightGoal), err
}

// UserFatGoal retrieves the user weight goal
func (c *API) UserFatGoal() (ret *types.UserFatGoal, err error) {
	var val interface{}
	if val, err = c.userBodyGoals("fat"); err != nil {
		return nil, err
	}
	return val.(*types.UserFatGoal), err
}

// UserBodyFatLog retrieves a list of all user's weight log entries for a given date.
// GET: /1/user/[user-id]/body/log/fat/date/[date].json
func (c *API) UserBodyFatLog(date *time.Time) (ret *types.BodyFatLog, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1(fmt.Sprintf("/body/log/fat/date/%s.json", date.Format(types.DateLayout)))); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.BodyFatLog{}
	err = json.Unmarshal(body, ret)
	return
}

// UserBodyWeightLog retrieves a list of all user's weight log entries for a given date.
// GET: /1/user/[user-id]/body/log/weight/date/[date].json
func (c *API) UserBodyWeightLog(date *time.Time) (ret *types.BodyWeightLog, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1(fmt.Sprintf("/body/log/weight/date/%s.json", date.Format(types.DateLayout)))); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.BodyWeightLog{}
	err = json.Unmarshal(body, ret)
	return
}
