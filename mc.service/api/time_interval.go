package api

// TimeInterval specifies a frequency to query for intraday stock data.
type TimeInterval uint8

const (
	TimeIntervalOneMinute TimeInterval = iota
	TimeIntervalFiveMinute
	TimeIntervalFifteenMinute
	TimeIntervalThirtyMinute
	TimeIntervalSixtyMinute
	TimeIntervalDaily
	TimeIntervalWeekly
	TimeIntervalMonthly
)

func (t TimeInterval) Name() string {
	switch t {
	case TimeIntervalOneMinute:
		return "TimeIntervalOneMinute"
	case TimeIntervalFiveMinute:
		return "TimeIntervalFiveMinute"
	case TimeIntervalFifteenMinute:
		return "TimeIntervalFifteenMinute"
	case TimeIntervalThirtyMinute:
		return "TimeIntervalThirtyMinute"
	case TimeIntervalSixtyMinute:
		return "TimeIntervalSixtyMinute"
	case TimeIntervalDaily:
		return "TimeIntervalDaily"
	case TimeIntervalWeekly:
		return "TimeIntervalWeekly"
	case TimeIntervalMonthly:
		return "TimeIntervalMonthly"
	default:
		return ""
	}
}

func (t TimeInterval) Interval() string {
	switch t {
	case TimeIntervalOneMinute:
		return "1min"
	case TimeIntervalFiveMinute:
		return "5min"
	case TimeIntervalFifteenMinute:
		return "15min"
	case TimeIntervalThirtyMinute:
		return "30min"
	case TimeIntervalSixtyMinute:
		return "60min"
	case TimeIntervalDaily:
		return "DAILY"
	case TimeIntervalWeekly:
		return "WEEKLY"
	case TimeIntervalMonthly:
		return "MONTHLY"
	default:
		return ""
	}
}
