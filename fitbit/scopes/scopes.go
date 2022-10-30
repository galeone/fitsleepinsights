// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package scopes defines [all the scopes] supported by the Fitbit API as constants.
// These constanst must be used, instead of using error-prone strings.
//
// [all the scopes]: https://dev.fitbit.com/build/reference/web-api/developer-guide/application-design/#Scopes

package scopes

const (
	// Includes activity data and exercise log related features, such as steps, distance, calories burned, and active minutes.
	Activity = "activity"
	// Includes the maximum or optimum rate at which the user’s heart, lungs, and muscles can effectively use oxygen during exercise.
	CardioFitness = "cardio_fitness"
	// Includes the continuous heart rate data and related analysis.
	Heartrate = "heartrate"
	// 	Includes the GPS and other location data.
	Location = "location"
	// Includes calorie consumption and nutrition related features, such as food/water logging, goals, and plans.
	Nutrition = "nutrition"
	// 	Includes measurements of blood oxygen level.
	OxygenSaturation = "oxygen_saturation"
	// Includes basic user information.
	Profile = "profile"
	// Includes measurements of average breaths per minute at night.
	RespiratoryRate = "respiratory_rate"
	// Includes user account and device settings, such as alarms.
	Settings = "settings"
	// Includes sleep logs and related sleep analysis.
	Sleep = "sleep"
	// Includes friend-related features, such as friend list and leaderboard.
	Social = "social"
	// 	Includes skin and core temperature data.
	Temperature = "temperature"
	// Includes weight and body fat information, such as body mass index, body fat percentage, and goals.
	Weight = "weight"
)

// All returns all the available scopes at once
func All() []string {
	// All the scopes
	return []string{Activity, CardioFitness, Heartrate, Location, Nutrition, OxygenSaturation, Profile, RespiratoryRate, Settings, Sleep, Social, Temperature, Weight}
}
