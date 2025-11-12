package alpha_vantage

import (
	"strings"
)

type TimeSeries uint8

// TimeSeries specifies a frequency to query for stock data.
const (
	TimeSeriesDaily TimeSeries = iota
	TimeSeriesDailyAdjusted
	TimeSeriesWeekly
	TimeSeriesWeeklyAdjusted
	TimeSeriesMonthly
	TimeSeriesMonthlyAdjusted
)

func (t TimeSeries) Name() string {
	switch t {
	case TimeSeriesDaily:
		return "TimeSeriesDaily"
	case TimeSeriesDailyAdjusted:
		return "TimeSeriesDailyAdjusted"
	case TimeSeriesWeekly:
		return "TimeSeriesWeekly"
	case TimeSeriesWeeklyAdjusted:
		return "TimeSeriesWeeklyAdjusted"
	case TimeSeriesMonthly:
		return "TimeSeriesMonthly"
	case TimeSeriesMonthlyAdjusted:
		return "TimeSeriesMonthlyAdjusted"
	default:
		return ""
	}
}

func (t TimeSeries) Function() string {
	switch t {
	case TimeSeriesDaily:
		return "TIME_SERIES_DAILY"
	case TimeSeriesDailyAdjusted:
		return "TIME_SERIES_DAILY_ADJUSTED"
	case TimeSeriesWeekly:
		return "TIME_SERIES_WEEKLY"
	case TimeSeriesWeeklyAdjusted:
		return "TIME_SERIES_WEEKLY_ADJUSTED"
	case TimeSeriesMonthly:
		return "TIME_SERIES_MONTHLY"
	case TimeSeriesMonthlyAdjusted:
		return "TIME_SERIES_MONTHLY_ADJUSTED"
	default:
		return ""
	}
}

func (t TimeSeries) TimeSeriesKey() string {
	switch t {
	case TimeSeriesDaily:
		return "Daily Time Series"
	case TimeSeriesDailyAdjusted:
		return "Daily Adjusted Time Series"
	case TimeSeriesWeekly:
		return "Weekly Time Series"
	case TimeSeriesWeeklyAdjusted:
		return "Weekly Adjusted Time Series"
	case TimeSeriesMonthly:
		return "Monthly Time Series"
	case TimeSeriesMonthlyAdjusted:
		return "Monthly Adjusted Time Series"
	default:
		return ""
	}
}

func (t TimeSeries) IsAdjusted() bool {
	return strings.HasSuffix(t.Function(), "_ADJUSTED")
}