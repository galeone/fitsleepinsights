// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/galeone/fitbit/v2"
	"github.com/galeone/fitsleepinsights/database/types"
	"github.com/labstack/echo/v4"
)

func TestTrainAndDeploy() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// 1. Fetch all user data
		authorizer := c.Get("fitbit").(*fitbit.Authorizer)
		var userID *string
		if userID, err = authorizer.UserID(); err != nil {
			return err
		}

		user := types.User{}
		user.UserID = *userID
		if err = _db.Model(types.User{}).Where(&user).Scan(&user); err != nil {
			return err
		}

		err = TrainAndDeployPredictor(&user, "SleepEfficiency")
		for err != nil {
			if strings.Contains(err.Error(), "DeadlineExceeded") {
				log.Println("DeadlineExceeded, retrying...")
				err = TrainAndDeployPredictor(&user, "SleepEfficiency")
			} else {
				break
			}
		}
		return
	}
}

type PredictionResult struct {
	// The prediction result
	Prediction []float32 `json:"prediction"`
	Target     string    `json:"target"`
}

func toFloat32(input []uint8) []float32 {
	output := make([]float32, len(input))
	for i, v := range input {
		output[i] = float32(v)
	}
	return output
}

func TestPredictSleepEfficiency() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// 1. Fetch all user data
		authorizer := c.Get("fitbit").(*fitbit.Authorizer)
		var userID *string
		if userID, err = authorizer.UserID(); err != nil {
			return err
		}

		user := types.User{}
		user.UserID = *userID
		if err = _db.Model(types.User{}).Where(&user).Scan(&user); err != nil {
			return err
		}

		var fetcher *fetcher
		if fetcher, err = NewFetcher(&user); err != nil {
			return err
		}

		todayData, _ := fetcher.FetchByDate(time.Now())

		var sleepEfficiency []uint8
		if sleepEfficiency, err = PredictSleepEfficiency(&user, []*UserData{todayData}); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, PredictionResult{
			Prediction: toFloat32(sleepEfficiency),
			Target:     "SleepEfficiency",
		})
	}
}
