package app

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
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
