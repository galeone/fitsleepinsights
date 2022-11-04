// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package database

import (
	"fmt"
	"os"

	"github.com/galeone/igor"
	"github.com/galeone/sleepbit/fitbit/types"

	_ "github.com/joho/godotenv/autoload"
)

// PQDB implements the fitbit.Storage interface
type PQDB struct {
	*igor.Database
}

// NewPQDB creates a new connection to a PostgreSQL server
// using the following environement variables:
// - DB_USER
// - DB_PASS
// - DB_NAME
// Loaded from a `.env` file, if present.
//
// It implements the `fitbit.Storage` interface.
func NewPQDB() *PQDB {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
	var err error
	var db *igor.Database
	if db, err = igor.Connect(connectionString); err != nil {
		panic(err.Error())
	}

	return &PQDB{
		db,
	}
}

func (s *PQDB) InsertAuhorizingUser(authorizing *types.AuthorizingUser) error {
	return s.Create(authorizing)
}

func (s *PQDB) UpsertAuthorizedUser(user *types.AuthorizedUser) error {
	var exists types.AuthorizedUser
	var err error
	if err = s.First(&exists, user.UserID); err != nil {
		// First time we see this user
		err = s.Create(user)
	} else {
		err = s.Updates(user)
	}
	return err
}

func (s *PQDB) AuthorizedUser(accessToken string) (*types.AuthorizedUser, error) {
	var dbToken types.AuthorizedUser
	var err error
	if err = s.Model(types.AuthorizedUser{}).Where(&types.AuthorizedUser{AccessToken: accessToken}).Scan(&dbToken); err != nil {
		return nil, err
	}
	return &dbToken, nil
}

func (s *PQDB) AuthorizingUser(id string) (*types.AuthorizingUser, error) {
	var authorizing types.AuthorizingUser
	var err error
	if err = s.First(&authorizing, id); err != nil {
		return nil, err
	}
	return &authorizing, nil
}
