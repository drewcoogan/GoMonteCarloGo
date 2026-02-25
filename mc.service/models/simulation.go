package models

import "time"

// SimulationSettingsResources will be the resources for the simulation settings, rest will be simple numbers provided by user
type SimulationSettingsResources struct {
	DistType             map[string]int `json:"disttype"`             // standar normal, student t
	SimulationUnitOfTime map[string]int `json:"simulationunitoftime"` // daily, weekly, monthly, quarterly, yearly
	SimulationDuration   map[string]int `json:"simulationduration"`   // number of units of time to simulate
}

// GetSimulationSettingsResources will return the simulation settings resources.
// This approach makes sure everything is mapped correctly so uses a shared resource
func GetSimulationSettingsResources() SimulationSettingsResources {
	distType := map[string]int{
		"standardNormal": StandardNormal,
		"studentT":       StudentT,
	}

	// we just have weekly for now, can test and expand later
	simulationUnitOfTime := map[string]int{
		"weekly": Weekly,
	}

	simulationDuration := map[string]int{
		"days":     Daily,
		"weeks":    Weekly,
		"months":   Monthly,
		"quarters": Quarterly,
		"years":    Yearly,
	}

	return SimulationSettingsResources{
		DistType:             distType,
		SimulationUnitOfTime: simulationUnitOfTime,
		SimulationDuration:   simulationDuration,
	}
}

// DistTypeToString returns the string name for storage given the dist type code
func DistTypeToString(code int) string {
	switch code {
	case StandardNormal:
		return "standardNormal"
	case StudentT:
		return "studentT"
	default:
		return ""
	}
}

// SimulationUnitOfTimeToString returns the string name for storage given the unit code
func SimulationUnitOfTimeToString(code int) string {
	switch code {
	case Weekly:
		return "weekly"
	case Daily:
		return "days"
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

// SimulationRequestSettings will be the request from the front end to the simulation controller
type SimulationRequestSettings struct {
	DistType             int `json:"disttype"`             // standar normal, student t
	SimulationUnitOfTime int `json:"simulationunitoftime"` // daily, weekly, monthly, quarterly, yearly
	SimulationDuration   int `json:"simulationduration"`   // number of units of time to simulate

	MaxLookback time.Duration `json:"maxlookback"`
	Iterations  int           `json:"iterations"`
	Seed        int64         `json:"seed"`

	DegreesOfFreedom int `json:"degreesoffreedom"` // degrees of freedom for student t distribution
}

// SimulationResponse will be the response from the simulation controller and what is sent to the front end
type SimulationResponse struct {
	RiskMetrics SimulationRiskMetrics `json:"riskMetrics"`
	SamplePaths []SamplePath          `json:"samplePaths"`
	Summary     SimulationStats       `json:"simulationStats"`
}

// ScarioRunRiskMetrics will be numbers on the page when looking at scenario results
type SimulationRiskMetrics struct {
	VaR95             float64 `json:"var95"`
	VaR99             float64 `json:"var99"`
	CVaR95            float64 `json:"cvar95"`
	CVaR99            float64 `json:"cvar99"`
	ProbabilityOfLoss float64 `json:"probabilityOfLoss"`
	MaxDrawdownP95    float64 `json:"maxDrawdownP95"`
	MeanFinalValue    float64 `json:"meanFinalValue"`
	MedianFinalValue  float64 `json:"medianFinalValue"`
}

// SamplePath will show the user a few of the paths the portfolio took
type SamplePath struct {
	Percentile float64   `json:"percentile"`
	Values     []float64 `json:"values"`
	Label      string    `json:"label"`
}

// ScenarioStats will show the user bands for the timeseries of value
type SimulationStats struct {
	Mean   []float64 `json:"mean"`
	StdDev []float64 `json:"stdDev"`
	P5     []float64 `json:"p5"`
	P25    []float64 `json:"p25"`
	P50    []float64 `json:"p50"`
	P75    []float64 `json:"p75"`
	P95    []float64 `json:"p95"`
}
