package app

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/galeone/fitbit"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb"
	fitbit_types "github.com/galeone/fitbit/types"
	"github.com/galeone/sleepbit/database/types"
	"github.com/labstack/echo/v4"
)

type fetcher struct {
	user *fitbit_pgdb.AuthorizedUser
}

// NewFetcher creates a new fetcher for the provided user
func NewFetcher(user *fitbit_pgdb.AuthorizedUser) (*fetcher, error) {
	if user == nil {
		return nil, errors.New("expected a valid user, nil given instead")
	}
	if user.ID == 0 {
		return nil, errors.New("expected a valid user with a valid ID. The provided user has ID = 0")
	}

	return &fetcher{user}, nil
}

func (f *fetcher) Date(date time.Time) {
}

func (f *fetcher) userActivityCaloriesTimeseries(date time.Time) (*types.ActivityCaloriesSeries, error) {
	value := types.ActivityCaloriesSeries{UserID: f.user.ID, Date: date}
	if err := _db.Model(types.ActivityCaloriesSeries{}).Where(&value).Scan(&value); err != nil {
		return nil, err
	}

	return &value, nil
}

func (f *fetcher) userActivityDailyGoal(date time.Time) (*types.Goal, error) {
	value := types.Goal{
		UserID:    f.user.ID,
		StartDate: date,
		EndDate:   date,
	}

	if err := _db.Model(types.Goal{}).Where(&value).Scan(&value); err != nil {
		return nil, err
	}
	return &value, nil
}

func (f *fetcher) userActivityLogList(date time.Time) (*[]types.ActivityLog, error) {
	activities := []types.ActivityLog{}

	if err := _db.Model(types.ActivityLog{}).Where(`user_id = ? AND date(start_time) = ?`, f.user.ID, date.Format(fitbit_types.DateLayout)).Scan(&activities); err != nil {
		return nil, err
	}

	for id, activity := range activities {
		if activity.ActiveZoneMinutesID.Valid {
			minutesInHRZone := []types.MinutesInHeartRateZone{}
			condition := types.MinutesInHeartRateZone{
				ActiveZoneMinutesID: activity.ActiveZoneMinutesID.Int64,
			}

			if err := _db.Model(types.MinutesInHeartRateZone{}).Where(&condition).Scan(&minutesInHRZone); err != nil {
				return nil, err
			}
			for _, minInHRZone := range minutesInHRZone {
				activities[id].ActiveZoneMinutes.MinutesInHeartRateZones = append(
					activities[id].ActiveZoneMinutes.MinutesInHeartRateZones, minInHRZone.MinutesInHeartRateZone)
			}
		}

		if activity.SourceID.Valid {
			source := types.LogSource{
				ID: activity.SourceID.String,
			}
			if err := _db.Model(types.LogSource{}).Where(&source).Scan(&source); err != nil {
				return nil, err
			}
			activities[id].Source.LogSource = source.LogSource
		}

		// Ignore errors: there could be activities without heart rate zones
		_ = _db.Model(types.HeartRateZone{}).Where(types.HeartRateZone{
			ActivityLogID: sql.NullInt64{
				Int64: activity.LogID,
				Valid: true,
			},
		}).Scan(&activities[id].HeartRateZones)
	}

	return &activities, nil
}

func (f *fetcher) userActivityWeeklyGoal(date time.Time) (*types.Goal, error) {
	value := types.Goal{}

	only_date := date.Format(fitbit_types.DateLayout)
	if err := _db.Model(types.Goal{}).Where(`user_id = ? AND start_date >= ? AND end_date <= ?`, f.user.ID, only_date, only_date).Scan(&value); err != nil {
		return nil, err
	}
	return &value, nil
}

func (f *fetcher) userBMITimeseries(date time.Time) (*types.BMISeries, error) {
	timestep := types.BMISeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.BMISeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userBodyFatTimeseries(date time.Time) (*types.BodyFatSeries, error) {
	timestep := types.BodyFatSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.BodyFatSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userBodyWeightTimeseries(date time.Time) (*types.BodyWeightSeries, error) {
	timestep := types.BodyWeightSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.BodyWeightSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userCaloriesBMRTimeseries(date time.Time) (*types.CaloriesBMRSeries, error) {
	timestep := types.CaloriesBMRSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.CaloriesBMRSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userCaloriesTimeseries(date time.Time) (*types.CaloriesSeries, error) {
	timestep := types.CaloriesSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.CaloriesSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userDistanceTimeseries(date time.Time) (*types.DistanceSeries, error) {
	timestep := types.DistanceSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.DistanceSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userFloorsTimeseries(date time.Time) (*types.FloorsSeries, error) {
	timestep := types.FloorsSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.FloorsSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userHeartRateTimeseries(date time.Time) (*types.HeartRateActivities, error) {
	hrActivity := types.HeartRateActivities{}
	hrActivity.UserID = f.user.ID
	hrActivity.Date = date
	if err := _db.Model(types.HeartRateActivities{}).Where(&hrActivity).Scan(&hrActivity); err != nil {
		return nil, err
	}
	// Ignore errors: there could be activities without HR zones
	_ = _db.Model(types.HeartRateZone{}).Where(types.HeartRateZone{
		HeartRateActivityID: sql.NullInt64{
			Int64: hrActivity.ID,
			Valid: true,
		},
		Type: "DEFAULT"}).
		Scan(&hrActivity.HeartRateZones)

	_ = _db.Model(types.HeartRateZone{}).Where(types.HeartRateZone{
		HeartRateActivityID: sql.NullInt64{
			Int64: hrActivity.ID,
			Valid: true,
		},
		Type: "CUSTOM"}).
		Scan(&hrActivity.CustomHeartRateZones)
	return &hrActivity, nil
}

func (f *fetcher) userMinutesFairlyActiveTimeseries(date time.Time) (*types.MinutesFairlyActiveSeries, error) {
	timestep := types.MinutesFairlyActiveSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.MinutesFairlyActiveSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userMinutesLightlyActiveTimeseries(date time.Time) (*types.MinutesLightlyActiveSeries, error) {
	timestep := types.MinutesLightlyActiveSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.MinutesLightlyActiveSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userMinutesSedentaryTimeseries(date time.Time) (*types.MinutesSedentarySeries, error) {
	timestep := types.MinutesSedentarySeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.MinutesSedentarySeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userMinutesVeryActiveTimeseries(date time.Time) (*types.MinutesVeryActiveSeries, error) {
	timestep := types.MinutesVeryActiveSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.MinutesVeryActiveSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userStepsTimeseries(date time.Time) (*types.StepsSeries, error) {
	timestep := types.StepsSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.StepsSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userElevationTimeseries(date time.Time) (*types.ElevationSeries, error) {
	timestep := types.ElevationSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.ElevationSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userCoreTemperature(date time.Time) (*types.CoreTemperature, error) {
	timestep := types.CoreTemperature{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.CoreTemperature{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userSkinTemperature(date time.Time) (*types.SkinTemperature, error) {
	timestep := types.SkinTemperature{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.SkinTemperature{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userCardioFitnessScore(date time.Time) (*types.CardioFitnessScore, error) {
	timestep := types.CardioFitnessScore{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.CardioFitnessScore{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userOxygenSaturation(date time.Time) (*types.OxygenSaturation, error) {
	timestep := types.OxygenSaturation{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.OxygenSaturation{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userHeartRateVariability(date time.Time) (*types.HeartRateVariabilityTimeSeries, error) {
	timestep := types.HeartRateVariabilityTimeSeries{}
	timestep.UserID = f.user.ID
	timestep.Date = date
	if err := _db.Model(types.HeartRateVariabilityTimeSeries{}).Where(&timestep).Scan(&timestep); err != nil {
		return nil, err
	}
	return &timestep, nil
}

func (f *fetcher) userSleepLogList(date time.Time) (*types.SleepLog, error) {
	value := types.SleepLog{
		UserID:      f.user.ID,
		DateOfSleep: date,
	}
	if err := _db.Model(types.SleepLog{}).Where(&value).Scan(&value); err != nil {
		return nil, err
	}

	sleepStageDetails := []types.SleepStageDetail{}
	if err := _db.Model(types.SleepStageDetail{}).Where(&types.SleepStageDetail{SleepLogID: value.LogID}).Scan(&sleepStageDetails); err != nil {
		return nil, err
	}
	for _, stage := range sleepStageDetails {
		if stage.SleepStage == "DEEP" {
			value.Levels.Summary.Deep = fitbit_types.SleepStageDetail{
				Count:               stage.Count,
				Minutes:             stage.Minutes,
				ThirtyDayAvgMinutes: stage.ThirtyDayAvgMinutes,
			}
			continue
		}
		if stage.SleepStage == "LIGHT" {
			value.Levels.Summary.Light = fitbit_types.SleepStageDetail{
				Count:               stage.Count,
				Minutes:             stage.Minutes,
				ThirtyDayAvgMinutes: stage.ThirtyDayAvgMinutes,
			}
			continue
		}
		if stage.SleepStage == "REM" {
			value.Levels.Summary.Rem = fitbit_types.SleepStageDetail{
				Count:               stage.Count,
				Minutes:             stage.Minutes,
				ThirtyDayAvgMinutes: stage.ThirtyDayAvgMinutes,
			}
			continue
		}
		if stage.SleepStage == "WAKE" {
			value.Levels.Summary.Wake = fitbit_types.SleepStageDetail{
				Count:               stage.Count,
				Minutes:             stage.Minutes,
				ThirtyDayAvgMinutes: stage.ThirtyDayAvgMinutes,
			}
			continue
		}
	}

	sleepData := []types.SleepData{}
	if err := _db.Model(types.SleepData{}).Where(&types.SleepData{
		SleepLogID: value.LogID,
	}).Scan(&sleepData); err != nil {
		return nil, err
	}
	// Data and short data is merged into data
	for _, data := range sleepData {
		value.Levels.Data = append(value.Levels.Data, fitbit_types.SleepData{
			DateTime: fitbit_types.FitbitDateTime{Time: data.DateTime},
			Level:    data.Level,
			Seconds:  data.Seconds,
		})
	}

	return &value, nil
}

// Create a struct that given all the return types of the methods used inside the Fetch method,
// is able to hold them all.
type UserData struct {
	Date                 time.Time
	Activities           *[]types.ActivityLog
	ActivityCalories     *types.ActivityCaloriesSeries
	BMI                  *types.BMISeries
	BodyFat              *types.BodyFatSeries
	BodyWeight           *types.BodyWeightSeries
	CaloriesBMR          *types.CaloriesBMRSeries
	Calories             *types.CaloriesSeries
	Distance             *types.DistanceSeries
	Floors               *types.FloorsSeries
	MinutesFairlyActive  *types.MinutesFairlyActiveSeries
	MinutesLightlyActive *types.MinutesLightlyActiveSeries
	MinutesSedentary     *types.MinutesSedentarySeries
	MinutesVeryActive    *types.MinutesVeryActiveSeries
	Steps                *types.StepsSeries
	HeartRate            *types.HeartRateActivities
	Elevation            *types.ElevationSeries
	SkinTemperature      *types.SkinTemperature
	CoreTemperature      *types.CoreTemperature
	OxygenSaturation     *types.OxygenSaturation
	CardioFitnessScore   *types.CardioFitnessScore
	HeartRateVariability *types.HeartRateVariabilityTimeSeries
	SleepLog             *types.SleepLog
}

func (f *fetcher) Fetch(date time.Time) UserData {
	userData := UserData{
		Date: date,
	}
	// All these methods should NOT return errors.
	// There are other methods that can return errors, but are not used inside Fetch.
	// e.g. the goals methods can return errors (when goals are not being set for that date).
	userData.Activities, _ = f.userActivityLogList(date)
	userData.ActivityCalories, _ = f.userActivityCaloriesTimeseries(date)
	userData.BMI, _ = f.userBMITimeseries(date)
	userData.BodyFat, _ = f.userBodyFatTimeseries(date)
	userData.BodyWeight, _ = f.userBodyWeightTimeseries(date)
	userData.CaloriesBMR, _ = f.userCaloriesBMRTimeseries(date)
	userData.Calories, _ = f.userCaloriesTimeseries(date)
	userData.Distance, _ = f.userDistanceTimeseries(date)
	userData.Floors, _ = f.userFloorsTimeseries(date)
	userData.MinutesFairlyActive, _ = f.userMinutesFairlyActiveTimeseries(date)
	userData.MinutesLightlyActive, _ = f.userMinutesLightlyActiveTimeseries(date)
	userData.MinutesSedentary, _ = f.userMinutesSedentaryTimeseries(date)
	userData.MinutesVeryActive, _ = f.userMinutesVeryActiveTimeseries(date)
	userData.Steps, _ = f.userStepsTimeseries(date)
	userData.HeartRate, _ = f.userHeartRateTimeseries(date)
	userData.Elevation, _ = f.userElevationTimeseries(date)
	userData.SkinTemperature, _ = f.userSkinTemperature(date)
	userData.CoreTemperature, _ = f.userCoreTemperature(date)
	userData.OxygenSaturation, _ = f.userOxygenSaturation(date)
	userData.CardioFitnessScore, _ = f.userCardioFitnessScore(date)
	userData.HeartRateVariability, _ = f.userHeartRateVariability(date)
	userData.SleepLog, _ = f.userSleepLogList(date)
	return userData
}

func Fetch() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// secure, under middleware
		authorizer := c.Get("fitbit").(*fitbit.Authorizer)
		var userID *string
		if userID, err = authorizer.UserID(); err != nil {
			return err
		}

		user := fitbit_pgdb.AuthorizedUser{}
		user.UserID = *userID
		if err = _db.Model(fitbit_pgdb.AuthorizedUser{}).Where(&user).Scan(&user); err != nil {
			return err
		}

		if fetcher, err := NewFetcher(&user); err == nil {
			yesterday := time.Now().Add(-time.Duration(24) * time.Hour).Truncate(time.Hour * 24)

			userData := fetcher.Fetch(yesterday)
			fmt.Println(userData.Date)
			fmt.Println(userData.HeartRate)
			fmt.Println(userData.HeartRateVariability)
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
}
