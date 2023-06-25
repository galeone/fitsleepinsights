// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// package datanase initialize the database connection and creates the schema.
package database

import (
	_ "embed"

	"github.com/galeone/igor"
	_ "github.com/joho/godotenv/autoload"
)

var (
	//go:embed schema/user.sql
	user string

	//go:embed schema/activities.sql
	activities string

	//go:embed schema/user_activity.sql
	user_activity string

	//go:embed schema/user_activity_timeseries.sql
	user_activity_timeseries string

	//go:embed schema/user_body.sql
	user_body string

	//go:embed schema/user_body_timeseries.sql
	user_body_timeseries string

	//go:embed schema/user_breathing_rate.sql
	user_breathing_rate string

	//go:embed schema/user_cardio_score.sql
	user_cardio_score string

	//go:embed schema/user_hr_timeseries.sql
	user_hr_timeseries string

	//go:embed schema/user_hr_variability.sql
	user_hr_variability string

	//go:embed schema/user_intraday.sql
	user_intraday string

	//go:embed schema/user_oxygen_saturation.sql
	user_oxygen_saturation string

	//go:embed schema/user_sleep.sql
	user_sleep string

	//go:embed schema/user_temperature.sql
	user_temperature string

	//go:embed schema/alter.sql
	alter string
)

func init() {
	// Database instance only local to this init function, used to initialize the database and the application startup.
	// The global database instance is initialized in app/globals.go.
	var db *igor.Database
	var err error

	if db, err = igor.Connect(_connectionString); err != nil {
		panic(err.Error())
	}
	defer db.DB().Close()

	//logger := log.New(os.Stdout, "igor: ", log.LUTC)
	//db.Log(logger)

	tx := db.Begin()

	// There's only one dependency between sql files: user_hr_timeseries.sql
	// uses a table defined in user_activity.sql.
	// All the other tables just depend on the users tables.

	if err = tx.Exec(user); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(activities); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_activity); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_activity_timeseries); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_body); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_body_timeseries); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_breathing_rate); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_cardio_score); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_hr_timeseries); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_hr_variability); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_intraday); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_oxygen_saturation); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_sleep); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(user_temperature); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Exec(alter); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Commit(); err != nil {
		panic(err.Error())
	}
}
