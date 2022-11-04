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

// UserHeartRateVariability retrieves the Heart Rate Variability (HRV) data for a date range.
// HRV data applies specifically to a user’s “main sleep,” which is the longest single period of time asleep on a given date.
//
// The endDate parameter is optional. When present it returns the summary, day-by-day, from startDate to endDate.
func (c *API) UserHeartRateVariability(startDate, endDate *time.Time) (ret *types.HeartRateVariability, err error) {
	var res *http.Response
	var sb strings.Builder

	// /1/user/[user-id]/hrv/date/[date].json
	sb.WriteString(fmt.Sprintf("/hrv/date/%s", startDate.Format(types.DateLayout)))
	if endDate != nil && !endDate.IsZero() {
		// /1/user/[user-id]/hrv/date/[start-date]/[end-date].json
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
	ret = &types.HeartRateVariability{}
	err = json.Unmarshal(body, ret)
	return
}
