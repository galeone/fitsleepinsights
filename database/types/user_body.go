// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type WeightGoal struct {
	types.WeightGoal
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (WeightGoal) TableName() string {
	return "weight_goals"
}

type FatGoal struct {
	types.WeightGoal
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (FatGoal) TableName() string {
	return "fat_goals"
}

type FatLog struct {
	types.UserFatLog
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (FatLog) TableName() string {
	return "fat_logs"
}

type WeightLog struct {
	types.UserWeightLog
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (WeightLog) TableName() string {
	return "weight_logs"
}
