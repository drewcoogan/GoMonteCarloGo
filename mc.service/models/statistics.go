package models

const (
	StandardNormal = iota
	StudentT
)

const (
	Daily     = 252
	Weekly    = 52
	Monthly   = 12
	Quarterly = 4
	Yearly    = 1
)

func ConvertFrequencyToString(inp int) string {
	switch inp {
	case Daily:
		return "days"
	case Weekly:
		return "weeks"
	case Monthly:
		return "months"
	case Quarterly:
		return "quarters"
	case Yearly:
		return "years"
	default:
		return ""
	}
}
