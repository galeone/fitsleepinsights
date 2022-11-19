package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/galeone/fitbit"
	"github.com/galeone/fitbit/types"
	"github.com/galeone/sleepbit/database"
)

func init() {
	_ = _db.Listen(database.NewUsersChannel, func(payload ...string) {
		if len(payload) != 1 {
			panic(fmt.Sprintf("Expected 1 payload on %s, got %d", database.NewUsersChannel, len(payload)))
		}
		accessToken := payload[0]
		dumpAll(accessToken, nil)
	})
}

// dumpAll fetches every data available on the user profile, up to this moment.
// This function is called:
//   - When the user gives the permission to the app (on the INSERT on the table
//     triggered by the database notification)
//   - Periodically by a go routine. In this case, the `after` varialbe is valid.
func dumpAll(accessToken string, after *time.Time) (err error) {
	authorizer := fitbit.NewAuthorizer(_db, _clientID, _clientSecret, _redirectURL)
	var dbToken *types.AuthorizedUser
	if dbToken, err = _db.AuthorizedUser(accessToken); err != nil {
		return err
	}
	if dbToken.UserID == "" {
		return errors.New("Invalid token. Please login again")
	}
	authorizer.SetToken(dbToken)

	//fb, err := client.NewClient(authorizer)
	return
}
