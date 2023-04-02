// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type SleepStageDetail struct {
	types.SleepStageDetail
	ID int64 `igor:"primary_key"`
}

func (SleepStageDetail) TableName() string {
	return "sleep_stage_details"
}

type SleepData struct {
	types.SleepData
	ID int64 `igor:"primary_key"`
}

func (SleepData) TableName() string {
	return "sleep_data"
}

type SleepLevel struct {
	types.SleepLevel
	ID          int64     `igor:"primary_key"`
	SleepData   SleepData `sql:"-"`
	SleepDataID int64
	ShortData   SleepData `sql:"-"`
	ShortDataID int64
	Summary     SleepStageDetail `sql:"-"`
	SummaryID   int64
}

func (SleepLevel) TableName() string {
	return "sleep_levels"
}

type SleepLog struct {
	types.SleepLog
	ID       int64               `igor:"primary_key"`
	User     pgdb.AuthorizedUser `sql:"-"`
	UserID   int64
	Levels   SleepLevel `sql:"-"`
	LevelsID int64
}

func (SleepLog) TableName() string {
	return "sleep_logs"
}

type SleepStagesSummary struct {
	types.SleepStagesSummary
	ID int64 `igor:"primary_key"`
}

func (SleepStagesSummary) TableName() string {
	return "sleep_stages_summary"
}

type SleepSummary struct {
	types.SleepSummary
	ID       int64               `igor:"primary_key"`
	User     pgdb.AuthorizedUser `sql:"-"`
	UserID   int64
	Stages   SleepStagesSummary `sql:"-"`
	StagesID int64
}

func (SleepSummary) TableName() string {
	return "sleep_summary"
}
