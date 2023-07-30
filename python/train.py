"""Train a decision tree using the user provided data as CSV, as CLI parameter."""

import argparse
import os
import sys

import pandas as pd
import tensorflow_decision_forests as tfdf
from google.cloud import storage

# https://cloud.google.com/vertex-ai/docs/training/containers-overview


def parse_args():
    """Parse the CLI arguments. Mandatory:
    - data-location: the full path of the csv
    - model-destination: the full path where to store the trained model
    """
    parser = argparse.ArgumentParser(description="Train a decision tree")
    parser.add_argument(
        "--data-location",
        help="The fullpath over GCP where to find the training data to use",
        required=True,
    )
    parser.add_argument(
        "--model-destination",
        help="The folder on GCP where to store the trained model",
        required=True,
    )
    parser.add_argument("--label", help="The target variable to predict", required=True)
    return parser.parse_args()


def main():
    """Application entrypoint."""

    print(f"TensorFlow Decision Forests version: {tfdf.__version__}")
    args = parse_args()

    project_id = os.environ["CLOUD_ML_PROJECT_ID"]
    storage_client = storage.Client(project=project_id)

    buckets = storage_client.list_buckets()
    print("Buckets:")
    bucket = None
    for buck in buckets:
        print(buck.name)
        if buck.name in args.data_location:
            bucket = buck

    if not bucket:
        print(
            f"Unable to find the bucket required by {args.data_location} among the buckets",
            file=sys.stderr,
        )
        return 1

    model = tfdf.keras.CartModel()
    file_name = args.data_location.replace(bucket.name, "")
    blob = bucket.blob(file_name)
    with blob.open("r") as file_pointer:
        dataset = pd.read_csv(file_pointer)
    tf_dataset = tfdf.keras.pd_dataframe_to_tf_dataset(dataset, label=args.label)

    model.fit(tf_dataset)
    print(model.summary())
    local_model_path = "model.pb"
    model.save(local_model_path)

    blob = bucket.blob(args.model_destination)
    with open(local_model_path, "rb") as file_pointer:
        local_model = file_pointer.readall()
    with blob.open("wb") as file_pointer:
        file_pointer.write(local_model)

    return 0


if __name__ == "__main__":
    sys.exit(main())
