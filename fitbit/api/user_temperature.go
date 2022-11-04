// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/galeone/sleepbit/fitbit/types"
)

// UserCoreTemperature retrievies Temperature (Core) data for a date range.
// Temperature (Core) data applies specifically to data logged manually by the user on a given day.
// It only returns a value for dates on which the Fitbit device was able to record Temperature (Core)
// data and the maximum date range cannot exceed 30 days.
//
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserCoreTemperature(startDate, endDate *time.Time) (ret *types.CoreTemperature, err error) {
	var res *http.Response
	var sb strings.Builder

	// /1/user/[user-id]/temp/core/date/[date].json
	sb.WriteString(fmt.Sprintf("/temp/core/date/%s", startDate.Format(types.DateLayout)))
	if endDate != nil && !endDate.IsZero() {
		// /1/user/[user-id]/temp/core/date/[start-date]/[end-date].json
		sb.WriteString(fmt.Sprintf("/%s", endDate.Format(types.DateLayout)))
	}
	sb.WriteString(".json")
	if res, err = c.req.Get(UserV1(sb.String())); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.CoreTemperature{}
	err = json.Unmarshal(body, ret)
	return
}

// UserSkinTemperature retrievies Temperature (Skin) data for a date range.
// It only returns a value for dates on which the Fitbit device was able to record
// Temperature (skin) data and the maximum date range cannot exceed 30 days.
//
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserSkinTemperature(startDate, endDate *time.Time) (ret *types.SkinTemperature, err error) {
	var res *http.Response
	var sb strings.Builder

	// /1/user/[user-id]/temp/skin/date/[date].json
	sb.WriteString(fmt.Sprintf("/temp/skin/date/%s", startDate.Format(types.DateLayout)))
	if endDate != nil && !endDate.IsZero() {
		// /1/user/[user-id]/temp/skin/date/[start-date]/[end-date].json
		sb.WriteString(fmt.Sprintf("/%s", endDate.Format(types.DateLayout)))
	}
	sb.WriteString(".json")
	if res, err = c.req.Get(UserV1(sb.String())); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.SkinTemperature{}
	err = json.Unmarshal(body, ret)
	return
}
