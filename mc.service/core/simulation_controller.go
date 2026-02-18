package core

import (
	"fmt"
	"math"
	"slices"

	"gonum.org/v1/gonum/stat"
	m "mc.data/models"
)

type ScenarioRunResponse struct {
	RiskMetrics ScenarioRunRiskMetrics `json:"riskMetrics"`
	SamplePaths []SamplePath           `json:"samplePaths"`
	Summary     ScenarioStats          `json:"scenarioStats"`
}

// ScarioRunRiskMetrics will be numbers on the page when looking at scenario results
type ScenarioRunRiskMetrics struct {
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
type ScenarioStats struct {
	Mean   []float64 `json:"mean"`
	StdDev []float64 `json:"stdDev"`
	P5     []float64 `json:"p5"`
	P25    []float64 `json:"p25"`
	P50    []float64 `json:"p50"`
	P75    []float64 `json:"p75"`
	P95    []float64 `json:"p95"`
}

// TODO: this is where we can add a queue to only run one scenario at a time
// can probably send and manage the queue after its validated and scenario is good
func (sc *ServiceContext) RunScenario(scenarioID int32, settings SimulationSettings) (*ScenarioRunResponse, error) {
	scenario, err := sc.PostgresConnection.GetScenarioByID(sc.Context, scenarioID)
	if err != nil {
		return nil, err
	}

	scenarioRunId, err := sc.PostgresConnection.InsertScenarioRunHistory(sc.Context, scenario.Id)
	if err != nil {
		return nil, err
	}

	if err := validateScenario(scenario); err != nil {
		return sc.markScenarioRunAsFailure(scenarioRunId, err.Error())
	}

	res, err := sc.RunMonteCarloSimulation(scenario, settings)
	if err != nil {
		return sc.markScenarioRunAsFailure(scenarioRunId, err.Error())
	}

	if err := sc.PostgresConnection.UpdateScenarioRunAsSuccess(sc.Context, scenarioRunId); err != nil {
		return nil, err // TODO: should we mark as failure here? do we care?
	}

	response := buildScenarioResponse(res)

	return response, nil
}

func validateScenario(scenario *m.Scenario) error {
	// make sure the total weight is 100%
	weightSum := 0.0
	for _, w := range scenario.Components {
		weightSum += w.Weight
	}

	if math.Abs(weightSum-1.0) > 1e-6 {
		return fmt.Errorf("weights must sum to 1.0, got %.6f", weightSum)
	}

	// make sure assets allocated to are unique
	v := make(map[int32]bool, len(scenario.Components))
	for _, a := range scenario.Components {
		if _, ok := v[a.AssetId]; !ok {
			return fmt.Errorf("duplicate assetId %d", a.AssetId)
		}
		v[a.AssetId] = true
	}

	return nil
}

func (sc *ServiceContext) markScenarioRunAsFailure(runId int32, errorMessage string) (*ScenarioRunResponse, error) {
	return nil, sc.PostgresConnection.UpdateScenarioRunAsFailure(sc.Context, runId, errorMessage)
}

func buildScenarioResponse(results []*SimulationResult) *ScenarioRunResponse {
	// sort once by final value (ascending). All quintile calculations use this order,
	// most of the rest dont care about order, so this is fine
	slices.SortFunc(results, func(a, b *SimulationResult) int {
		if a.FinalValue < b.FinalValue {
			return -1
		}
		if a.FinalValue > b.FinalValue {
			return 1
		}
		return 0
	})

	riskMetrics := calculateRiskMetrics(results)
	samplePaths := selectSamplePaths(results)
	summary := calculateSummaryStats(results)

	return &ScenarioRunResponse{
		RiskMetrics: riskMetrics,
		SamplePaths: samplePaths,
		Summary:     summary,
	}
}

func calculateRiskMetrics(results []*SimulationResult) ScenarioRunRiskMetrics {
	n := len(results)

	finalValues := make([]float64, n)
	totalReturns := make([]float64, n)
	maxDrawdowns := make([]float64, n)

	for i, res := range results {
		finalValues[i] = res.FinalValue
		totalReturns[i] = res.TotalReturn
		maxDrawdowns[i] = res.MaxDrawdown
	}

	var95 := stat.Quantile(0.05, stat.Empirical, totalReturns, nil)
	var99 := stat.Quantile(0.01, stat.Empirical, totalReturns, nil)
	cvar95 := calculateCVaR(totalReturns, 0.05)
	cvar99 := calculateCVaR(totalReturns, 0.01)

	lossCount := 0
	for _, r := range totalReturns {
		if r < 0 {
			lossCount++
		}
	}
	probabilityOfLoss := float64(lossCount) / float64(n)

	// maxDrawdowns needs to be sorted as its not proportioal to final value
	slices.Sort(maxDrawdowns)
	maxDrawdownP95 := stat.Quantile(0.95, stat.Empirical, maxDrawdowns, nil)

	meanFinal := stat.Mean(finalValues, nil)
	medianFinal := stat.Quantile(0.50, stat.Empirical, finalValues, nil)

	return ScenarioRunRiskMetrics{
		VaR95:             var95,
		VaR99:             var99,
		CVaR95:            cvar95,
		CVaR99:            cvar99,
		ProbabilityOfLoss: probabilityOfLoss,
		MaxDrawdownP95:    maxDrawdownP95,
		MeanFinalValue:    meanFinal,
		MedianFinalValue:  medianFinal,
	}
}

func selectSamplePaths(results []*SimulationResult) []SamplePath {
	n := len(results)

	// results are already sorted by FinalValue from buildScenarioResponse
	percentiles := []struct {
		percentile float64
		label      string
	}{
		{0.05, "5th Percentile"},
		{0.25, "25th Percentile"},
		{0.50, "Median"},
		{0.75, "75th Percentile"},
		{0.95, "95th Percentile"},
	}

	// plus two are for the max drawdown and max volatility
	samplePaths := make([]SamplePath, 0, len(percentiles)+2)
	for _, p := range percentiles {
		idx := int(p.percentile * float64(n-1))
		samplePaths = append(samplePaths, SamplePath{
			Percentile: p.percentile,
			Values:     results[idx].PathValues,
			Label:      p.label,
		})
	}

	// below are two metrics that are pre calculated in the monte carlo simulation service
	// maximum drawdown
	maxDrawdownIdx := 0
	maxDrawdownValue := results[0].MaxDrawdown

	// most volatile path
	maxVolatilityIdx := 0
	maxVolatilityValue := results[0].AnnualizedVolatility

	for i, res := range results {
		if res.MaxDrawdown > maxDrawdownValue {
			maxDrawdownValue = res.MaxDrawdown
			maxDrawdownIdx = i
		}

		if res.AnnualizedVolatility > maxVolatilityValue {
			maxVolatilityValue = res.AnnualizedVolatility
			maxVolatilityIdx = i
		}
	}

	samplePaths = append(samplePaths, SamplePath{
		Percentile: -1,
		Values:     results[maxDrawdownIdx].PathValues,
		Label:      "Maximum Drawdown",
	})

	samplePaths = append(samplePaths, SamplePath{
		Percentile: -1,
		Values:     results[maxVolatilityIdx].PathValues,
		Label:      "Highest Volatility",
	})

	return samplePaths
}

func calculateSummaryStats(results []*SimulationResult) ScenarioStats {
	nResults := len(results)
	nSteps := len(results[0].PathValues)

	mean := make([]float64, nSteps)
	stdDev := make([]float64, nSteps)
	p5 := make([]float64, nSteps)
	p25 := make([]float64, nSteps)
	p50 := make([]float64, nSteps)
	p75 := make([]float64, nSteps)
	p95 := make([]float64, nSteps)

	for t := range nSteps {
		// get values at a point in time
		values := make([]float64, nResults)
		for i := range nResults {
			values[i] = results[i].PathValues[t]
		}

		// stat.Quantile requires the slice to be sorted in increasing order
		slices.Sort(values)

		mean[t] = stat.Mean(values, nil)
		stdDev[t] = stat.StdDev(values, nil)
		p5[t] = stat.Quantile(0.05, stat.Empirical, values, nil)
		p25[t] = stat.Quantile(0.25, stat.Empirical, values, nil)
		p50[t] = stat.Quantile(0.50, stat.Empirical, values, nil)
		p75[t] = stat.Quantile(0.75, stat.Empirical, values, nil)
		p95[t] = stat.Quantile(0.95, stat.Empirical, values, nil)
	}

	return ScenarioStats{
		Mean:   mean,
		StdDev: stdDev,
		P5:     p5,
		P25:    p25,
		P50:    p50,
		P75:    p75,
		P95:    p95,
	}
}

// calculateCVaR calculates the conditional value at risk for a given alpha aka taking the mean of the tail
func calculateCVaR(sortedReturns []float64, alpha float64) float64 {
	nReturns := len(sortedReturns)
	cutoff := int(math.Ceil(alpha * float64(nReturns)))
	return stat.Mean(sortedReturns[:cutoff], nil)
}
