package models

// Path roles for SamplePath.Role (API contract for clients).
const (
	PathRolePercentile    = "percentile"
	PathRoleMaxDrawdown   = "maxDrawdown"
	PathRoleMaxVolatility = "maxVolatility"
	PathRoleSample        = "sample"
)

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
	// MeanFinalValue and MedianFinalValue are mean/median per-path annualized return (decimal, e.g. 0.07 = 7%/yr).
	MeanFinalValue   float64 `json:"meanFinalValue"`
	MedianFinalValue float64 `json:"medianFinalValue"`
}

// SamplePath is one simulated value path. Role distinguishes percentiles / exemplars vs extra sample draws.
type SamplePath struct {
	Role       string    `json:"role"`
	Percentile float64   `json:"percentile"`
	Label      string    `json:"label"`
	Values     []float64 `json:"values"`
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
