package models

import (
	"time"

	"github.com/guregu/null/v6"
)

type TimeSeriesResult struct{
	Metadata *TimeSeriesMetadata
	TimeSeries []*TimeSeriesData
}

// so these are made to handle both daily and intraday, i was going to try and be cute and make one return objects, but its going to more pain than its worth
// i think i can keep majority of av.go, as its using reflection for just mapping, but we need the objects retuned from pgsql to match exactly what the object is
type TimeSeriesMetadata struct{
	Information   null.String
	Symbol        string
	LastRefreshed time.Time 
	Interval      null.String
	OutputSize    null.String
	TimeZone      string
}

type TimeSeriesData struct{
	Timestamp      time.Time
	Open           null.Float
	High           null.Float
	Low            null.Float
	Close          null.Float
	AdjustedClose  null.Float
	Volume         null.Float
	DividendAmount null.Float
}

type TimeSeriesIntradayResult struct{
	
}

type TimeSeriesIntradayMetadata struct{
	Symbol        string
	LastRefreshed time.Time // use full date time in db
	Interval      string
}

type TimeSeriesIntradayData struct{
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}