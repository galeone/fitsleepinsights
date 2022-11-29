package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/galeone/fitbit"
	"github.com/galeone/fitbit/client"
	"github.com/galeone/sleepbit/database"
)

func init() {
	_ = _db.Listen(database.NewUsersChannel, func(payload ...string) {
		fmt.Println("notification received")
		if len(payload) != 1 {
			panic(fmt.Sprintf("Expected 1 payload on %s, got %d", database.NewUsersChannel, len(payload)))
		}
		accessToken := payload[0]
		if err := dumpAll(accessToken, nil); err != nil {
			fmt.Println(err.Error())
		}
	})
}

// dumpAll fetches every data available on the user profile, up to this moment.
// This function is called:
//   - When the user gives the permission to the app (on the INSERT on the table
//     triggered by the database notification)
//   - Periodically by a go routine. In this case, the `after` varialbe is valid.
func dumpAll(accessToken string, after *time.Time) error {
	authorizer := fitbit.NewAuthorizer(_db, _clientID, _clientSecret, _redirectURL)
	if dbToken, err := _db.AuthorizedUser(accessToken); err != nil {
		return err
	} else {
		if dbToken.UserID == "" {
			return errors.New("Invalid token. Please login again")
		}
		authorizer.SetToken(dbToken)
	}

	var fb *client.Client
	var err error
	if fb, err = client.NewClient(authorizer); err != nil {
		return err
	}

	if res, err := fb.UserSleepGoalReport(); err != nil {
		return err
	} else {
		fmt.Println(res)
	}
	return nil
}
