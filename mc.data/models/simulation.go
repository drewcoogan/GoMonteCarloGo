package models

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
