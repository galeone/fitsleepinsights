// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type Goal struct {
	types.Goal
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (Goal) TableName() string {
	return "goals"
}

type ManualValuesSpecified struct {
	types.ManualValuesSpecified
	ID int64 `igor:"primary_key"`
}

func (ManualValuesSpecified) TableName() string {
	return "manual_values_specified"
}

type HeartRateZone struct {
	types.HeartRateZone
	ID int64 `igor:"primary_key"`
}

func (HeartRateZone) TableName() string {
	return "heart_rate_zones"
}

type LogSource struct {
	types.LogSource
	ID int64 `igor:"primary_key"`
}

func (LogSource) TableName() string {
	return "log_sources"
}

type MinutesInHeartRateZone struct {
	types.MinutesInHeartRateZone
	ID int64 `igor:"primary_key"`
}

func (MinutesInHeartRateZone) TableName() string {
	return "minutes_in_heart_rate_zone"
}

type LoggedActivityLevel struct {
	types.LoggedActivityLevel
	ID int64 `igor:"primary_key"`
}

func (LoggedActivityLevel) TableName() string {
	return "logged_activity_levels"
}

type ActiveZoneMinutes struct {
	types.ActiveZoneMinutes
	ID int64 `igor:"primary_key"`
}

func (ActiveZoneMinutes) TableName() string {
	return "active_zone_minutes"
}

type ActivityLog struct {
	types.ActivityLog
	ID                    int64                 `igor:"primary_key"`
	User                  pgdb.AuthorizedUser   `sql:"-"`
	ActiveZoneMinutes     ActiveZoneMinutes     `sql:"-"`
	ManualValuesSpecified ManualValuesSpecified `sql:"-"`
	Source                LogSource             `sql:"-"`
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
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (ActivitiesSummary) TableName() string {
	return "activities_summaries"
}

type DailyActivitySummary struct {
	types.DailyActivitySummary
	ID      int64               `igor:"primary_key"`
	User    pgdb.AuthorizedUser `sql:"-"`
	Goal    Goal                `sql:"-"`
	Summary ActivitiesSummary   `sql:"-"`
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
	ID       int64               `igor:"primary_key"`
	User     pgdb.AuthorizedUser `sql:"-"`
	Distance Distance            `sql:"-"`
	Steps    LifetimeTimeStep    `sql:"-"`
	Floors   LifetimeTimeStep    `sql:"-"`
}

func (LifetimeActivity) TableName() string {
	return "life_time_activities"
}

type LifetimeStats struct {
	types.LifeTimeStats
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (LifetimeStats) TableName() string {
	return "life_time_stats"
}

type BestStatsSource struct {
	types.BestStatsSource
	ID      int64            `igor:"primary_key"`
	Total   LifetimeActivity `sql:"-"`
	Tracker LifetimeActivity `sql:"-"`
}

func (BestStatsSource) TableName() string {
	return "best_stats_sources"
}

type LifetimeStatsSource struct {
	types.LifetimeStatsSource
	ID      int64            `igor:"primary_key"`
	Total   LifetimeActivity `sql:"-"`
	Tracker LifetimeActivity `sql:"-"`
}

func (LifetimeStatsSource) TableName() string {
	return "lifetime_stats_sources"
}

type UserLifeTimeStats struct {
	types.UserLifeTimeStats
	ID       int64               `igor:"primary_key"`
	Best     BestStatsSource     `sql:"-"`
	Lifetime LifetimeStatsSource `sql:"-"`
}

func (UserLifeTimeStats) TableName() string {
	return "user_life_time_stats"
}

type FavoriteActivity struct {
	types.FavoriteActivity
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (FavoriteActivity) TableName() string {
	return "favorite_activities"
}

type MinimalActivity struct {
	types.MinimalActivity
	ID   int64               `igor:"primary_key"`
	User pgdb.AuthorizedUser `sql:"-"`
}

func (MinimalActivity) TableName() string {
	return "minimal_activities"
}

type FrequentActivities struct {
	types.FrequentActivities
	ID              int64           `igor:"primary_key"`
	MinimalActivity MinimalActivity `sql:"-"`
}

func (FrequentActivities) TableName() string {
	return "frequent_activities"
}

type RecentActivity struct {
	types.RecentActivities
	ID              int64           `igor:"primary_key"`
	MinimalActivity MinimalActivity `sql:"-"`
}

func (RecentActivity) TableName() string {
	return "recent_activities"
}
