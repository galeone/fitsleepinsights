package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/galeone/sleepbit/fitbit/types"
	"github.com/galeone/tcx"
)

func (c *API) userActivityGoal(period string) (ret *types.UserGoal, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1(fmt.Sprintf("/activities/goals/%s.json", period))); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.UserGoal{}
	err = json.Unmarshal(body, ret)
	return
}

// UserActivityDailyGoal retrieves the user current daily activity goal.
//
// GET: /1/user/[user-id]/activities/goals/daily.json
func (c *API) UserActivityDailyGoal() (ret *types.UserGoal, err error) {
	return c.userActivityGoal("daily")
}

// UserActivityWeeklyGoal retrieves the user weekly activity goals.
//
// GET: /1/user/[user-id]/activities/goals/weekly.json
func (c *API) UserActivityWeeklyGoal() (ret *types.UserGoal, err error) {
	return c.userActivityGoal("weekly")
}

// UserActivityLogList retrieves a list of a user's activity log entries before or after a given day.
//
// GET: /1/user/[user-id]/activities/list.json
func (c *API) UserActivityLogList(pagination *types.Pagination) (ret *types.ActivityLogList, err error) {
	var sb strings.Builder
	sb.WriteString("/activities/list.json?sort=")
	sb.WriteString(pagination.Sort)
	sb.WriteString("&offset=")
	sb.WriteString(strconv.Itoa(int(pagination.Offset)))
	sb.WriteString("&limit=")
	sb.WriteString(strconv.Itoa(int(pagination.Limit)))

	if !pagination.BeforeDate.IsZero() {
		sb.WriteString("&beforeDate=")
		sb.WriteString(url.QueryEscape(pagination.BeforeDate.Format(types.DateTimeLayout)))
	}

	if !pagination.AfterDate.IsZero() {
		sb.WriteString("&afterDate=")
		sb.WriteString(url.QueryEscape(pagination.AfterDate.Format(types.DateTimeLayout)))
	}

	path := UserV1(sb.String())

	var res *http.Response
	if res, err = c.req.Get(path); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.ActivityLogList{}
	err = json.Unmarshal(body, ret)
	return
}

// UserActivityTCX retrieves the details of a user's location using GPS and heart rate data during a logged exercise.
//
// GET: /1/user/[user-id]/activities/[log-id].tcx
func (c *API) UserActivityTCX(activityLogID int64) (ret *tcx.TCXDB, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1(fmt.Sprintf("/activities/%d.tcx?includePartialTCX=true", activityLogID))); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &tcx.TCXDB{}
	err = xml.Unmarshal(body, ret)
	return
}

// UserDailyActivitySummary retrieves a summary and list of a user’s activities and activity log entries for a given day.
//
// GET: /1/user/[user-id]/activities/date/[date].json
func (c *API) UserDailyActivitySummary(date *time.Time) (ret *types.DailyActivitySummary, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1(fmt.Sprintf("/activities/date/%s.json", date.Format(types.DateLayout)))); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.DailyActivitySummary{}
	err = json.Unmarshal(body, ret)

	return
}

// UserFavoriteActivities retrieves a list of a user's favorite activities.
//
// GET: /1/user/[user-id]/activities/favorite.json
func (c *API) UserFavoriteActivities() (ret *types.FavoriteActivities, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1("/activities/favorite.json")); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.FavoriteActivities{}
	err = json.Unmarshal(body, ret)
	return
}

// UserFrequentActivities retrieves a list of a user's frequent activities.
//
// GET: /1/user/[user-id]/activities/frequent.json
func (c *API) UserFrequentActivities() (ret *types.FrequentActivities, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1("/activities/frequent.json")); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.FrequentActivities{}
	err = json.Unmarshal(body, ret)
	return
}

// UserRecentActivities retrieves a list of a user's recent activities.
//
// GET: /1/user/[user-id]/activities/recent.json
func (c *API) UserRecentActivities() (ret *types.RecentActivities, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1("/activities/recent.json")); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.RecentActivities{}
	err = json.Unmarshal(body, ret)
	return
}

// UserLifetimeStats retrieves the user's activity statistics..
//
// GET: /1/user/[user-id]/activities.json
func (c *API) UserLifetimeStats() (ret *types.UserLifeTimeStats, err error) {
	var res *http.Response
	if res, err = c.req.Get(UserV1("/activities/favorite.json")); err != nil {
		return
	}
	var body []byte
	if body, err = c.resRead(res); err != nil {
		return
	}
	ret = &types.UserLifeTimeStats{}
	err = json.Unmarshal(body, ret)
	return
}
