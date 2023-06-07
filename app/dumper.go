package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/galeone/fitbit"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb"
	fitbit_client "github.com/galeone/fitbit/client"
	fitbit_types "github.com/galeone/fitbit/types"
	"github.com/galeone/fitsleepinsights/database"
	"github.com/galeone/fitsleepinsights/database/types"
	"github.com/galeone/tcx"
	"github.com/labstack/echo/v4"
)

func init() {
	_ = _db.Listen(database.NewUsersChannel, func(payload ...string) {
		log.Println("notification received")
		if len(payload) != 1 {
			panic(fmt.Sprintf("Expected 1 payload on %s, got %d", database.NewUsersChannel, len(payload)))
		}
		accessToken := payload[0]
		if dumper, err := NewDumper(accessToken); err == nil {
			if err := dumper.DumpNewer(); err != nil {
				log.Printf("dumper.DumpNewer: %s", err)
			}
		} else {
			log.Println("here: ", err.Error(), "at received: ", accessToken)
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

func (d *dumper) userActivityCaloriesTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.ActivityCaloriesSeries
	if value, err = d.fb.UserActivityCaloriesTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.ActivityCaloriesSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.ActivityCaloriesSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
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
				log.Println("skipping activity ", activity.LogID, ": already present")
				continue
			}

			tx := _db.Begin()
			// There are activities without active zone minutes
			if activity.ActiveZoneMinutes.TotalMinutes > 0 {
				activeZoneMinutes := types.ActiveZoneMinutes{
					ActiveZoneMinutes: activity.ActiveZoneMinutes,
				}

				if err = tx.Create(&activeZoneMinutes); err != nil {
					log.Println(err)
					break
				}
				for _, minInHRZone := range activity.ActiveZoneMinutes.MinutesInHeartRateZones {
					minInHRZoneRow := types.MinutesInHeartRateZone{
						MinutesInHeartRateZone: minInHRZone,
						ActiveZoneMinutesID:    activeZoneMinutes.ID,
					}
					if err = tx.Create(&minInHRZoneRow); err != nil {
						log.Println(err)
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
						log.Println(err)
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
						log.Println(err)
					} else {
						activityRow.Tcx = sql.NullString{
							String: string(textBytes),
							Valid:  true,
						}
					}
				} else {
					log.Println(err)
					// Do not break: who cares about failing fetch of TCX data (Fitbit has several problems with that)
				}
			}

			if err = tx.Create(&activityRow); err != nil {
				log.Println(err)
				break
			}

			// Once we have the activity stored, we can save the array types returned by the API
			for _, activityLevel := range activity.ActivityLevel {
				activityLevelRow := types.LoggedActivityLevel{
					LoggedActivityLevel: activityLevel,
					ActivityLogID:       activityRow.LogID,
				}
				if err = tx.Create(&activityLevelRow); err != nil {
					log.Println(err)
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
					log.Println(err)
					break
				}
			}

			if err = tx.Commit(); err != nil {
				log.Println(err)
				break
			}
		}

		// An empty url is a valid url for url.Parse!
		if value.Pagination.Next == "" {
			break
		}
		if nextURL, err := url.Parse(value.Pagination.Next); err != nil {
			log.Println(err)
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
		log.Println(err)
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
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.BMISeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.BMISeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userBodyFatTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.BodyFatSeries
	if value, err = d.fb.UserBodyFatTimeSeries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.BodyFatSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.BodyFatSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userBodyWeightTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.BodyWeightSeries
	if value, err = d.fb.UserBodyWeightTimeSeries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.BodyWeightSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.BodyWeightSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userCaloriesBMRTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.CaloriesBMRSeries
	if value, err = d.fb.UserCaloriesBMRTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.CaloriesBMRSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.CaloriesBMRSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userCaloriesTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.CaloriesSeries
	if value, err = d.fb.UserCaloriesTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.CaloriesSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.CaloriesSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userDistanceTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.DistanceSeries
	if value, err = d.fb.UserDistanceTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.DistanceSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.DistanceSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userFloorsTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.FloorsSeries
	if value, err = d.fb.UserFloorsTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.FloorsSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.FloorsSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userHeartRateTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.HeartRateSeries
	if value, err = d.fb.UserHeartRateTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, hrActivity := range value.ActivitiesHeart {
		hrActivityInsert := types.HeartRateActivities{
			UserID: d.User.ID,
			Date:   hrActivity.DateTime.Time,
		}
		if hrActivity.Value.RestingHeartRate > 0 {
			hrActivityInsert.RestingHeartRate = sql.NullInt64{
				Valid: true,
				Int64: hrActivity.Value.RestingHeartRate,
			}
		} else {
			hrActivityInsert.RestingHeartRate = sql.NullInt64{
				Valid: false,
			}
		}

		// No error = found
		if err = tx.Model(types.HeartRateActivities{}).Where(&hrActivityInsert).Scan(&hrActivityInsert); err == nil {
			log.Println("Skipping ", hrActivityInsert)
			continue
		}

		if err = tx.Create(&hrActivityInsert); err != nil {
			log.Println(err)
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
				log.Println(err)
				break
			}
		}
		for _, customHrZone := range hrActivity.Value.CustomHeartRateZones {
			if err := insertHrZone(&customHrZone, "CUSTOM"); err != nil {
				log.Println(err)
				break
			}
		}
	}

	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userMinutesFairlyActiveTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.MinutesFairlyActiveSeries
	if value, err = d.fb.UserMinutesFairlyActiveTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.MinutesFairlyActiveSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.MinutesFairlyActiveSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userMinutesLightlyActiveTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.MinutesLightlyActiveSeries
	if value, err = d.fb.UserMinutesLightlyActiveTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.MinutesLightlyActiveSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.MinutesLightlyActiveSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userMinutesSedentaryTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.MinutesSedentarySeries
	if value, err = d.fb.UserMinutesSedentaryTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.MinutesSedentarySeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.MinutesSedentarySeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userMinutesVeryActiveTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.MinutesVeryActiveSeries
	if value, err = d.fb.UserMinutesVeryActiveTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.MinutesVeryActiveSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.MinutesVeryActiveSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userStepsTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.StepsSeries
	if value, err = d.fb.UserStepsTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.StepsSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.StepsSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userElevationTimeseries(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.ElevationSeries
	if value, err = d.fb.UserElevationTimeseries(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TimeSeries {
		timestep := types.ElevationSeries{}
		timestep.UserID = d.User.ID
		timestep.Date = t.DateTime.Time
		if timestep.Value, err = strconv.ParseFloat(t.Value, 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.ElevationSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userCoreTemperature(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.CoreTemperature
	if value, err = d.fb.UserCoreTemperature(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TempCore {
		timestep := types.CoreTemperature{
			CoreTemperatureTimePoint: t,
			UserID:                   d.User.ID,
			Date:                     t.DateTime.Time,
		}

		// No error = found
		if err = tx.Model(types.CoreTemperature{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userSkinTemperature(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.SkinTemperature
	if value, err = d.fb.UserSkinTemperature(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.TempSkin {
		timestep := types.SkinTemperature{
			SkinTemperatureTimePoint: t,
			UserID:                   d.User.ID,
			Date:                     t.DateTime.Time,
			Value:                    t.Value.NightlyRelative,
		}

		// No error = found
		if err = tx.Model(types.SkinTemperature{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userCardioFitnessScore(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.CardioFitnessScore
	if value, err = d.fb.UserCardioFitnessScore(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.CardioScore {
		timestep := types.CardioFitnessScore{
			CardioScoreTimePoint: t,
			UserID:               d.User.ID,
			Date:                 t.DateTime.Time,
		}
		vo2MaxRange := strings.Split(t.Value.Vo2Max, "-")
		if len(vo2MaxRange) != 2 {
			return fmt.Errorf("expected a vo2max range, got: %s", t.Value.Vo2Max)
		}
		if timestep.Vo2MaxLowerBound, err = strconv.ParseFloat(vo2MaxRange[0], 64); err != nil {
			log.Println(err)
			break
		}
		if timestep.Vo2MaxUpperBound, err = strconv.ParseFloat(vo2MaxRange[1], 64); err != nil {
			log.Println(err)
			break
		}

		// No error = found
		if err = tx.Model(types.CardioFitnessScore{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userOxygenSaturation(startDate, endDate *time.Time) (err error) {
	var values *fitbit_types.OxygenSaturations
	if values, err = d.fb.UserOxygenSaturation(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range *values {
		timestep := types.OxygenSaturation{
			UserID: d.User.ID,
			Date:   t.DateTime.Time,
			Avg:    t.Value.Avg,
			Max:    t.Value.Max,
			Min:    t.Value.Min,
		}

		// No error = found
		if err = tx.Model(types.OxygenSaturation{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userHeartRateVariability(startDate, endDate *time.Time) (err error) {
	var value *fitbit_types.HeartRateVariability
	if value, err = d.fb.UserHeartRateVariability(startDate, endDate); err != nil {
		log.Println(err)
		return
	}
	tx := _db.Begin()
	for _, t := range value.Hrv {
		timestep := types.HeartRateVariabilityTimeSeries{
			UserID:     d.User.ID,
			Date:       t.DateTime.Time,
			DailyRmssd: t.Value.DailyRmssd,
			DeepRmssd:  t.Value.DeepRmssd,
		}

		// No error = found
		if err = tx.Model(types.HeartRateVariabilityTimeSeries{}).Where(&timestep).Scan(&timestep); err == nil {
			log.Println("Skipping ", t)
			continue
		}
		if err = tx.Create(&timestep); err != nil {
			log.Println(err)
			break
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}

	return
}

func (d *dumper) userSleepLogList(startDate, endDate *time.Time) (err error) {
	var sleepLogs *fitbit_types.SleepLogs
	if sleepLogs, err = d.fb.UserSleepLog(startDate, endDate); err != nil {
		log.Println(err)
		return
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
			log.Println("Skipping ", insert)
			continue
		}
		if err = tx.Create(&insert); err != nil {
			log.Println(err)
			break
		}

		sleepStage := func(stage *fitbit_types.SleepStageDetail, name string) error {
			insertStage := types.SleepStageDetail{
				SleepStageDetail: *stage,
				SleepLogID:       insert.LogID,
				SleepStage:       name,
			}
			return tx.Create(&insertStage)
		}

		if err = sleepStage(&sleepLog.Levels.Summary.Deep, "DEEP"); err != nil {
			log.Println(err)
			break
		}
		if err = sleepStage(&sleepLog.Levels.Summary.Light, "LIGHT"); err != nil {
			log.Println(err)
			break
		}
		if err = sleepStage(&sleepLog.Levels.Summary.Rem, "REM"); err != nil {
			log.Println(err)
			break
		}
		if err = sleepStage(&sleepLog.Levels.Summary.Wake, "WAKE"); err != nil {
			log.Println(err)
			break
		}

		sleepData := func(data []fitbit_types.SleepData) error {
			for _, sleepData := range data {
				levelDataInsert := types.SleepData{
					SleepData:  sleepData,
					SleepLogID: sleepLog.LogID,
					DateTime:   sleepData.DateTime.Time,
				}
				if err := tx.Create(&levelDataInsert); err != nil {
					return err
				}
			}
			return nil
		}

		if err = sleepData(sleepLog.Levels.Data); err != nil {
			log.Println(err)
		}
		if err = sleepData(sleepLog.Levels.ShortData); err != nil {
			log.Println(err)
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
	}
	return
}

// Dump fetches every data available on the user profile, up to this moment.
// This function is called:
//   - When the user gives the permission to the app (on the INSERT on the table
//     triggered by the database notification)
//   - Periodically by a go routine. In this case, the `after` variable is valid.
func (d *dumper) Dump(after *time.Time, dumpTCX bool) error {
	var startDate time.Time
	var endDate *time.Time

	// ALWAYS dump data up to yesterday, since this is complete data.
	// Today data is changing.
	today := time.Now().Truncate(time.Hour * 24)
	yesterday := time.Now().Add(-time.Duration(24) * time.Hour).Truncate(time.Hour * 24)
	endDate = &yesterday

	dumpAll := after == nil
	var days int
	if dumpAll {
		// In this case, we want to dump "all" the past data up to yesterday.
		// Try to fetch a completely arbitrary number of days.
		days = 365
		startDate = endDate.Add(-time.Duration(24*120) * time.Hour).Truncate(time.Hour * 24)
	} else {
		startDate = *after
		days = int(yesterday.Sub(startDate).Hours()) / 24
	}

	// There are functions that don't have an "after" period
	// because Fitbit allows to get only the daily data.
	d.userActivityDailyGoal()
	d.userActivityWeeklyGoal()

	d.userActivityLogList(&startDate, dumpTCX)
	d.userActivityCaloriesTimeseries(&startDate, endDate)
	d.userBMITimeseries(&startDate, endDate)
	d.userBodyFatTimeseries(&startDate, endDate)
	d.userBodyWeightTimeseries(&startDate, endDate)
	d.userCaloriesBMRTimeseries(&startDate, endDate)
	d.userCaloriesTimeseries(&startDate, endDate)
	d.userDistanceTimeseries(&startDate, endDate)
	d.userFloorsTimeseries(&startDate, endDate)
	d.userMinutesFairlyActiveTimeseries(&startDate, endDate)
	d.userMinutesLightlyActiveTimeseries(&startDate, endDate)
	d.userMinutesSedentaryTimeseries(&startDate, endDate)
	d.userMinutesVeryActiveTimeseries(&startDate, endDate)
	d.userStepsTimeseries(&startDate, endDate)
	d.userHeartRateTimeseries(&startDate, endDate)
	d.userElevationTimeseries(&startDate, endDate)

	if days > 30 {
		// NOTE: every loop should loop using "gcd" days
		// from startDate to endDate to do not lose days.
		gcd := func(a, b int) int {
			for b != 0 {
				t := b
				b = a % b
				a = t
			}
			return a
		}

		// Only last 30 days for Skin/Core temp & Oxygen saturation
		ago := gcd(days, 30)
		newEndDate := startDate.Add(time.Duration(ago*24) * time.Hour)
		newStartDate := startDate
		for newEndDate.Before(today) {
			d.userSkinTemperature(&newStartDate, &newEndDate)
			d.userCoreTemperature(&newStartDate, &newEndDate)
			d.userOxygenSaturation(&newStartDate, &newEndDate)
			d.userCardioFitnessScore(&newStartDate, &newEndDate)
			d.userHeartRateVariability(&newStartDate, &newEndDate)
			newStartDate = newEndDate
			newEndDate = newEndDate.Add(time.Duration(ago*24) * time.Hour)
		}
		// 100 days for SleepLogList.
		ago = gcd(days, 100)
		newEndDate = startDate.Add(time.Duration(ago*24) * time.Hour)
		newStartDate = startDate
		for newEndDate.Before(today) {
			d.userSleepLogList(&newStartDate, &newEndDate)
			newStartDate = newEndDate
			newEndDate = newEndDate.Add(time.Duration(ago*24) * time.Hour)
		}

	} else {
		d.userSkinTemperature(&startDate, endDate)
		d.userCoreTemperature(&startDate, endDate)
		d.userOxygenSaturation(&startDate, endDate)
		d.userCardioFitnessScore(&startDate, endDate)
		d.userHeartRateVariability(&startDate, endDate)
		// if days <= 30 -> days <= 100 (lol)
		d.userSleepLogList(&startDate, endDate)
	}

	return nil
}

func (d *dumper) DumpNewer() error {
	// Fetch the latest activity logged to get a date to start from.
	var after *time.Time
	var last time.Time
	// err = empty -> no data previously stored
	if err := _db.Model(types.ActivityLog{}).Select("max(start_time)").Where(&types.ActivityLog{UserID: d.User.ID}).Scan(&last); err != nil {
		after = nil
	} else {
		after = &last
	}
	dumpTCX := false
	return d.Dump(after, dumpTCX)
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
			if err = dumper.DumpNewer(); err != nil {
				return err
			}
		} else {
			log.Println(err.Error())
		}
		return err
	}
}
