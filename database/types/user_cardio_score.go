// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import (
	"time"

	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
)

type BreathingRate struct {
	types.BreathingRate
	ID     int64               `igor:"primary_key"`
	User   pgdb.AuthorizedUser `sql:"-"`
	UserID int64
}

func (BreathingRate) TableName() string {
	return "breathing_rate"
}

type CardioFitnessScore struct {
	types.CardioScoreTimePoint
	ID       int64               `igor:"primary_key"`
	User     pgdb.AuthorizedUser `sql:"-"`
	UserID   int64
	DateTime types.FitbitDate `sql:"-"` // it's a date
	Date     time.Time
	// The value in the API is a string nested in a struct.
	// The values are in the format xx-yy, with xx min vo2max
	// during the date, yy the highest
	Value            string `sql:"-"`
	Vo2MaxLowerBound float64
	Vo2MaxUpperBound float64
}

func (CardioFitnessScore) TableName() string {
	return "cardio_fitness_score"
}
