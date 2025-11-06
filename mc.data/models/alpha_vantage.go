package models

import (
	"time"

	"github.com/guregu/null/v6"
)

type TimeSeriesResult struct {
	MetaData *TimeSeriesMetaData
	TimeSeries []*TimeSeriesData
}

type TimeSeriesMetaData struct {
	Information   null.String
	Symbol        null.String
	LastRefreshed time.Time 
	Interval      null.String
	OutputSize    null.String
	TimeZone      null.String
}

type TimeSeriesData struct {
	Timestamp      time.Time
	Open           null.Float
	High           null.Float
	Low            null.Float
	Close          null.Float
	AdjustedClose  null.Float
	Volume         null.Float
	DividendAmount null.Float
}