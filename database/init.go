// package datanase initialize the database connection and creates the schema.
package database

import (
	"fmt"
	"os"

	"github.com/galeone/igor"
	"github.com/galeone/sleepbit/fitbit/types"
	_ "github.com/joho/godotenv/autoload"
)

var db *igor.Database

func init() {
	var err error

	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
	if db, err = igor.Connect(connectionString); err != nil {
		panic(err.Error())
	}

	//logger := log.New(os.Stdout, "igor: ", log.LUTC)
	//db.Log(logger)

	tx := db.Begin()

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
	return db
}
