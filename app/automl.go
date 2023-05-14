// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"github.com/labstack/echo/v4"
)

func TestAutoML() echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
	/*
	   // 3. Use automl to train the model
	   // 4. Use automl to deploy the model
	   // 5. Save the model id and endpoint id in the database

	   	return func(c echo.Context) (err error) {
	   		// 1. Fetch all user data
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

	   		var fetcher *fetcher
	   		if fetcher, err = NewFetcher(&user); err != nil {
	   			return err
	   		}
	   		var allUserData []*UserData
	   		if allUserData, err = fetcher.FetchAll(); err != nil {
	   			return err
	   		}

	   		// 2. Prepare training data
	   		// ref: https://cloud.google.com/vertex-ai/docs/tabular-data/classification-regression/prepare-data#csv

	   		ctx := context.Background()

	   		// 3. Create a dataset from the training data (associate the bucket with the dataset)

	   		var client *automl.DatasetClient
	   		base := fmt.Sprintf("%s-aiplatform.googleapis.com:443", _vaiLocation)
	   		if client, err = automl.NewDatasetClient(ctx, option.WithEndpoint(base)); err != nil {
	   			return err
	   		}
	   		defer client.Close()

	   		req := &automlpb.ExportDataRequest{
	   			Name: "projects/fitbit-ml/locations/us-central1/datasets/TBL5916027185625631744",
	   			OutputConfig: &automlpb.OutputConfig{
	   				Destination: &automlpb.OutputConfig_GcsDestination{
	   					GcsDestination: &automlpb.GcsDestination{
	   						OutputUriPrefix: "gs://fitbit-ml",
	   					},
	   				},
	   			},
	   		}

	   		return nil
	   	}
	*/
}
