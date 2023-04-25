package app

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/galeone/fitbit"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb"
	fitbit_client "github.com/galeone/fitbit/client"
	fitbit_types "github.com/galeone/fitbit/types"
	"github.com/galeone/sleepbit/database"
	"github.com/galeone/sleepbit/database/types"
	"github.com/galeone/tcx"
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
			dumpTCX := false
			if err := dumper.Dump(all, dumpTCX); err != nil {
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

func (d *dumper) userActivityCaloriesTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.ActivityCaloriesSeries
	if value, err = d.fb.UserActivityCaloriesTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.ActivityCaloriesSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.ActivityCaloriesSeries{}
		// No error = found
		if err = tx.Model(types.ActivityCaloriesSeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

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

	if err = _db.Model(types.Goal{}).Where(&insert).Scan(&insert); err != nil {
		return _db.Create(&insert)
	}
	return
}

// The parameter dumpTCX is required because every TCX dump request is an API call.
// Fitbit limits 250 API calls user/hour. Thus, on the first dump we have
// 1 single API call for the activity list. But 100 API calls for the TCX.
func (d *dumper) userActivityLogList(after *time.Time, dumpTCX bool) (err error) {
	var sort string
	var AfterDate fitbit_types.FitbitDateTime
	var BeforeDate fitbit_types.FitbitDateTime
	var limit int64 = 100
	var offset int64 = 0

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
		Limit:      limit,
		Next:       "",
		Offset:     offset,
		Previous:   "",
		Sort:       sort,
	}

	for {
		var value *fitbit_types.ActivityLogList
		if value, err = d.fb.UserActivityLogList(&pagination); err != nil {
			if strings.Contains(err.Error(), "504") {
				// NOTE: Fitbit API is a shit.
				// It only allows us to request the latest 100 activities.
				// Thus, if we are asking for other activities it doesn't return a meaningful error
				// But it returns a Gateway Timeout (WTF).
				break
			}
			return
		}
		if len(value.Activities) == 0 {
			break
		}

		for _, activity := range value.Activities {
			activityRow := types.ActivityLog{}
			if err = _db.First(&activityRow, activity.LogID); err == nil {
				fmt.Println("skipping activity ", activity.LogID, ": already present")
				continue
			}

			tx := _db.Begin()
			// There are activities without active zone minutes
			if activity.ActiveZoneMinutes.TotalMinutes > 0 {
				activeZoneMinutes := types.ActiveZoneMinutes{
					ActiveZoneMinutes: activity.ActiveZoneMinutes,
				}

				if err = tx.Create(&activeZoneMinutes); err != nil {
					fmt.Println(err)
					break
				}
				for _, minInHRZone := range activity.ActiveZoneMinutes.MinutesInHeartRateZones {
					minInHRZoneRow := types.MinutesInHeartRateZone{
						MinutesInHeartRateZone: minInHRZone,
						ActiveZoneMinutesID:    activeZoneMinutes.ID,
					}
					if err = tx.Create(&minInHRZoneRow); err != nil {
						fmt.Println(err)
						break
					}
				}
				activityRow.ActiveZoneMinutesID = sql.NullInt64{
					Int64: activeZoneMinutes.ID,
					Valid: true,
				}
			}

			// There are activities without source
			if activity.Source != nil {
				var source types.LogSource // NOTE: First requires the dest field to be zero to work correctly
				// If not present, then create
				if err = tx.First(&source, activity.Source.ID); err != nil {
					source = types.LogSource{
						LogSource: *activity.Source,
						ID:        activity.Source.ID,
					}
					if err = tx.Create(&source); err != nil {
						fmt.Println(err)
						break
					}
				}

				// Handle optional FK
				activityRow.SourceID = sql.NullString{
					String: source.ID,
					Valid:  true,
				}

			}

			// Primary Key (not serial)
			activityRow.LogID = activity.LogID
			// All the retrieved fields
			activityRow.ActivityLog = activity
			// Non optional FKs: child already created
			activityRow.UserID = d.User.ID
			// Overwritten time fields
			activityRow.OriginalStartTime = activity.OriginalStartTime.Time
			activityRow.StartTime = activity.StartTime.Time
			// Fields that the API for some reason puts on a different type, but have a 1:1 relationship
			// with the activity, and so they have been merged
			activityRow.ManualInsertedCalories = activity.ManualValuesSpecified.Calories
			activityRow.ManualInsertedSteps = activity.ManualValuesSpecified.Steps
			activityRow.ManualInsertedDistance = activity.ManualValuesSpecified.Distance

			if dumpTCX {
				var xml *tcx.TCXDB
				if xml, err = d.fb.UserActivityTCX(activity.LogID); err == nil {
					if textBytes, err := tcx.ToBytes(*xml); err != nil {
						fmt.Println(err)
					} else {
						activityRow.Tcx = sql.NullString{
							String: string(textBytes),
							Valid:  true,
						}
					}
				} else {
					fmt.Println(err)
					// Do not break: who cares about failing fetch of TCX data (Fitbit has several problems with that)
				}
			}

			if err = tx.Create(&activityRow); err != nil {
				fmt.Println(err)
				break
			}

			// Once we have the activity stored, we can save the array types returned by the API
			for _, activityLevel := range activity.ActivityLevel {
				activityLevelRow := types.LoggedActivityLevel{
					LoggedActivityLevel: activityLevel,
					ActivityLogID:       activityRow.LogID,
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
				}
				if err = tx.Create(&hrZoneRow); err != nil {
					fmt.Println(err)
					break
				}
			}

			if err = tx.Commit(); err != nil {
				fmt.Println(err)
				break
			}
		}

		if nextURL, err := url.Parse(value.Pagination.Next); err != nil {
			fmt.Println(err)
			break
		} else {
			next := nextURL.Query()
			offset, _ = strconv.ParseInt(next.Get("offset"), 10, 64)
			if afterDate := next.Get("afterDate"); afterDate != "" {
				AfterDate.Time, _ = time.Parse(fitbit_types.DateTimeLayout, afterDate)
			}
			if beforeDate := next.Get("beforeDate"); beforeDate != "" {
				BeforeDate.Time, _ = time.Parse(fitbit_types.DateTimeLayout, beforeDate)
			}
			pagination = fitbit_types.Pagination{
				AfterDate:  AfterDate,
				BeforeDate: BeforeDate,
				Limit:      limit,
				Next:       "",
				Offset:     offset,
				Previous:   "",
				Sort:       sort,
			}
		}

	}
	return
}

func (d *dumper) userActivityWeeklyGoal() (err error) {
	var value *fitbit_types.UserGoal
	if value, err = d.fb.UserActivityWeeklyGoal(); err != nil {
		return err
	}

	now := time.Now()
	// StartDate = monday of the current week
	// EndDate = sunday of the same week
	insert := types.Goal{Goal: value.Goals}
	insert.UserID = d.User.ID
	insert.StartDate = now.Add(-time.Duration((int(now.Weekday()) - 1) * 24 * int(time.Hour)))
	insert.EndDate = insert.StartDate.Add(time.Duration(24 * 6 * time.Hour))

	if err = _db.Model(types.Goal{}).Where(&insert).Scan(&insert); err != nil {
		return _db.Create(&insert)
	}

	return
}

func (d *dumper) userBMITimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.BMISeries
	if value, err = d.fb.UserBMITimeSeries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.BMISeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.BMISeries{}).Where(&timestep).Scan(&timestep); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userBodyFatTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.BodyFatSeries
	if value, err = d.fb.UserBodyFatTimeSeries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.BodyFatSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.BodyFatSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userBodyWeightTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.BodyWeightSeries
	if value, err = d.fb.UserBodyWeightTimeSeries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.BodyWeightSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.BodyWeightSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userCaloriesBMRTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.CaloriesBMRSeries
	if value, err = d.fb.UserCaloriesBMRTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.CaloriesBMRSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.CaloriesBMRSeries{}
		// No error = found
		if err = tx.Model(types.CaloriesBMRSeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userCaloriesTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.CaloriesSeries
	if value, err = d.fb.UserCaloriesTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.CaloriesSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.CaloriesSeries{}
		// No error = found
		if err = tx.Model(types.CaloriesSeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userDistanceTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.DistanceSeries
	if value, err = d.fb.UserDistanceTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.DistanceSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.DistanceSeries{}
		// No error = found
		if err = tx.Model(types.DistanceSeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userFloorsTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.FloorsSeries
	if value, err = d.fb.UserFloorsTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.FloorsSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.FloorsSeries{}
		// No error = found
		if err = tx.Model(types.FloorsSeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userHeartRateTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.HeartRateSeries
	if value, err = d.fb.UserHeartRateTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, hrActivity := range value.ActivitiesHeart {
		hrActivityInsert := types.HeartRateActivities{}
		hrActivityInsert.HeartRateActivities = hrActivity
		hrActivityInsert.UserID = d.User.ID
		hrActivityInsert.Date = hrActivity.DateTime.Time

		// No error = found
		if err = tx.Model(types.HeartRateActivities{}).Where(&hrActivityInsert).Scan(&hrActivityInsert); err == nil {
			fmt.Println("Skipping ", hrActivityInsert)
			continue
		}

		if err = tx.Create(&hrActivityInsert); err != nil {
			fmt.Println(err)
			break
		}

		insertHrZone := func(hrZone *fitbit_types.HeartRateZone, zoneType string) error {
			value := types.HeartRateZone{
				HeartRateZone: *hrZone,
				Type:          zoneType,
				HeartRateActivityID: sql.NullInt64{
					Valid: true,
					Int64: hrActivityInsert.ID,
				},
			}
			return tx.Create(&value)
		}
		for _, hrZone := range hrActivity.Value.HeartRateZones {
			if err := insertHrZone(&hrZone, "DEFAULT"); err != nil {
				fmt.Println(err)
				break
			}
		}
		for _, customHrZone := range hrActivity.Value.CustomHeartRateZones {
			if err := insertHrZone(&customHrZone, "CUSTOM"); err != nil {
				fmt.Println(err)
				break
			}
		}
	}

	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userMinutesFairlyActiveTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.MinutesFairlyActiveSeries
	if value, err = d.fb.UserMinutesFairlyActiveTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.MinutesFairlyActiveSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.MinutesFairlyActiveSeries{}
		// No error = found
		if err = tx.Model(types.MinutesFairlyActiveSeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userMinutesLightlyActiveTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.MinutesLightlyActiveSeries
	if value, err = d.fb.UserMinutesLightlyActiveTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.MinutesLightlyActiveSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.MinutesLightlyActiveSeries{}
		// No error = found
		if err = tx.Model(types.MinutesLightlyActiveSeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userMinutesSedentaryTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.MinutesSedentarySeries
	if value, err = d.fb.UserMinutesSedentaryTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.MinutesSedentarySeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.MinutesSedentarySeries{}
		// No error = found
		if err = tx.Model(types.MinutesSedentarySeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userMinutesVeryActiveTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.MinutesVeryActiveSeries
	if value, err = d.fb.UserMinutesVeryActiveTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.MinutesVeryActiveSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.MinutesVeryActiveSeries{}
		// No error = found
		if err = tx.Model(types.MinutesVeryActiveSeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userStepsTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.StepsSeries
	if value, err = d.fb.UserStepsTimeseries(startDate, endDate); err != nil {
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.StepsSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			fmt.Println(err)
			break
		}

		dest := types.StepsSeries{}
		// No error = found
		if err = tx.Model(types.StepsSeries{}).Where(&timestep).Scan(&dest); err == nil {
			fmt.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			fmt.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return
}

func (d *dumper) userSleepLogList(startDate, endDate *time.Time) (err error) {
	var sleepLogs *fitbit_types.SleepLogs
	if sleepLogs, err = d.fb.UserSleepLog(startDate, endDate); err != nil {
		return err
	}

	tx := _db.Begin()
	for _, sleepLog := range sleepLogs.Sleep {
		insert := types.SleepLog{
			SleepLog:    sleepLog,
			LogID:       sleepLog.LogID,
			UserID:      d.User.ID,
			DateOfSleep: sleepLog.DateOfSleep.Time,
			StartTime:   sleepLog.StartTime.Time,
			EndTime:     sleepLog.EndTime.Time,
		}

		// No error = found
		if err = tx.Model(types.SleepLog{}).Where(&insert).Scan(&insert); err == nil {
			fmt.Println("Skipping ", insert)
			continue
		}
		if err = tx.Create(&insert); err != nil {
			fmt.Println(err)
			break
		}

		sleepStage := func(stage *fitbit_types.SleepStageDetail, name string) error {
			insertStage := types.SleepStageDetail{
				SleepStageDetail: *stage,
				SleepLogID:       insert.LogID,
			}
			if err = tx.Create(&insertStage); err != nil {
				return err
			}
			return nil
		}

		if err = sleepStage(&sleepLog.Levels.Summary.Deep, "DEEP"); err != nil {
			fmt.Println(err)
			break
		}
		if err = sleepStage(&sleepLog.Levels.Summary.Light, "LIGHT"); err != nil {
			fmt.Println(err)
			break
		}
		if err = sleepStage(&sleepLog.Levels.Summary.Rem, "REM"); err != nil {
			fmt.Println(err)
			break
		}
		if err = sleepStage(&sleepLog.Levels.Summary.Wake, "WAKE"); err != nil {
			fmt.Println(err)
			break
		}

		for _, sleepData := range sleepLog.Levels.Data {
			levelDataInsert := types.SleepData{
				SleepData:  sleepData,
				SleepLogID: sleepLog.LogID,
				DateTime:   sleepData.DateTime.Time,
			}
			if err = tx.Create(&levelDataInsert); err != nil {
				return err
			}
			return nil
		}
	}
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
	}
	return
}

// Dump fetches every data available on the user profile, up to this moment.
// This function is called:
//   - When the user gives the permission to the app (on the INSERT on the table
//     triggered by the database notification)
//   - Periodically by a go routine. In this case, the `after` variable is valid.
func (d *dumper) Dump(after *time.Time, dumpTCX bool) error {
	// Date super-old in the past (but not too old to make the server return an error)
	//startDate, _ := time.Parse(fitbit_types.DateLayout, "2009-01-01")
	endDate := time.Now()
	//startDate := endDate.Add(-time.Duration(24*60) * time.Hour)
	startDate := endDate.Add(-time.Duration(24*1) * time.Hour)

	// There are functions that don't have an "after" period
	// because Fitbit allows to get only the daily data.

	fmt.Println(d.userActivityDailyGoal())
	fmt.Println(d.userActivityWeeklyGoal())

	// NOTE: this is not a dump ALL activities. But only the latest 100 activities
	// because hte Fitbit API limit (for no reason) this endpoint data.
	fmt.Println(d.userActivityLogList(nil, dumpTCX))
	fmt.Println(d.userActivityCaloriesTimeseries(&startDate, &endDate))
	fmt.Println(d.userBMITimeseries(&startDate, &endDate))
	fmt.Println(d.userBodyFatTimeseries(&startDate, &endDate))
	fmt.Println(d.userBodyWeightTimeseries(&startDate, &endDate))
	fmt.Println(d.userCaloriesBMRTimeseries(&startDate, &endDate))
	fmt.Println(d.userCaloriesTimeseries(&startDate, &endDate))
	fmt.Println(d.userDistanceTimeseries(&startDate, &endDate))
	fmt.Println(d.userFloorsTimeseries(&startDate, &endDate))
	fmt.Println(d.userMinutesFairlyActiveTimeseries(&startDate, &endDate))
	fmt.Println(d.userMinutesLightlyActiveTimeseries(&startDate, &endDate))
	fmt.Println(d.userMinutesSedentaryTimeseries(&startDate, &endDate))
	fmt.Println(d.userMinutesVeryActiveTimeseries(&startDate, &endDate))
	fmt.Println(d.userStepsTimeseries(&startDate, &endDate))
	fmt.Println(d.userHeartRateTimeseries(&startDate, &endDate))
	fmt.Println(d.userSleepLogList(&startDate, &endDate))
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
			dumpTCX := false
			if err := dumper.Dump(all, dumpTCX); err != nil {
				fmt.Printf("dumper.Dump(all): %s", err)
			}
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
}
