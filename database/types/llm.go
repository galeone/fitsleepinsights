package types

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

type Report struct {
	ID         int64 `igor:"primary_key"`
	UserID     int64
	StartDate  time.Time
	EndDate    time.Time
	ReportType string
	Report     string
	Embedding  pgvector.Vector
}

func (r *Report) TableName() string {
	return "reports"
}
