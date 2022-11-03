// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /activities/goals/%s.json

type UserGoal struct {
	Goals Goal `json:"goals"`
}

type Goal struct {
	ActiveMinutes int64   `json:"activeMinutes,omitempty"`
	CaloriesOut   int64   `json:"caloriesOut,omitempty"`
	Distance      float64 `json:"distance"`
	Steps         int64   `json:"steps"`
}

// /activities/list.json?afterDate=2022-10-29&sort=asc&offset=0&limit=2

type ActivityLogList struct {
	Activities []ActivityLog `json:"activities"`
	Pagination Pagination    `json:"pagination"`
}

type ActivityLog struct {
	ActiveDuration        int64                 `json:"activeDuration"`
	ActiveZoneMinutes     ActiveZoneMinutes     `json:"activeZoneMinutes"`
	ActivityLevel         []LoggedActivityLevel `json:"activityLevel"`
	ActivityName          string                `json:"activityName"`
	ActivityTypeID        int64                 `json:"activityTypeId"`
	AverageHeartRate      int64                 `json:"averageHeartRate"`
	Calories              int64                 `json:"calories"`
	Distance              float64               `json:"distance"`
	DistanceUnit          string                `json:"distanceUnit"`
	Duration              int64                 `json:"duration"`
	ElevationGain         int64                 `json:"elevationGain"`
	HasActiveZoneMinutes  bool                  `json:"hasActiveZoneMinutes"`
	HeartRateLink         string                `json:"heartRateLink"`
	HeartRateZones        []HeartRateZone       `json:"heartRateZones"`
	LastModified          string                `json:"lastModified"`
	LogID                 int64                 `json:"logId"`
	LogType               string                `json:"logType"`
	ManualValuesSpecified ManualValuesSpecified `json:"manualValuesSpecified"`
	OriginalDuration      int64                 `json:"originalDuration"`
	OriginalStartTime     FitbitDateTime        `json:"originalStartTime"`
	Pace                  float64               `json:"pace"`
	Source                LogSource             `json:"source"`
	Speed                 float64               `json:"speed"`
	StartTime             FitbitDateTime        `json:"startTime"`
	Steps                 int64                 `json:"steps"`
	TcxLink               string                `json:"tcxLink"`
}

type Pagination struct {
	AfterDate  FitbitDateTime `json:"afterDate,omitempty"`
	BeforeDate FitbitDateTime `json:"beforeDate,omitempty"`
	Limit      int64          `json:"limit"`
	Next       string         `json:"next"`
	Offset     int64          `json:"offset"`
	Previous   string         `json:"previous"`
	Sort       string         `json:"sort"`
}

type ManualValuesSpecified struct {
	Calories bool `json:"calories"`
	Distance bool `json:"distance"`
	Steps    bool `json:"steps"`
}

type HeartRateZone struct {
	CaloriesOut float64 `json:"caloriesOut"`
	Max         int64   `json:"max"`
	Min         int64   `json:"min"`
	Minutes     int64   `json:"minutes"`
	Name        string  `json:"name"`
}

type LogSource struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	TrackerFeatures []string `json:"trackerFeatures"`
	Type            string   `json:"type"`
	URL             string   `json:"url"`
}

type MinutesInHeartRateZone struct {
	MinuteMultiplier int64  `json:"minuteMultiplier"`
	Minutes          int64  `json:"minutes"`
	Order            int64  `json:"order"`
	Type             string `json:"type"`
	ZoneName         string `json:"zoneName"`
}

type LoggedActivityLevel struct {
	Minutes int64  `json:"minutes"`
	Name    string `json:"name"`
}

type ActiveZoneMinutes struct {
	MinutesInHeartRateZones []MinutesInHeartRateZone `json:"minutesInHeartRateZones"`
	TotalMinutes            int64                    `json:"totalMinutes"`
}

// /activities/date/%s.json

type DailyActivitySummary struct {
	Activities []ActivitiesSummary `json:"activities"`
	Goals      Goal                `json:"goals"`
	Summary    ActivitiesSummary   `json:"summary"`
}

type ActivitiesSummary struct {
	ActiveScore          int64           `json:"activeScore"`
	ActivityCalories     int64           `json:"activityCalories"`
	CaloriesBMR          int64           `json:"caloriesBMR"`
	CaloriesOut          int64           `json:"caloriesOut"`
	Distances            []Distance      `json:"distances"`
	FairlyActiveMinutes  int64           `json:"fairlyActiveMinutes"`
	HeartRateZones       []HeartRateZone `json:"heartRateZones"`
	LightlyActiveMinutes int64           `json:"lightlyActiveMinutes"`
	MarginalCalories     int64           `json:"marginalCalories"`
	RestingHeartRate     int64           `json:"restingHeartRate"`
	SedentaryMinutes     int64           `json:"sedentaryMinutes"`
	Steps                int64           `json:"steps"`
	VeryActiveMinutes    int64           `json:"veryActiveMinutes"`
}

type Distance struct {
	Activity string  `json:"activity"`
	Distance float64 `json:"distance"`
}

// activities.json

type UserLifeTimeStats struct {
	Best     BestStatsSource     `json:"best"`
	Lifetime LifetimeStatsSource `json:"lifetime"`
}

type LifeTimeStats struct {
	ActiveScore int64   `json:"activeScore"`
	CaloriesOut int64   `json:"caloriesOut"`
	Distance    float64 `json:"distance"`
	Steps       int64   `json:"steps"`
	Floors      int64   `json:"floors"`
}

type LifeTimeTimeStep struct {
	Date  FitbitDate `json:"date"`
	Value float64    `json:"value"`
}

type LifeTimeActivities struct {
	Distance LifeTimeTimeStep `json:"distance"`
	Steps    LifeTimeTimeStep `json:"steps"`
	Floors   LifeTimeTimeStep `json:"floors"`
}

type BestStatsSource struct {
	Total   LifeTimeActivities `json:"total"`
	Tracker LifeTimeActivities `json:"tracker"`
}

type LifetimeStatsSource struct {
	Total   LifeTimeStats `json:"total"`
	Tracker LifeTimeStats `json:"tracker"`
}

// /activities/favorite.json
type FavoriteActivities []FavoriteActivity

type FavoriteActivity struct {
	ActivityID  int64  `json:"activityId"`
	Description string `json:"description"`
	Mets        int64  `json:"mets"`
	Name        string `json:"name"`
}

type MinimalActivity struct {
	ActivityID  int64   `json:"activityId"`
	Calories    int64   `json:"calories"`
	Description string  `json:"description"`
	Distance    float64 `json:"distance"`
	Duration    int64   `json:"duration"`
	Name        string  `json:"name"`
}

// /activities/frequent.json
type FrequentActivities []MinimalActivity

// /activities/recent.json
type RecentActivities []MinimalActivity
