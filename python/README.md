Steps

1. Write the `train.py` script
2. Write the `Dockerfile`
3. Authenticate, build and push

```
gcloud auth configure-docker europe-west6-docker.pkg.dev
docker buildx build -f Dockerfile -t europe-west6-docker.pkg.dev/train-and-deploy-experiment/tfdf/cart:0.0.1 .
docker push europe-west6-docker.pkg.dev/train-and-deploy-experiment/tfdf/cart:0.0.1
```

TODO: create in GO a custom Job. Example on how to do it in Java: https://cloud.google.com/vertex-ai/docs/training/create-custom-job#create_custom_job-java

Local training:

```
GOOGLE_APPLICATION_CREDENTIALS=../train-and-deploy-experiment-497d2e4f4272.json CLOUD_ML_PROJECT_ID=train-and-deploy-experiment python train.py --data-location gs://train-and-deploy-experiment-user-data/1/2023-07-24_2023-03-02.csv --model-destination gs://train-and-deploy-experiment-user-data/1/ --label SleepEfficiency
```
