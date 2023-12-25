// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"strconv"
	"time"

	pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/galeone/fitbit/v2/types"
)

type SleepStageDetail struct {
	types.SleepStageDetail
	ID         int64 `igor:"primary_key"`
	SleepLogID int64
	SleepStage string
}

func (SleepStageDetail) TableName() string {
	return "sleep_stage_details"
}

type SleepData struct {
	types.SleepData
	ID         int64 `igor:"primary_key"`
	SleepLogID int64
	DateTime   time.Time
}

func (SleepData) TableName() string {
	return "sleep_data"
}

type SleepLog struct {
	types.SleepLog
	LogID  int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
	// Levels has a 1:1 relationship with SleepLog. So instead of using a SleepLevel type (changed to int64 since ignored)
	// we can just remove this useless connection point and use the LogID as FK for all the other data.
	Levels      types.SleepLevel `sql:"-"`
	DateOfSleep time.Time
	EndTime     time.Time
	StartTime   time.Time
}

func (SleepLog) Headers() []string {
	return []string{
		"SleepDuration",
		"SleepEfficiency",
		"MinutesAfterWakeup",
		"MinutesAsleep",
		"MinutesAwake",
		"MinutesToFallAsleep",
		"TimeInBed",

		"LightSleepMinutes",
		"LightSleepCount",
		"DeepSleepMinutes",
		"DeepSleepCount",
		"RemSleepMinutes",
		"RemSleepCount",
		"WakeSleepMinutes",
		"WakeSleepCount",
	}
}

func (f *SleepLog) Values() []string {
	return []string{
		strconv.FormatInt(f.Duration, 10),
		strconv.FormatInt(f.Efficiency, 10),
		strconv.FormatInt(f.MinutesAfterWakeup, 10),
		strconv.FormatInt(f.MinutesAsleep, 10),
		strconv.FormatInt(f.MinutesAwake, 10),
		strconv.FormatInt(f.MinutesToFallAsleep, 10),
		strconv.FormatInt(f.TimeInBed, 10),

		strconv.FormatInt(f.Levels.Summary.Light.Minutes, 10),
		strconv.FormatInt(f.Levels.Summary.Light.Count, 10),
		strconv.FormatInt(f.Levels.Summary.Deep.Minutes, 10),
		strconv.FormatInt(f.Levels.Summary.Deep.Count, 10),
		strconv.FormatInt(f.Levels.Summary.Rem.Minutes, 10),
		strconv.FormatInt(f.Levels.Summary.Rem.Count, 10),
		strconv.FormatInt(f.Levels.Summary.Wake.Minutes, 10),
		strconv.FormatInt(f.Levels.Summary.Wake.Count, 10),
	}
}

func (SleepLog) TableName() string {
	return "sleep_logs"
}
