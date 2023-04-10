// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"database/sql"
	"time"

	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type Goal struct {
	types.Goal
	ID        int64               `igor:"primary_key"`
	User      pgdb.AuthorizedUser `sql:"-"`
	UserID    int64
	StartDate time.Time
	EndDate   time.Time
}

func (Goal) TableName() string {
	return "goals"
}

type HeartRateZone struct {
	types.HeartRateZone
	ID            int64 `igor:"primary_key"`
	ActivityLogID sql.NullInt64
}

func (HeartRateZone) TableName() string {
	return "heart_rate_zones"
}

type LogSource struct {
	types.LogSource
	// String because the API sets the primary key and
	// for example Fitbit4Android has an alphanumeric ID.
	ID string `igor:"primary_key"`
}

func (LogSource) TableName() string {
	return "log_sources"
}

type MinutesInHeartRateZone struct {
	types.MinutesInHeartRateZone
	ID int64 `igor:"primary_key"`
	// FK
	ActiveZoneMinutesID int64
}

func (MinutesInHeartRateZone) TableName() string {
	return "minutes_in_heart_rate_zone"
}

type LoggedActivityLevel struct {
	types.LoggedActivityLevel
	ID            int64 `igor:"primary_key"`
	ActivityLogID int64
}

func (LoggedActivityLevel) TableName() string {
	return "logged_activity_levels"
}

type ActiveZoneMinutes struct {
	types.ActiveZoneMinutes
	ID                      int64                    `igor:"primary_key"`
	MinutesInHeartRateZones []MinutesInHeartRateZone `sql:"-"`
}

func (ActiveZoneMinutes) TableName() string {
	return "active_zone_minutes"
}

type ActivityLog struct {
	types.ActivityLog
	// Disable all the concrete types and add the corresponding Type+ID field
	// that contains the foreign key value.
	LogID               int64               `igor:"primary_key"`
	User                pgdb.AuthorizedUser `sql:"-"`
	UserID              int64
	ActiveZoneMinutes   ActiveZoneMinutes `sql:"-"`
	ActiveZoneMinutesID sql.NullInt64
	Source              LogSource `sql:"-"`
	SourceID            sql.NullString
	// Disable these fields: will be populated manually
	// since those have dedicated tables where the info are stored.
	ActivityLevel  []LoggedActivityLevel `sql:"-"`
	HeartRateZones []HeartRateZone       `sql:"-"`
	// Overwrite the custom FitbitDate(Time) fields and overwrite with the SQL-compatible time.Time
	OriginalStartTime time.Time `json:"originalStartTime"`
	StartTime         time.Time `json:"startTime"`
	// Fields that the API for some reason puts on a different type, but have a 1:1 relationship
	// with the activity, and so they have been merged
	ManualValuesSpecified  interface{} `sql:"-"`
	ManualInsertedCalories bool
	ManualInsertedDistance bool
	ManualInsertedSteps    bool
	// Nullable types
	HeartRateLink sql.NullString // e.g. custom activities don't have hr tracking
	TcxLink       sql.NullString
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

type Distance struct {
	types.Distance
	ID int64 `igor:"primary_key"`
}

func (Distance) TableName() string {
	return "distances"
}

type ActivitiesSummary struct {
	types.ActivitiesSummary
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (ActivitiesSummary) TableName() string {
	return "activities_summaries"
}

type DailyActivitySummary struct {
	types.DailyActivitySummary
	ID        int64               `igor:"primary_key"`
	User      pgdb.AuthorizedUser `sql:"-"`
	UserID    int64
	Goal      Goal `sql:"-"`
	GoalID    int64
	Summary   ActivitiesSummary `sql:"-"`
	SummaryID int64
}

func (DailyActivitySummary) TableName() string {
	return "daily_activity_summaries"
}

type LifetimeTimeStep struct {
	types.LifeTimeTimeStep
	ID int64 `igor:"primary_key"`
}

func (LifetimeTimeStep) TableName() string {
	return "life_time_time_steps"
}

type LifetimeActivity struct {
	types.LifeTimeActivities
	ID         int64               `igor:"primary_key"`
	User       pgdb.AuthorizedUser `sql:"-"`
	UserID     int64
	Distance   Distance `sql:"-"`
	DistanceID int64
	Steps      LifetimeTimeStep `sql:"-"`
	StepsID    int64
	Floors     LifetimeTimeStep `sql:"-"`
	FloorsID   int64
}

func (LifetimeActivity) TableName() string {
	return "life_time_activities"
}

type LifetimeStats struct {
	types.LifeTimeStats
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (LifetimeStats) TableName() string {
	return "life_time_stats"
}

type BestStatsSource struct {
	types.BestStatsSource
	ID        int64            `igor:"primary_key"`
	Total     LifetimeActivity `sql:"-"`
	TotalID   int64
	Tracker   LifetimeActivity `sql:"-"`
	TrackerID int64
}

func (BestStatsSource) TableName() string {
	return "best_stats_sources"
}

type LifetimeStatsSource struct {
	types.LifetimeStatsSource
	ID        int64            `igor:"primary_key"`
	Total     LifetimeActivity `sql:"-"`
	TotalID   int64
	Tracker   LifetimeActivity `sql:"-"`
	TrackerID int64
}

func (LifetimeStatsSource) TableName() string {
	return "lifetime_stats_sources"
}

type UserLifeTimeStats struct {
	types.UserLifeTimeStats
	ID         int64           `igor:"primary_key"`
	Best       BestStatsSource `sql:"-"`
	BestID     int64
	Lifetime   LifetimeStatsSource `sql:"-"`
	LifetimeID int64
}

func (UserLifeTimeStats) TableName() string {
	return "user_life_time_stats"
}

type FavoriteActivity struct {
	types.FavoriteActivity
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (FavoriteActivity) TableName() string {
	return "favorite_activities"
}

type MinimalActivity struct {
	types.MinimalActivity
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (MinimalActivity) TableName() string {
	return "minimal_activities"
}

type FrequentActivities struct {
	types.FrequentActivities
	ID                int64           `igor:"primary_key"`
	MinimalActivity   MinimalActivity `sql:"-"`
	MinimalActivityID int64
}

func (FrequentActivities) TableName() string {
	return "frequent_activities"
}

type RecentActivity struct {
	types.RecentActivities
	ID                int64           `igor:"primary_key"`
	MinimalActivity   MinimalActivity `sql:"-"`
	MinimalActivityID int64
}

func (RecentActivity) TableName() string {
	return "recent_activities"
}
