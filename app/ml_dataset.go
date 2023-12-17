package app

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"google.golang.org/protobuf/types/known/structpb"
)

// csvHeaders returns the headers for the CSV file
func csvHeaders(userData []*UserData) []string {
	if len(userData) == 0 {
		return []string{}
	}

	headers := []string{"ID"}
	headers = append(headers, userData[0].Headers()...)
	return headers
}

// userDataToCSV converts the slice of UserData to a CSV string
// Add a column ID, because vertex.ai requires it. It's not used for training
// The label column will be decided by the user ideally. Right now we use the sleep efficiency
// The date is already in the csv in the vertex.ai expected format, so we don't need to add it.
// ref: https://cloud.google.com/vertex-ai/docs/tabular-data/bp-tabular
func userDataToCSV(userData []*UserData) (ret string, err error) {
	if len(userData) == 0 {
		return ret, errors.New("empty userData slice")
	}

	headers := csvHeaders(userData)

	buffer := bytes.NewBufferString("")
	w := csv.NewWriter(buffer)
	w.Write(headers)

	for id, u := range userData {
		if err := w.Write(append([]string{strconv.FormatInt(int64(id), 10)}, u.Values()...)); err != nil {
			return ret, err
		}
	}

	return buffer.String(), nil
}

// UNUSED UserDataToPredictionInstance converts a slice of UserData to a slice of structpb.Value.
// It skips all the columns that are not used for training, that you should pass in the skipColumns parameter.
func userDataToPredictionInstanceReflection(userData []*UserData, skipColumns []string) ([]*structpb.Value, error) {
	if len(userData) == 0 {
		return nil, errors.New("empty userData slice")
	}

	var instances []*structpb.Value = make([]*structpb.Value, len(userData))

	idsToSkip := make(map[int]bool)
	indirect := reflect.Indirect(reflect.ValueOf(userData[0]))
	indirectType := indirect.Type()
	totFields := indirectType.NumField()
	for i := 0; i < totFields; i++ {
		field := indirectType.Field(i).Name
		for _, skipColumn := range skipColumns {
			if field == skipColumn {
				idsToSkip[i] = true
				break
			}
		}
	}

	for i, u := range userData {
		rawInstance := map[string]interface{}{}
		for j, v := range u.Values() {
			if _, ok := idsToSkip[j]; ok {
				continue
			}
			rawInstance[indirectType.Field(j).Name] = v
		}
		var err error
		instances[i], err = structpb.NewValue(rawInstance)
		if err != nil {
			return nil, err
		}
	}

	return instances, nil
}

// UserDataToPredictionInstance converts a slice of UserData to a slice of structpb.Value.
// It skips all the columns that are not used for training, that you should pass in the skipColumns parameter.
func UserDataToPredictionInstance(userData []*UserData, skipColumns []string) ([]*structpb.Value, error) {
	if len(userData) == 0 {
		return nil, errors.New("empty userData slice")
	}

	var instances []*structpb.Value = make([]*structpb.Value, len(userData))

	idsToSkip := make(map[int]bool)
	columns := userData[0].Headers()
	for i, column := range columns {
		for _, skipColumn := range skipColumns {
			if column == skipColumn {
				idsToSkip[i] = true
				break
			}
		}
	}

	for i, u := range userData {
		rawInstance := map[string]interface{}{}
		for j, v := range u.Values() {
			if _, ok := idsToSkip[j]; ok {
				continue
			}
			// If the value is empty, is a missing value and we consider it as a float 0.
			if v == "" {
				// In theory tree-based models should be able to handle missing values, but in practice they don't.
				// This is a limitation of the deployment of tfdf models in Vertex AI, I guess.

				// Column "ActivitiesNameConcatenation" is a string column, so we can't set it to 0.
				// We set it to an empty string.
				if columns[j] == "ActivitiesNameConcatenation" {
					rawInstance[columns[j]] = ""
				} else {
					rawInstance[columns[j]] = 0.0
				}
				continue
			}
			// If the value can be converted to float, we convert it.
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				rawInstance[columns[j]] = f
				continue
			}
			// Otherwise is a string.
			rawInstance[columns[j]] = v
		}
		var err error
		instances[i], err = structpb.NewValue(rawInstance)
		if err != nil {
			return nil, err
		}
	}

	fmt.Println(instances)
	return instances, nil
}
