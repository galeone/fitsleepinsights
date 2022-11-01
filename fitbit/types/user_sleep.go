// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

// /sleep/date/%s.json
// /sleep/date/%s/%s.json
// /sleep/list.json?afterDate=2022-10-31&sort=asc&offset=0&limit=2
type SleepLogs struct {
	Sleep      []SleepLog   `json:"sleep"`
	Summary    SleepSummary `json:"summary,omitempty"`
	Pagination Pagination   `json:"pagination"`
}

type SleepStageDetail struct {
	Count               int64 `json:"count"`
	Minutes             int64 `json:"minutes"`
	ThirtyDayAvgMinutes int64 `json:"thirtyDayAvgMinutes"`
}

type SleepLevel struct {
	Data      []SleepData `json:"data"`
	ShortData []SleepData `json:"shortData"`
	Summary   SleepStages `json:"summary"`
}

type SleepLog struct {
	DateOfSleep         FitbitDate `json:"dateOfSleep"`
	Duration            int64      `json:"duration"`
	Efficiency          int64      `json:"efficiency"`
	EndTime             string     `json:"endTime"`
	InfoCode            int64      `json:"infoCode"`
	IsMainSleep         bool       `json:"isMainSleep"`
	Levels              SleepLevel `json:"levels"`
	LogID               int64      `json:"logId"`
	LogType             string     `json:"logType"`
	MinutesAfterWakeup  int64      `json:"minutesAfterWakeup"`
	MinutesAsleep       int64      `json:"minutesAsleep"`
	MinutesAwake        int64      `json:"minutesAwake"`
	MinutesToFallAsleep int64      `json:"minutesToFallAsleep"`
	StartTime           string     `json:"startTime"`
	TimeInBed           int64      `json:"timeInBed"`
	Type                string     `json:"type"`
}

type SleepData struct {
	DateTime FitbitDateTime `json:"dateTime"`
	Level    string         `json:"level"`
	Seconds  int64          `json:"seconds"`
}

type SleepStages struct {
	Deep  SleepStageDetail `json:"deep"`
	Light SleepStageDetail `json:"light"`
	Rem   SleepStageDetail `json:"rem"`
	Wake  SleepStageDetail `json:"wake"`
}

type SleepStagesSummary struct {
	Deep  int64 `json:"deep"`
	Light int64 `json:"light"`
	Rem   int64 `json:"rem"`
	Wake  int64 `json:"wake"`
}

type SleepSummary struct {
	Stages             SleepStagesSummary `json:"stages"`
	TotalMinutesAsleep int64              `json:"totalMinutesAsleep"`
	TotalSleepRecords  int64              `json:"totalSleepRecords"`
	TotalTimeInBed     int64              `json:"totalTimeInBed"`
}

// /sleep/goal.json

type SleepGoalReport struct {
	Consistency SleepConsistency `json:"consistency"`
	Goal        SleepGoal        `json:"goal"`
}

type SleepConsistency struct {
	AwakeRestlessPercentage float64 `json:"awakeRestlessPercentage"`
	FlowID                  int64   `json:"flowId"`
	RecommendedSleepGoal    int64   `json:"recommendedSleepGoal"`
	TypicalDuration         int64   `json:"typicalDuration"`
	TypicalWakeupTime       string  `json:"typicalWakeupTime"`
}

type SleepGoal struct {
	Bedtime     string         `json:"bedtime"`
	MinDuration int64          `json:"minDuration"`
	UpdatedOn   FitbitDateTime `json:"updatedOn"`
	WakeupTime  string         `json:"wakeupTime"`
}
