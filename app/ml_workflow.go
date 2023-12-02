// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	vai "cloud.google.com/go/aiplatform/apiv1beta1"
	vaipb "cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb"
	fitbit_pgdb "github.com/galeone/fitbit-pgdb/v2"
	"github.com/galeone/fitsleepinsights/database/types"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	storage "cloud.google.com/go/storage"
)

func TrainAndDeployPredictor(user *fitbit_pgdb.AuthorizedUser, targetColumn string) (err error) {

	var fetcher *fetcher
	if fetcher, err = NewFetcher(user); err != nil {
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
		if err = bucket.Create(ctx, _vaiProjectID, &storage.BucketAttrs{
			Location: _vaiLocation, // Important to have all the resources in the same location
			Name:     bucketName,
		}); err != nil {
			return err
		}
	}

	// Upload an object with storage.Writer.
	// Format date as YYYY-MM-DD
	format := "2006-01-02"
	start := allUserData[len(allUserData)-1].Date.Format(format)
	end := allUserData[0].Date.Format(format)
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

	// 3. Create a custom training job with a custom container
	// ref: https://cloud.google.com/vertex-ai/docs/training/create-custom-job#create_custom_job-java
	modelName := fmt.Sprintf("%s-predictor-%d", targetColumn, user.ID)
	var modelClient *vai.ModelClient
	if modelClient, err = vai.NewModelClient(ctx, option.WithEndpoint(_vaiEndpoint)); err != nil {
		return err
	}
	defer modelClient.Close()

	imageURI := fmt.Sprintf("%s-docker.pkg.dev/%s/tfdf/cart:0.0.1", _vaiLocation, _vaiProjectID)

	var customJobClient *vai.JobClient
	if customJobClient, err = vai.NewJobClient(ctx, option.WithEndpoint(_vaiEndpoint)); err != nil {
		return err
	}
	defer customJobClient.Close()

	req := &vaipb.CreateCustomJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", _vaiProjectID, _vaiLocation),
		CustomJob: &vaipb.CustomJob{
			DisplayName: fmt.Sprintf("%s-%d", targetColumn, user.ID),
			JobSpec: &vaipb.CustomJobSpec{
				BaseOutputDirectory: &vaipb.GcsDestination{
					OutputUriPrefix: fmt.Sprintf("gs://%s/%d/", bucketName, user.ID),
				},
				WorkerPoolSpecs: []*vaipb.WorkerPoolSpec{
					{
						Task: &vaipb.WorkerPoolSpec_ContainerSpec{
							ContainerSpec: &vaipb.ContainerSpec{
								ImageUri: imageURI,
								Args: []string{
									"--data-location",
									fmt.Sprintf("gs://%s/%s", bucketName, csvOnBucket),
									//"--model-destination",
									//fmt.Sprintf("gs://%s/%d/", bucketName, user.ID),
									"--label",
									targetColumn,
								},
								Env: []*vaipb.EnvVar{
									{
										Name:  "CLOUD_ML_PROJECT_ID",
										Value: _vaiProjectID,
									},
								},
							},
						},
						ReplicaCount: 1,
						MachineSpec: &vaipb.MachineSpec{
							MachineType:      "n1-standard-4",
							AcceleratorCount: 0,
						},
						DiskSpec: &vaipb.DiskSpec{
							BootDiskType:   "pd-ssd",
							BootDiskSizeGb: 100,
						},
					},
				},
			},
		},
	}

	var resp *vaipb.CustomJob
	if resp, err = customJobClient.CreateCustomJob(ctx, req); err != nil {
		return err
	}

	customJobName := resp.GetName()

	// Wait for the job to finish
	for status := resp.GetState(); status != vaipb.JobState_JOB_STATE_SUCCEEDED &&
		status != vaipb.JobState_JOB_STATE_FAILED && status != vaipb.JobState_JOB_STATE_CANCELLED; status = resp.GetState() {

		if resp, err = customJobClient.GetCustomJob(ctx, &vaipb.GetCustomJobRequest{
			Name: customJobName,
		}); err != nil {
			return err
		}

		log.Println(resp.GetState())
		// sleep 1 second
		time.Sleep(1 * time.Second)
	}

	// Upload the model to the model registry
	// ref: https://cloud.google.com/vertex-ai/docs/model-registry/import-model#custom-container
	var uploadOp *vai.UploadModelOperation
	if uploadOp, err = modelClient.UploadModel(ctx, &vaipb.UploadModelRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", _vaiProjectID, _vaiLocation),
		Model: &vaipb.Model{
			Name:        modelName,
			DisplayName: modelName,
			// MetadataSchemaUri: "gs://google-cloud-aiplatform/schema/trainingjob/definition/custom_task_1.0.0.yaml",
			ContainerSpec: &vaipb.ModelContainerSpec{
				// use a prebuilt container, so we can create a shared pool of resources later
				ImageUri: "europe-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-12:latest",
			},
			ArtifactUri: fmt.Sprintf("gs://%s/%d/model", bucketName, user.ID),
		},
	}); err != nil {
		return err
	}

	var uploadModelResponse *vaipb.UploadModelResponse
	if uploadModelResponse, err = uploadOp.Wait(ctx); err != nil {
		return err
	}
	log.Println(uploadModelResponse.GetModel())

	var endpointClient *vai.EndpointClient
	if endpointClient, err = vai.NewEndpointClient(ctx, option.WithEndpoint(_vaiEndpoint)); err != nil {
		return err
	}
	defer endpointClient.Close()

	var createEndpointOp *vai.CreateEndpointOperation
	if createEndpointOp, err = endpointClient.CreateEndpoint(ctx, &vaipb.CreateEndpointRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", _vaiProjectID, _vaiLocation),
		Endpoint: &vaipb.Endpoint{
			Name:        modelName,
			DisplayName: modelName,
		},
	}); err != nil {
		return err
	}

	// After creating the endpoint we can get a meaningful name
	// like projects/1064343834149/locations/europe-west6/endpoints/6066638969137790976
	// but it doesn't contain the display name or the name we choose, so it's unclear how to get
	// this information back
	var endpoint *vaipb.Endpoint
	if endpoint, err = createEndpointOp.Wait(ctx); err != nil {
		return err
	}

	log.Println("endpoint name:", endpoint.GetName())

	var resourcePoolClient *vai.DeploymentResourcePoolClient
	if resourcePoolClient, err = vai.NewDeploymentResourcePoolClient(ctx, option.WithEndpoint(_vaiEndpoint)); err != nil {
		return err
	}
	defer resourcePoolClient.Close()

	deploymentResourcePoolId := "resource-pool"
	var deploymentResourcePool *vaipb.DeploymentResourcePool = nil
	iter := resourcePoolClient.ListDeploymentResourcePools(ctx, &vaipb.ListDeploymentResourcePoolsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", _vaiProjectID, _vaiLocation),
	})
	var item *vaipb.DeploymentResourcePool
	for item, _ = iter.Next(); err == nil; item, err = iter.Next() {
		log.Println(item.GetName())
		if strings.Contains(item.GetName(), deploymentResourcePoolId) {
			deploymentResourcePool = item
			log.Println("Found deployment resource pool: ", deploymentResourcePool.GetName())
			break
		}
	}

	if deploymentResourcePool == nil {
		log.Println("Creating a new deployment resource pool")
		// Create a deployment resource pool: FOR SHARED RESOURCES ONLY
		var createDeploymentResourcePoolOp *vai.CreateDeploymentResourcePoolOperation
		if createDeploymentResourcePoolOp, err = resourcePoolClient.CreateDeploymentResourcePool(ctx, &vaipb.CreateDeploymentResourcePoolRequest{
			Parent:                   fmt.Sprintf("projects/%s/locations/%s", _vaiProjectID, _vaiLocation),
			DeploymentResourcePoolId: deploymentResourcePoolId,
			DeploymentResourcePool: &vaipb.DeploymentResourcePool{
				DedicatedResources: &vaipb.DedicatedResources{
					MachineSpec: &vaipb.MachineSpec{
						MachineType:      "n1-standard-4",
						AcceleratorCount: 0,
					},
					MinReplicaCount: 1,
					MaxReplicaCount: 1,
				},
			},
		}); err != nil {
			return err
		}

		if deploymentResourcePool, err = createDeploymentResourcePoolOp.Wait(ctx); err != nil {
			return err
		}
		log.Println("Created resource pool: ", deploymentResourcePool.GetName())
	}

	// Shared doesn't work with custom containers
	var deployModelOp *vai.DeployModelOperation
	if deployModelOp, err = endpointClient.DeployModel(ctx, &vaipb.DeployModelRequest{
		Endpoint: endpoint.GetName(),
		DeployedModel: &vaipb.DeployedModel{
			DisplayName: modelName,
			Model:       uploadModelResponse.GetModel(),
			//EnableContainerLogging: true, // enable logging only for custom containers
			PredictionResources: &vaipb.DeployedModel_SharedResources{
				SharedResources: deploymentResourcePool.GetName(),
			},
		},
	}); err != nil {
		return err
	}

	if _, err = deployModelOp.Wait(ctx); err != nil {
		return err
	}

	return _db.Create(&types.Predictor{
		UserID:   user.ID,
		Target:   targetColumn,
		Endpoint: endpoint.GetName(),
	})
}

func PredictSleepEfficiency(user *fitbit_pgdb.AuthorizedUser, userData []*UserData) (uint8, error) {
	var err error
	ctx := context.Background()

	var predictionClient *vai.PredictionClient
	if predictionClient, err = vai.NewPredictionClient(ctx, option.WithEndpoint(_vaiEndpoint)); err != nil {
		return 0, err
	}
	defer predictionClient.Close()
	var instances []*structpb.Value
	var toSkip []string = []string{
		// potential labels excluded during training
		"MinutesAfterWakeup",
		"MinutesAsleep",
		"MinutesAwake",
		"MinutesToFallAsleep",
		"TimeInBed",
		"LightSleepMinutes",
		"LightSleepCount",
		"DeepSleepMinutes",
		"DeepSleepCount",
		"RemSleepMinutes",
		"RemSleepCount",
		"WakeSleepMinutes",
		"WakeSleepCount",
		"SleepDuration",
		"SleepEfficiency",
		// ID and Date are not required
		"ID",
		"Date",
	}

	if instances, err = UserDataToPredictionInstance(userData, toSkip); err != nil {
		return 0, err
	}

	// Get the predictor
	var predictor types.Predictor
	predictor.UserID = user.ID
	predictor.Target = "SleepEfficiency"
	if err = _db.Model(types.Predictor{}).Where(predictor).Scan(&predictor); err != nil {
		return 0, err
	}

	var predictResponse *vaipb.PredictResponse
	if predictResponse, err = predictionClient.Predict(ctx, &vaipb.PredictRequest{
		Endpoint:  predictor.Endpoint,
		Instances: instances,
	}); err != nil {
		return 0, err
	}

	predictionsBatch := predictResponse.GetPredictions()

	if len(predictionsBatch) == 0 {
		return 0, fmt.Errorf("no predictions")
	}

	// Get the argmax
	var max float64 = 0
	var maxIndex int = 0
	values := predictionsBatch[0].GetListValue().GetValues()
	for i, value := range values {
		if value.GetNumberValue() > max {
			max = value.GetNumberValue()
			maxIndex = i
		}
	}

	return uint8(maxIndex), nil
}
