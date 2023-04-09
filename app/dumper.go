package app

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/galeone/fitbit"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb"
	fitbit_client "github.com/galeone/fitbit/client"
	fitbit_types "github.com/galeone/fitbit/types"
	"github.com/galeone/sleepbit/database"
	"github.com/galeone/sleepbit/database/types"
	"github.com/labstack/echo/v4"
)

func init() {
	_ = _db.Listen(database.NewUsersChannel, func(payload ...string) {
		fmt.Println("notification received")
		if len(payload) != 1 {
			panic(fmt.Sprintf("Expected 1 payload on %s, got %d", database.NewUsersChannel, len(payload)))
		}
		accessToken := payload[0]
		if dumper, err := NewDumper(accessToken); err == nil {
			var all *time.Time = nil
			if err := dumper.Dump(all); err != nil {
				fmt.Printf("dumper.Dump(all): %s", err)
			}
		} else {
			fmt.Println("here: ", err.Error())
		}
	})
}

type dumper struct {
	fb   *fitbit_client.Client
	User *fitbit_pgdb.AuthorizedUser
}

// NewDumper creates a new dumper using the provided access token
func NewDumper(accessToken string) (*dumper, error) {
	authorizer := fitbit.NewAuthorizer(_db, _clientID, _clientSecret, _redirectURL)
	if dbToken, err := _db.AuthorizedUser(accessToken); err != nil {
		return nil, err
	} else {
		if dbToken.UserID == "" {
			return nil, errors.New("invalid token. Please login again")
		}
		authorizer.SetToken(dbToken)
	}

	var fb *fitbit_client.Client
	var err error
	if fb, err = fitbit_client.NewClient(authorizer); err != nil {
		return nil, err
	}

	var abstractUser *fitbit_types.AuthorizedUser
	if abstractUser, err = _db.AuthorizedUser(accessToken); err != nil {
		return nil, err
	}

	var user fitbit_pgdb.AuthorizedUser
	condition := fitbit_pgdb.AuthorizedUser{}
	condition.UserID = abstractUser.UserID
	if err = _db.Model(fitbit_pgdb.AuthorizedUser{}).Where(&condition).Scan(&user); err != nil {
		return nil, err
	}
	return &dumper{fb, &user}, err
}

// TODO: https://pkg.go.dev/github.com/galeone/fitbit@v1.0.0/client ALL

func (d *dumper) acvitityCaloriesTimeseries(startDate, endDate *time.Time) (err error) {
	/*
		var value *fitbit_types.ActivityCaloriesSeries
		if value, err = fp.UserActivityCaloriesTimeseries(startDate, endDate); err != nil {
			return
		}
		for _, t := range value.TimeSeries {
		}
	*/
	return
}

func (d *dumper) userActivityDailyGoal() (err error) {
	var value *fitbit_types.UserGoal
	if value, err = d.fb.UserActivityDailyGoal(); err != nil {
		return err
	}

	now := time.Now()
	insert := types.Goal{Goal: value.Goals}
	insert.UserID = d.User.ID
	insert.StartDate = now
	insert.EndDate = now

	return _db.Create(&insert)
}

func (d *dumper) userActivityLogList(after *time.Time) (err error) {
	var sort string
	var AfterDate fitbit_types.FitbitDateTime
	var BeforeDate fitbit_types.FitbitDateTime
	if after != nil {
		sort = "asc"
		AfterDate.Time = *after
	} else {
		sort = "desc"
		BeforeDate.Time = time.Now()
	}

	pagination := fitbit_types.Pagination{
		AfterDate:  AfterDate,
		BeforeDate: BeforeDate,
		Limit:      100,
		Next:       "",
		Offset:     0,
		Previous:   "",
		Sort:       sort,
	}

	// TODO: ripeti richiesta in loop finché la pagination ritornata ha senso
	// prendi tutti i prima e tutti i dopo (autoconfigurato secondo)
	// la prima richiesta fatta

	var value *fitbit_types.ActivityLogList
	if value, err = d.fb.UserActivityLogList(&pagination); err != nil {
		return
	}

	for _, activity := range value.Activities {
		tx := _db.Begin()
		// Child first
		activeZoneMinutes := types.ActiveZoneMinutes{
			ActiveZoneMinutes: activity.ActiveZoneMinutes,
			ID:                0,
		}

		if err = tx.Create(&activeZoneMinutes); err != nil {
			fmt.Println(err)
			break
		}

		manualValuesSpecified := types.ManualValuesSpecified{
			ManualValuesSpecified: activity.ManualValuesSpecified,
			// Just a number to create the query. Under the hood the ID is a bigserial
			ID: 42,
		}
		if err = tx.Create(&manualValuesSpecified); err != nil {
			fmt.Println(err)
			break
		}

		activityRow := types.ActivityLog{}

		// There are activities without source
		if activity.Source != nil {
			// the API returns a string ID, but it contains a number
			if numericID, err := strconv.ParseInt(activity.Source.ID, 10, 64); err == nil {
				source := types.LogSource{
					LogSource: *activity.Source,
					ID:        numericID,
				}
				// If not present, then create
				if tx.First(&source, numericID) != nil {
					if err = tx.Create(&source); err != nil {
						fmt.Println(err)
						break
					}
				}

				// Handle optional FK
				activityRow.SourceID = sql.NullInt64{
					Int64: source.ID,
					Valid: true,
				}
			} else {
				fmt.Println(err)
				break
			}
		}

		// Primary Key (not serial)
		activityRow.LogID = activity.LogID
		// All the retrieved fields
		activityRow.ActivityLog = activity
		// Non optional FKs: child already created
		activityRow.UserID = d.User.ID
		activityRow.ManualValuesSpecifiedID = manualValuesSpecified.ID
		activityRow.ActiveZoneMinutesID = activeZoneMinutes.ID
		// Overwritten time fields
		activityRow.OriginalStartTime = activity.OriginalStartTime.Time
		activityRow.StartTime = activity.StartTime.Time

		if err = tx.Create(&activityRow); err != nil {
			fmt.Println(err)
			break
		}

		// Once we have the activity stored, we can save the array types returned by the API
		for _, activityLevel := range activity.ActivityLevel {
			// TODO: i random values per PK non van bene, NON bisogna passare un valore
			// provare a ridefinirli come sql.NullInt64 come primo tentativo, e non passarli alla struct
			activityLevelRow := types.LoggedActivityLevel{
				LoggedActivityLevel: activityLevel,
				ActivityLogID:       activityRow.LogID,
				ID:                  1, // random. it's bigserial
			}
			if err = tx.Create(&activityLevelRow); err != nil {
				fmt.Println(err)
				break
			}
		}

		for _, hrZone := range activity.HeartRateZones {
			hrZoneRow := types.HeartRateZone{
				HeartRateZone: hrZone,
				ActivityLogID: sql.NullInt64{
					Int64: activityRow.LogID,
					Valid: true,
				},
				ID: 1, // random, bigserial
			}
			if err = tx.Create(&hrZoneRow); err != nil {
				fmt.Println(err)
				break
			}
		}

		for _, minInHRZone := range activity.ActiveZoneMinutes.MinutesInHeartRateZones {
			minInHRZoneRow := types.MinutesInHeartRateZone{
				MinutesInHeartRateZone: minInHRZone,
				ActiveZoneMinutesID:    activeZoneMinutes.ID,
				ID:                     1, // random
			}
			if err = tx.Create(&minInHRZoneRow); err != nil {
				fmt.Println(err)
				break
			}
		}

		if err = tx.Commit(); err != nil {
			fmt.Println(err)
			break
		}
	}
	return
}

// Dump fetches every data available on the user profile, up to this moment.
// This function is called:
//   - When the user gives the permission to the app (on the INSERT on the table
//     triggered by the database notification)
//   - Periodically by a go routine. In this case, the `after` variable is valid.
func (d *dumper) Dump(after *time.Time) error {
	// There are functions that don't have an "after" period
	// because Fitbit allows to get only the daily data.

	// fmt.Println(d.userActivityDailyGoal())
	fmt.Println(d.userActivityLogList(nil))
	return nil
}

func Dump() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// secure, under middleware
		authorizer := c.Get("fitbit").(*fitbit.Authorizer)
		var userID *string
		if userID, err = authorizer.UserID(); err != nil {
			return err
		}

		var user fitbit_pgdb.AuthorizedUser
		condition := fitbit_pgdb.AuthorizedUser{}
		condition.UserID = *userID
		if err = _db.Model(fitbit_pgdb.AuthorizedUser{}).Where(&condition).Scan(&user); err != nil {
			return err
		}

		if dumper, err := NewDumper(user.AccessToken); err == nil {
			var all *time.Time = nil
			if err := dumper.Dump(all); err != nil {
				fmt.Printf("dumper.Dump(all): %s", err)
			}
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
}
