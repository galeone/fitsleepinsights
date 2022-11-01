// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package types

import "time"

// /body/log/%s/goal.json

type UserWeightGoal struct {
	Goal WeightGoal `json:"goal"`
}

type WeightGoal struct {
	GoalType        string    `json:"goalType"`
	StartDate       time.Time `json:"startDate"`
	StartWeight     int64     `json:"startWeight"`
	Weight          int64     `json:"weight"`
	WeightThreshold float64   `json:"weightThreshold"`
}

// /body/log/%s/goal.json

type UserFatGoal struct {
	Goal FatGoal `json:"goal"`
}

type FatGoal struct {
	Fat int64 `json:"fat"`
}

// /body/log/fat/date/%s.json

type BodyFatLog struct {
	// TODO
	Fat []interface{} `json:"fat"`
}

// /body/log/weight/date/%s.json

type BodyWeightLog struct {
	// TODO
	Weight []interface{} `json:"weight"`
}
