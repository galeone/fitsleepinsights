// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"context"
	"fmt"

	vai "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestVertexAI() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.Background()

		base := fmt.Sprintf("%s-aiplatform.googleapis.com:443", _vaiLocation)
		client, err := vai.NewPredictionClient(ctx, option.WithEndpoint(base))

		if err != nil {
			return err
		}
		defer client.Close()

		instance, err := structpb.NewValue(map[string]interface{}{
			"price":           5250000,
			"area":            5500,
			"bathrooms":       2,
			"stories":         1,
			"mainroad":        1,
			"guestroom":       0,
			"basement":        1,
			"hotwaterheating": 0,
			"airconditioning": 0,
			"parking":         0,
			"prefarea":        0,
			"semi-furnished":  1,
			"unfurnished":     0,
			"areaperbedroom":  1833.333333,
			"bbratio":         0.666666667,
		})
		if err != nil {
			return nil
		}

		instances := []*structpb.Value{
			instance,
		}

		response, err := client.Predict(ctx, &aiplatformpb.PredictRequest{
			Instances: instances,
			Endpoint:  fmt.Sprintf("projects/%s/locations/%s/endpoints/%s", _vaiProjectID, _vaiLocation, _vaiEndpointID),
		})
		if err != nil {
			return err
		}
		fmt.Println(response.String())

		return nil
	}

}
