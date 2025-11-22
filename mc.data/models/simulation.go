package models

import "time"

type TimeSeriesReturn struct {
	Id        int32     `db:"source_id"`
	Timestamp time.Time `db:"timestamp"`
	LogReturn float64   `db:"log_return"`
}

type TickerReturns struct {
	SourceID   int32
	Ticker     string
	Returns    []float64
	Dates      []time.Time
	MeanReturn float64
	StdDev     float64
}
