package models

import "time"

type TimeSeriesReturn struct {
	Id        int32     `db:"source_id"`
	Timestamp time.Time `db:"timestamp"`
	LogReturn float64   `db:"log_return"`
}
