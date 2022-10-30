package database

import (
	"fmt"
	"os"

	"github.com/galeone/igor"
	"github.com/galeone/sleepbit/fitbit/types"

	_ "github.com/joho/godotenv/autoload"
)

// Storage implements the fitbit.Storage interface
type Storage struct {
	db *igor.Database
}

func NewStorage() *Storage {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
	var err error
	var db *igor.Database
	if db, err = igor.Connect(connectionString); err != nil {
		panic(err.Error())
	}

	return &Storage{
		db,
	}
}

func (s *Storage) InsertAuhorizingUser(authorizing *types.AuthorizingUser) error {
	return db.Create(authorizing)
}

func (s *Storage) UpsertAuthorizedUser(user *types.AuthorizedUser) error {
	var exists types.AuthorizedUser
	var err error
	if err = db.First(&exists, user.UserID); err != nil {
		// First time we see this user
		err = db.Create(user)
	} else {
		err = db.Updates(user)
	}
	return err
}

func (s *Storage) AuthorizedUser(accessToken string) (*types.AuthorizedUser, error) {
	var dbToken types.AuthorizedUser
	var err error
	if err = db.Model(types.AuthorizedUser{}).Where(&types.AuthorizedUser{AccessToken: accessToken}).Scan(&dbToken); err != nil {
		return nil, err
	}
	return &dbToken, nil
}

func (s *Storage) AuthorizingUser(id string) (*types.AuthorizingUser, error) {
	var authorizing types.AuthorizingUser
	var err error
	if err = db.First(&authorizing, id); err != nil {
		return nil, err
	}
	return &authorizing, nil
}
