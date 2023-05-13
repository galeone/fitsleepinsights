package app

import (
	"bytes"
	"encoding/csv"
	"errors"
)

// userDataToCSV converts the slice of UserData to a CSV string
func userDataToCSV(userData []*UserData) (ret string, err error) {
	if len(userData) == 0 {
		return ret, errors.New("empty userData slice")
	}

	headers := userData[0].Headers()

	buffer := bytes.NewBufferString("")
	w := csv.NewWriter(buffer)
	w.Write(headers)

	for _, u := range userData {
		if err := w.Write(u.Values()); err != nil {
			return ret, err
		}
	}

	return buffer.String(), nil
}
