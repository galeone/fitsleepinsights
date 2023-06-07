package app

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/galeone/fitbit"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb"
	fitbit_types "github.com/galeone/fitbit/types"
	"github.com/galeone/fitsleepinsights/database/types"
	"github.com/labstack/echo/v4"
)

type fetcher struct {
	user *fitbit_pgdb.AuthorizedUser
}

type DailyActivities []types.ActivityLog

// Headers returns the headers of the CSV file.
// For the DailyActivities type, the headers are only the column names that can
// be aggregated using a sum (e.g. no activityID, no activityParentID, etc.)
func (DailyActivities) Headers() []string {
	// TODO: this is a coarse summary of the activities, but a lot of useful information is missing
	return []string{
		"NumberOfAutoDetectedActivities",
		"NumberOfTrackerActivities",

		"ActiveDurationSum",
		"ActiveZoneMinutesSum",

		"MinutesInCardioZoneSum",
		"MinutesInFatBurnZoneSum",
		"MinutesInPeakZoneSum",
		"MinutesInOutOfZoneSum",

		// The names are concatenated without repetitions
		// So if there are: walk, weights, walk, weights, walk, run
		// the result will be: walk, weights, run
		"ActivitiesNameConcatenation",

		// For these fields the sum is computed only
		"CaloriesSum",
		"DistanceSum",
		"DurationSum",
		"ElevationGainSum",
		"StepsSum",

		// For these fields the average is computed only
		// on the activities that have a valid value for the field
		"AveragePace",
		"AverageSpeed",
		"AverageHeartRate",
	}
}

func (f *DailyActivities) Values() []string {
	var NumberOfAutoDetectedActivities, NumberOfTrackerActivities int64
	var ActiveDurationSum, ActiveZoneMinutesSum int64

	var ActivitiesNameConcatenation string
	var names map[string]bool = make(map[string]bool)

	var MinutesInCardioZoneSum, MinutesInFatBurnZoneSum, MinutesInPeakZoneSum, MinutesInOutOfZoneSum int64

	var CaloriesSum, DurationSum, ElevationGainSum, StepsSum int64
	var DistanceSum float64

	var paces, speeds, heartRates []float64
	var AveragePace, AverageSpeed, AverageHeartRate float64
	for _, activity := range *f {
		// check if activity.ActivityName is a key of names
		// if not, add it to the names map
		if _, ok := names[activity.ActivityName]; !ok {
			names[activity.ActivityName] = true
			ActivitiesNameConcatenation += activity.ActivityName + " "
		}
		if activity.LogType == "tracker" {
			NumberOfTrackerActivities++
		}
		if activity.LogType == "auto_detected" {
			NumberOfAutoDetectedActivities++
		}

		ActiveDurationSum += activity.ActiveDuration
		ActiveZoneMinutesSum += activity.ActiveZoneMinutes.TotalMinutes

		for _, activeZoneMinute := range activity.ActiveZoneMinutes.MinutesInHeartRateZones {
			if activeZoneMinute.Type == "CARDIO" {
				MinutesInCardioZoneSum += activeZoneMinute.Minutes
			}
			if activeZoneMinute.Type == "FAT_BURN" {
				MinutesInFatBurnZoneSum += activeZoneMinute.Minutes
			}
			if activeZoneMinute.Type == "PEAK" {
				MinutesInPeakZoneSum += activeZoneMinute.Minutes
			}
			if activeZoneMinute.Type == "OUT_OF_ZONE" {
				MinutesInOutOfZoneSum += activeZoneMinute.Minutes
			}
		}
		CaloriesSum += activity.Calories
		DistanceSum += activity.Distance
		DurationSum += activity.Duration
		ElevationGainSum += activity.ElevationGain
		StepsSum += activity.Steps

		if activity.Pace > 0 {
			paces = append(paces, activity.Pace)
		}
		if activity.Speed > 0 {
			speeds = append(speeds, activity.Speed)
		}
		if activity.AverageHeartRate > 0 {
			heartRates = append(heartRates, float64(activity.AverageHeartRate))
		}
	}

	reduceMean := func(values []float64) float64 {
		var sum float64
		for _, value := range values {
			sum += value
		}
		return sum / float64(len(values))
	}

	if len(paces) > 0 {
		AveragePace = reduceMean(paces)
	}
	if len(speeds) > 0 {
		AverageSpeed = reduceMean(speeds)
	}
	if len(heartRates) > 0 {
		AverageHeartRate = reduceMean(heartRates)
	}

	// Split ActivitiesNameConcatenation using " " as delimiter and sort it alphabetically
	// This is done to have a consistent order of the activities
	// e.g. walk, weights, walk, weights, walk, run
	// becomes: run, walk, walk, walk, weights, weights
	// and then: run, walk, weights
	// Duplicates are already handled during the creation of ActivitiesNameConcatenation
	ActivitiesNameConcatenation = strings.Trim(ActivitiesNameConcatenation, " ")
	activities := strings.Split(ActivitiesNameConcatenation, " ")
	sort.Strings(activities)
	ActivitiesNameConcatenation = strings.Join(activities, " ")

	return []string{
		fmt.Sprintf("%d", NumberOfAutoDetectedActivities),
		fmt.Sprintf("%d", NumberOfTrackerActivities),

		fmt.Sprintf("%d", ActiveDurationSum),
		fmt.Sprintf("%d", ActiveZoneMinutesSum),

		fmt.Sprintf("%d", MinutesInCardioZoneSum),
		fmt.Sprintf("%d", MinutesInFatBurnZoneSum),
		fmt.Sprintf("%d", MinutesInPeakZoneSum),
		fmt.Sprintf("%d", MinutesInOutOfZoneSum),

		ActivitiesNameConcatenation,

		fmt.Sprintf("%d", CaloriesSum),
		strconv.FormatFloat(DistanceSum, 'f', 2, 64),
		fmt.Sprintf("%d", DurationSum),
		fmt.Sprintf("%d", ElevationGainSum),
		fmt.Sprintf("%d", StepsSum),

		strconv.FormatFloat(AveragePace, 'f', 2, 64),
		strconv.FormatFloat(AverageSpeed, 'f', 2, 64),
		strconv.FormatFloat(AverageHeartRate, 'f', 2, 64),
	}
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

func (f *fetcher) userActivityLogList(date time.Time) (*DailyActivities, error) {
	activities := DailyActivities{}

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
	Activities           *DailyActivities
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

// Headers returns the headers of the CSV file
// The order of the headers is important as it will be used to generate the CSV file
// The order of the values must match the order of the headers
// If a value is nil, the CSV cell will be empty
func (UserData) Headers() []string {
	ret := []string{
		"Date",
	}
	ret = append(ret, DailyActivities{}.Headers()...)
	ret = append(ret, types.ActivityCaloriesSeries{}.Headers()...)
	ret = append(ret, types.BMISeries{}.Headers()...)
	ret = append(ret, types.BodyFatSeries{}.Headers()...)
	ret = append(ret, types.BodyWeightSeries{}.Headers()...)
	ret = append(ret, types.CaloriesBMRSeries{}.Headers()...)
	ret = append(ret, types.CaloriesSeries{}.Headers()...)
	ret = append(ret, types.DistanceSeries{}.Headers()...)
	ret = append(ret, types.FloorsSeries{}.Headers()...)
	ret = append(ret, types.MinutesFairlyActiveSeries{}.Headers()...)
	ret = append(ret, types.MinutesLightlyActiveSeries{}.Headers()...)
	ret = append(ret, types.MinutesSedentarySeries{}.Headers()...)
	ret = append(ret, types.MinutesVeryActiveSeries{}.Headers()...)
	ret = append(ret, types.StepsSeries{}.Headers()...)
	ret = append(ret, types.HeartRateActivities{}.Headers()...)
	ret = append(ret, types.ElevationSeries{}.Headers()...)
	ret = append(ret, types.SkinTemperature{}.Headers()...)
	ret = append(ret, types.CoreTemperature{}.Headers()...)
	ret = append(ret, types.OxygenSaturation{}.Headers()...)
	ret = append(ret, types.CardioFitnessScore{}.Headers()...)
	ret = append(ret, types.HeartRateVariabilityTimeSeries{}.Headers()...)
	ret = append(ret, types.SleepLog{}.Headers()...)
	return ret
}

func (u *UserData) Values() []string {
	ret := []string{
		// Date format required by Vertex AI
		u.Date.Format(time.RFC3339),
	}

	if u.Activities == nil {
		ret = append(ret, make([]string, len(DailyActivities{}.Headers()))...)
	} else {
		ret = append(ret, u.Activities.Values()...)
	}

	if u.ActivityCalories == nil {
		ret = append(ret, make([]string, len(types.ActivityCaloriesSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.ActivityCalories.Values()...)
	}

	if u.BMI == nil {
		ret = append(ret, make([]string, len(types.BMISeries{}.Headers()))...)
	} else {
		ret = append(ret, u.BMI.Values()...)
	}

	if u.BodyFat == nil {
		ret = append(ret, make([]string, len(types.BodyFatSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.BodyFat.Values()...)
	}

	if u.BodyWeight == nil {
		ret = append(ret, make([]string, len(types.BodyWeightSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.BodyWeight.Values()...)
	}

	if u.CaloriesBMR == nil {
		ret = append(ret, make([]string, len(types.CaloriesBMRSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.CaloriesBMR.Values()...)
	}

	if u.Calories == nil {
		ret = append(ret, make([]string, len(types.CaloriesSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.Calories.Values()...)
	}

	if u.Distance == nil {
		ret = append(ret, make([]string, len(types.DistanceSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.Distance.Values()...)
	}

	if u.Floors == nil {
		ret = append(ret, make([]string, len(types.FloorsSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.Floors.Values()...)
	}

	if u.MinutesFairlyActive == nil {
		ret = append(ret, make([]string, len(types.MinutesFairlyActiveSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.MinutesFairlyActive.Values()...)
	}

	if u.MinutesLightlyActive == nil {
		ret = append(ret, make([]string, len(types.MinutesLightlyActiveSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.MinutesLightlyActive.Values()...)
	}

	if u.MinutesSedentary == nil {
		ret = append(ret, make([]string, len(types.MinutesSedentarySeries{}.Headers()))...)
	} else {
		ret = append(ret, u.MinutesSedentary.Values()...)
	}

	if u.MinutesVeryActive == nil {
		ret = append(ret, make([]string, len(types.MinutesVeryActiveSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.MinutesVeryActive.Values()...)
	}

	if u.Steps == nil {
		ret = append(ret, make([]string, len(types.StepsSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.Steps.Values()...)
	}

	if u.HeartRate == nil {
		ret = append(ret, make([]string, len(types.HeartRateActivities{}.Headers()))...)
	} else {
		ret = append(ret, u.HeartRate.Values()...)
	}

	if u.Elevation == nil {
		ret = append(ret, make([]string, len(types.ElevationSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.Elevation.Values()...)
	}

	if u.SkinTemperature == nil {
		ret = append(ret, make([]string, len(types.SkinTemperature{}.Headers()))...)
	} else {
		ret = append(ret, u.SkinTemperature.Values()...)

	}

	if u.CoreTemperature == nil {
		ret = append(ret, make([]string, len(types.CoreTemperature{}.Headers()))...)
	} else {
		ret = append(ret, u.CoreTemperature.Values()...)
	}

	if u.OxygenSaturation == nil {
		ret = append(ret, make([]string, len(types.OxygenSaturation{}.Headers()))...)
	} else {
		ret = append(ret, u.OxygenSaturation.Values()...)
	}

	if u.CardioFitnessScore == nil {
		ret = append(ret, make([]string, len(types.CardioFitnessScore{}.Headers()))...)
	} else {

		ret = append(ret, u.CardioFitnessScore.Values()...)
	}

	if u.HeartRateVariability == nil {
		ret = append(ret, make([]string, len(types.HeartRateVariabilityTimeSeries{}.Headers()))...)
	} else {
		ret = append(ret, u.HeartRateVariability.Values()...)
	}

	if u.SleepLog == nil {
		ret = append(ret, make([]string, len(types.SleepLog{}.Headers()))...)
	} else {
		ret = append(ret, u.SleepLog.Values()...)
	}
	return ret
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

type FetchAllStrategy int

const (
	FetchAllWithSleepLog FetchAllStrategy = iota
	FetchAllWithActivityLog
)

// FetchAll fetches all the user data. It uses the oldest sleep log date as first date
// and yesterday as last date.
func (f *fetcher) FetchAll(strategy FetchAllStrategy) ([]*UserData, error) {
	// Get all the dates
	var dates []time.Time
	yesterday := time.Now().AddDate(0, 0, -1).Truncate(time.Hour * 24)
	switch strategy {
	case FetchAllWithSleepLog:
		// Condition on efficiency > 0 to avoid fetching sleep logs that are not complete
		if err := _db.Model(types.SleepLog{}).Select("distinct date(date_of_sleep) as d").Where(`date_of_sleep <= ? AND user_id = ? AND efficiency > 0`, yesterday, f.user.ID).Order("d desc").Scan(&dates); err != nil {
			return nil, err
		}
	case FetchAllWithActivityLog:
		if err := _db.Model(types.ActivityLog{}).Select("distinct date(start_time) as d").Where(`start_time <= ? AND user_id = ?`, yesterday, f.user.ID).Order("d desc").Scan(&dates); err != nil {
			return nil, err
		}
	}
	if len(dates) == 0 {
		return nil, errors.New("user has zero activities")
	}

	var userDataList []*UserData
	for _, date := range dates {
		userData := f.Fetch(date)
		userDataList = append(userDataList, &userData)
	}
	return userDataList, nil
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
			if all, err := fetcher.FetchAll(FetchAllWithSleepLog); err == nil {
				if csv, err := userDataToCSV(all); err == nil {
					// Save completecsv to file
					file, _ := os.Create("complete.csv")
					io.WriteString(file, csv)
				} else {
					log.Println(err)
				}
			} else {
				log.Println(err)
			}
		} else {
			log.Println(err.Error())
		}

		return err
	}
}
