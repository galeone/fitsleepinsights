package types

import (
	"time"

	pgdb "github.com/galeone/fitbit-pgdb/v2"
)

// Predictor is a struct containing the information about the predictors
// created for the user.
// The target column is the target variable the model has been trained to predict.
// Endpoint is the endpoint of the model.
type Predictor struct {
	ID        int64               `igor:"primary_key"`
	User      pgdb.AuthorizedUser `sql:"-"`
	UserID    int64
	CreatedAt time.Time
	Target    string
	Endpoint  string
}

func (Predictor) TableName() string {
	return "predictors"
}
