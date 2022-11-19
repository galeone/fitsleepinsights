// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// package datanase initialize the database connection and creates the schema.
package database

import (
	"fmt"
	"os"

	"github.com/galeone/fitbit/types"
	"github.com/galeone/igor"
	_ "github.com/joho/godotenv/autoload"
)

var _db *igor.Database

func init() {
	var err error

	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
	if _db, err = igor.Connect(connectionString); err != nil {
		panic(err.Error())
	}

	//logger := log.New(os.Stdout, "igor: ", log.LUTC)
	//db.Log(logger)

	tx := _db.Begin()

	var authorizedUser types.AuthorizedUser
	if err = tx.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS "%s" (
		user_id TEXT NOT NULL PRIMARY KEY,
		token_type TEXT NOT NULL,
		scope TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		expires_in INTEGER NOT NULL,
		access_token TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE(access_token))`, authorizedUser.TableName())); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	// Create the trigger that sends a notitificaiton every time a new
	// user is added into the authorizedUser table

	/*
		if err = tx.Exec(
			`CREATE FUNCTION IF NOT EXISTS notify_new_user()
			RETURNS TRIGGER
			LANGUAGE plpgsql
			AS $$
			BEGIN
				PERFORM pg_notify('new_users', NEW.user_id);
				RETURN NULL;
			END $$`); err != nil {
			_ = tx.Rollback()
			panic(err.Error())
		}

		if err = tx.Exec(fmt.Sprintf(
			`CREATE TRIGGER IF NOT EXISTS after_insert_user
			AFTER INSERT ON %s
			FOR EACH ROW EXECUTE FUNCTION notify_new_user()`, authorizedUser.TableName())); err != nil {
			_ = tx.Rollback()
			panic(err.Error())
		}
	*/

	var authorizingUser types.AuthorizingUser
	if err = tx.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS "%s" (
		csrftoken TEXT NOT NULL PRIMARY KEY,
		code TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`, authorizingUser.TableName())); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Commit(); err != nil {
		panic(err.Error())
	}
}

// Get returns the valid instance to the *igor.Database
func Get() *igor.Database {
	return _db
}
