package models

import (
	"time"
)

type TimeSeriesResult struct {
	Metadata   *TimeSeriesMetadata
	TimeSeries []*TimeSeriesData
}

type TimeSeriesIntradayResult struct {
	Metadata   *TimeSeriesMetadata
	TimeSeries []*TimeSeriesIntradayData
}

type TimeSeriesMetadata struct {
	Id            int32     `db:"id"`
	Symbol        string    `db:"symbol"`
	LastRefreshed time.Time `db:"last_refreshed"`
}

type TimeSeriesData struct {
	SourceId  int32     `db:"source_id"`
	Timestamp time.Time `db:"timestamp"`
	TimeSeriesOHLCV
	AdjustedClose  float64 `db:"adjusted_close"`
	DividendAmount float64 `db:"dividend_amount"`
}

type TimeSeriesIntradayData struct {
	SourceId  int32     `db:"source_id"`
	Timestamp time.Time `db:"timestamp"`
	TimeSeriesOHLCV
}

type TimeSeriesOHLCV struct {
	Open   float64 `db:"open"`
	High   float64 `db:"high"`
	Low    float64 `db:"low"`
	Close  float64 `db:"close"`
	Volume float64 `db:"volume"`
}
