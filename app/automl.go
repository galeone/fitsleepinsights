// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	vai "cloud.google.com/go/aiplatform/apiv1beta1"
	vaipb "cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb"
	"github.com/galeone/fitbit"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	storage "cloud.google.com/go/storage"
)

func TestAutoML() echo.HandlerFunc {
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
		if allUserData, err = fetcher.FetchAll(FetchAllWithSleepLog); err != nil {
			return err
		}

		// 2. Prepare training data: convert them to csv
		// ref: https://cloud.google.com/vertex-ai/docs/tabular-data/classification-regression/prepare-data#csv
		var csv string
		if csv, err = userDataToCSV(allUserData); err != nil {
			return err
		}
		ctx := context.Background()

		// Create the storage client using the service account key file.
		var storageClient *storage.Client
		if storageClient, err = storage.NewClient(ctx, option.WithCredentialsFile(_vaiServiceAccountKey)); err != nil {
			return err
		}
		defer storageClient.Close()

		// GCP bucket name are terrible: they must be globally unique, and they must be DNS compliant
		// ref: https://cloud.google.com/storage/docs/naming-buckets
		// Globally unique: we can use the project id as a prefix
		bucketName := fmt.Sprintf("%s-user-data", _vaiProjectID)
		bucket := storageClient.Bucket(bucketName)
		if _, err = bucket.Attrs(ctx); err != nil {
			// GCP bucket.Attrs returns an error if the bucket does not exist
			// In theory it should be storage.ErrBucketNotExist, but in practice it's a generic error
			// So we try to create the bucket hoping that the error is due to the bucket not existing
			if err = bucket.Create(ctx, _vaiProjectID, nil); err != nil {
				return err
			}
		}

		// Upload an object with storage.Writer.
		// Format date as YYYY-MM-DD
		format := "2006-01-02"
		start := allUserData[0].Date.Format(format)
		end := allUserData[len(allUserData)-1].Date.Format(format)
		// csv on bucket organized in the format: user_id/start_date_end_date.csv
		csvOnBucket := fmt.Sprintf("%d/%s_%s.csv", user.ID, start, end)
		obj := bucket.Object(csvOnBucket)
		if _, err = obj.Attrs(ctx); err == storage.ErrObjectNotExist {
			w := obj.NewWriter(ctx)
			if _, err := w.Write([]byte(csv)); err != nil {
				return err
			}
			if err := w.Close(); err != nil {
				return err
			}
		}

		// 3. Create a dataset from the training data (associate the bucket with the dataset)
		// reference: https://github.com/googleapis/google-cloud-go/issues/6649#issuecomment-1242515615
		var datasetClient *vai.DatasetClient
		vaiEndpoint := fmt.Sprintf("%s-aiplatform.googleapis.com:443", _vaiLocation)
		if datasetClient, err = vai.NewDatasetClient(
			ctx,
			option.WithEndpoint(vaiEndpoint),
			option.WithCredentialsFile(_vaiServiceAccountKey)); err != nil {
			return err
		}
		defer datasetClient.Close()

		//datasetParent := fmt.Sprintf("projects/%s/locations/%s", _vaiProjectID, _vaiLocation)
		//datasetFullPath := fmt.Sprintf("%s/datasets/%s", datasetParent, userDataset)

		// Check if the dataset already exists
		// The dataset name can't be the same as the CSV name (because the CSV name contains slashes)
		// So we use the user id as a prefix
		datasetName := fmt.Sprintf("%d_%s_%s", user.ID, start, end)
		var dataset *vaipb.Dataset
		datasetsIterator := datasetClient.ListDatasets(ctx, &vaipb.ListDatasetsRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", _vaiProjectID, _vaiLocation),
			Filter: fmt.Sprintf(`display_name="%s"`, datasetName),
		})

		if dataset, err = datasetsIterator.Next(); err != nil {
			// Create the dataset
			// ref: https://pkg.go.dev/cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb#CreateDatasetRequest

			// No metadata schema because it's a tabular dataset, and "tabular dataset does not support data import"
			// ref: https://github.com/googleapis/python-aiplatform/blob/6f3b34b39824717e7a995ca1f279230b41491f15/google/cloud/aiplatform/datasets/_datasources.py#LL223C30-L223C75
			// But we need to pass the metadata as a structpb.Value
			// https://github.com/googleapis/python-aiplatform/blob/6f3b34b39824717e7a995ca1f279230b41491f15/google/cloud/aiplatform/datasets/_datasources.py#L48
			// The correct format is: {"input_config": {"gcs_source": {"uri": ["gs://bucket/path/to/file.csv"]}}}
			// Ref the code here: https://cloud.google.com/vertex-ai/docs/samples/aiplatform-create-dataset-tabular-gcs-sample

			csvURI := fmt.Sprintf("gs://%s/%s", bucketName, csvOnBucket)
			var metadata structpb.Struct
			err = metadata.UnmarshalJSON([]byte(fmt.Sprintf(`{"input_config": {"gcs_source": {"uri": ["%s"]}}}`, csvURI)))
			if err != nil {
				return err
			}

			req := &vaipb.CreateDatasetRequest{
				// Required. The resource name of the Location to create the Dataset in.
				// Format: `projects/{project}/locations/{location}`
				Parent: fmt.Sprintf("projects/%s/locations/%s", _vaiProjectID, _vaiLocation),
				Dataset: &vaipb.Dataset{
					DisplayName: datasetName,
					Description: fmt.Sprintf("User %d data", user.ID),
					// No metadata schema because it's a tabular dataset, and "tabular dataset does not support data import"
					// ref: https://github.com/googleapis/python-aiplatform/blob/6f3b34b39824717e7a995ca1f279230b41491f15/google/cloud/aiplatform/datasets/_datasources.py#LL223C30-L223C75
					MetadataSchemaUri: "gs://google-cloud-aiplatform/schema/dataset/metadata/tabular_1.0.0.yaml",
					// But we need to pass the metadata as a structpb.Value
					// https://github.com/googleapis/python-aiplatform/blob/6f3b34b39824717e7a995ca1f279230b41491f15/google/cloud/aiplatform/datasets/_datasources.py#L48
					Metadata: structpb.NewStructValue(&metadata),
				},
			}

			var createDatasetOp *vai.CreateDatasetOperation
			if createDatasetOp, err = datasetClient.CreateDataset(ctx, req); err != nil {
				return err
			}
			if dataset, err = createDatasetOp.Wait(ctx); err != nil {
				return err
			}
		}
		log.Println(dataset.GetName(), dataset.GetDisplayName(), dataset.GetMetadataSchemaUri(), dataset.GetMetadata())
		datasetNameSplit := strings.Split(dataset.GetName(), "/")
		datasetId := datasetNameSplit[len(datasetNameSplit)-1]

		// Associate the bucket with the dataset is required only when the dataset is NOT tabular.
		// ref: https://pkg.go.dev/cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb#ImportDataRequest

		// 4. Export the dataset to a training pipeline
		// ref: https://pkg.go.dev/cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb#ExportDataRequest
		// Perhaps not even export data is required with tabular datasets ?

		/*
			gcsDestination := fmt.Sprintf("gs://%s/%d/", bucketName, user.ID)
			exportDataReq := &vaipb.ExportDataRequest{
				// ref: https://pkg.go.dev/cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb#ExportDataRequest
				// Required. The name of the Dataset resource.
				// Format:
				// `projects/{project}/locations/{location}/datasets/{dataset}`
				// NOTE: the last parameter is the dataset ID and not the dataset display name!
				Name: fmt.Sprintf("projects/%s/locations/%s/datasets/%s", _vaiProjectID, _vaiLocation, datasetId),
				ExportConfig: &vaipb.ExportDataConfig{
					Destination: &vaipb.ExportDataConfig_GcsDestination{
						GcsDestination: &vaipb.GcsDestination{
							OutputUriPrefix: gcsDestination,
						},
					},
					//	Split: &vaipb.ExportDataConfig_FractionSplit{
					//		FractionSplit: &vaipb.ExportFractionSplit{
					//			TrainingFraction:   0.8,
					//			ValidationFraction: 0.1,
					//			TestFraction:       0.1,
					//		},
					//	},
				},
			}

			var op *vai.ExportDataOperation
			if op, err = datasetClient.ExportData(ctx, exportDataReq); err != nil {
				return err
			}
			if _, err = op.Wait(ctx); err != nil {
				if s, ok := status.FromError(err); ok {
					log.Println(s.Message())
					for _, d := range s.Proto().Details {
						log.Println(d)
					}
				}
				return err
			} else {
				log.Println("Export data operation finished")
			}
		*/

		// 5. Create the training pipeline: https://pkg.go.dev/cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb#CreateTrainingPipelineRequest
		// Use the java documentation as reference: // https://cloud.google.com/vertex-ai/docs/samples/aiplatform-create-training-pipeline-tabular-regression-sample

		var modelDisplayName string = "sleep-efficiency-" + strconv.Itoa(int(user.ID))
		var targetColumn string = "SleepEfficiency"

		var pipelineClient *vai.PipelineClient
		if pipelineClient, err = vai.NewPipelineClient(ctx, option.WithEndpoint(vaiEndpoint)); err != nil {
			return err
		}

		var trainingPipeline *vaipb.TrainingPipeline

		// Create the Training Task Inputs
		// Info gathered from the REST API: https://cloud.google.com/vertex-ai/docs/training/automl-api?hl=it#regression
		var trainingTaskInput structpb.Struct
		// reference: https://cloud.google.com/vertex-ai/docs/reference/rpc/google.cloud.aiplatform.v1/schema/trainingjob.definition#automltablesinputs
		err = trainingTaskInput.UnmarshalJSON([]byte(
			fmt.Sprintf(
				`{
					"targetColumn": "%s",
        			"predictionType": "regression",
        			"trainBudgetMilliNodeHours": "1000",
        			"optimizationObjective": "minimize-rmse",
        			"transformations": []
				}`, targetColumn)))
		if err != nil {
			return err
		}
		// TODO: add transformations. The pipeline starts but it fails with the following error:
		// "Transformations cannot be empty."
		// use https://cloud.google.com/vertex-ai/docs/reference/rpc/google.cloud.aiplatform.v1/schema/trainingjob.definition#google.cloud.aiplatform.v1.schema.trainingjob.definition.AutoMlTablesInputs.Transformation

		if trainingPipeline, err = pipelineClient.CreateTrainingPipeline(ctx, &vaipb.CreateTrainingPipelineRequest{
			// Required. The resource name of the Location to create the TrainingPipeline
			// in. Format: `projects/{project}/locations/{location}`
			Parent: fmt.Sprintf("projects/%s/locations/%s", _vaiProjectID, _vaiLocation),
			TrainingPipeline: &vaipb.TrainingPipeline{
				DisplayName:            modelDisplayName,
				TrainingTaskDefinition: "gs://google-cloud-aiplatform/schema/trainingjob/definition/automl_tables_1.0.0.yaml",
				InputDataConfig: &vaipb.InputDataConfig{
					DatasetId: datasetId,
				},
				TrainingTaskInputs: structpb.NewStructValue(&trainingTaskInput),
			},
		}); err != nil {
			if s, ok := status.FromError(err); ok {
				log.Println(s.Message())
				for _, d := range s.Proto().Details {
					log.Println(d)
				}
			}
			return err
		}

		// 6. Get the training pipeline ID and print all the ohter infos
		pipelineID := trainingPipeline.GetName()
		fmt.Println("Training pipeline ID:", pipelineID)
		fmt.Println("Training pipeline display name:", trainingPipeline.GetDisplayName())
		fmt.Println("Training pipeline input data config:", trainingPipeline.GetInputDataConfig())
		fmt.Println("Training pipeline training task inputs:", trainingPipeline.GetTrainingTaskInputs())
		fmt.Println("Training pipeline state:", trainingPipeline.GetState())
		fmt.Println("Training pipeline error:", trainingPipeline.GetError())
		fmt.Println("Training pipeline create time:", trainingPipeline.GetCreateTime())
		fmt.Println("Training pipeline start time:", trainingPipeline.GetStartTime())
		fmt.Println("Training pipeline end time:", trainingPipeline.GetEndTime())

		// TODO: instead of using automl, choose a simple decision tree

		// When creating the training pipeline the Date and ID field must be excluded from the training data
		// ref: https://cloud.google.com/vertex-ai/docs/training/preparing-tabular

		// use PredictionServiceClient and the Explain method to get the explanation of the prediction
		return nil
	}
}
