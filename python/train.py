"""Train a decision tree using the user provided data as CSV, as CLI parameter."""

import argparse
import os
import sys
from glob import glob
from pathlib import Path

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
        required=False,
        # AIP_MODEL_DIR
        # ref: https://cloud.google.com/vertex-ai/docs/reference/rest/v1/CustomJobSpec#FIELDS.base_output_directory
        # When this variable is used, the model uploaded becomes a Vertex AI model
        default=os.environ["AIP_MODEL_DIR"],
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
    bucket = None
    for buck in buckets:
        if buck.name in args.data_location:
            bucket = buck

    if not bucket:
        print(
            f"Unable to find the bucket required by {args.data_location} among the buckets",
            file=sys.stderr,
        )
        return 1

    model = tfdf.keras.CartModel()
    file_name = args.data_location.replace(f"gs://{bucket.name}/", "")
    blob = bucket.blob(file_name)
    with blob.open("r") as file_pointer:
        dataset = pd.read_csv(file_pointer)

    features = dataset.columns
    if args.label not in features:
        print(
            f"Label {args.label} not found among the features of {args.data_location}",
            file=sys.stderr,
        )
        return 1

    potential_labels = {
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
        # default label
        "SleepEfficiency",
    }
    if args.label not in potential_labels:
        print(
            f"\"{args.label}\" not found among the supported labels: {','.join(potential_labels)}",
            file=sys.stderr,
        )
        return 1

    # remove the real label from the potential labels
    potential_labels = potential_labels - {args.label}

    # Remove all the rows with an invalid label (may happen when you don't sleep)
    dataset = dataset[pd.notnull(dataset[args.label])]

    # Remove all the columns with features that are too related sleep (potential labels) or wrong
    # Date: wrong
    # ID: wrong
    dataset = dataset.drop("Date", axis=1)
    dataset = dataset.drop("ID", axis=1)
    for sleep_feature in potential_labels:
        dataset = dataset.drop(sleep_feature, axis=1)

    # Convert to TensorFlow dataset
    tf_dataset = tfdf.keras.pd_dataframe_to_tf_dataset(dataset, label=args.label)

    model.fit(tf_dataset)
    print(model.summary())
    local_model_path = "trained_model"  # NOTE: do not name it saved_model.pb or model, otherwise replaces will fail
    model.save(local_model_path)

    model_destination_folder = args.model_destination.replace(
        f"gs://{bucket.name}/", ""
    )

    files = glob(f"{local_model_path}/**", recursive=True)
    for file in files:
        if Path(file).is_file():
            # directly upload the model files (without the folder) inside model destination folder
            dest = Path(model_destination_folder) / Path(
                file.replace(f"{local_model_path}/", "")
            )
            blob = bucket.blob(dest.as_posix())

            blob.upload_from_filename(file)

    return 0


if __name__ == "__main__":
    sys.exit(main())
