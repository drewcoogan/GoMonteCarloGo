package core

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"slices"
	"time"

	"gonum.org/v1/gonum/stat"
	dm "mc.data/models"
	sm "mc.service/models"
)

// TODO: this is where we can add a queue to only run one scenario at a time
// can probably send and manage the queue after its validated and scenario is good
func (sc *ServiceContext) RunSimulation(scenarioID int32, settings sm.SimulationRequestSettings) (*dm.SimulationResponse, error) {
	start := time.Now()
	scenario, err := sc.PostgresConnection.GetScenarioByID(sc.Context, scenarioID)
	if err != nil {
		log.Printf("Error getting scenario id %v: %v", scenarioID, err)
		return nil, err
	}

	log.Printf("Recieved request to run scenario: %v", scenario.Name)
	maxLookbackDate := time.Now().Add(-settings.MaxLookback)
	log.Printf("Inserting scenario %v to simulation run history (time: %v)", scenario.Name, time.Since(start))
	dmSimulationRunHistory := sm.MapSimulationRequestSettingsToSimulationRunHistory(settings, maxLookbackDate)
	simulationRunId, err := sc.PostgresConnection.InsertSimulationRunHistory(sc.Context, scenario.Id, dmSimulationRunHistory)
	if err != nil {
		log.Printf("Error inserting scenario %v to simulation run history: %v", scenario.Name, err)
		return nil, err
	}

	log.Printf("Validating scenario %v (time: %v)", scenario.Name, time.Since(start))
	if err := validateScenario(scenario); err != nil {
		log.Printf("Error validating scenario %v: %v", scenario.Name, err)
		return sc.markSimulationRunAsFailure(simulationRunId, err.Error())
	}

	log.Printf("Getting series returns for scenario %v (time: %v)", scenario.Name, time.Since(start))
	seriesReturns, err := sc.getSeriesReturns(scenario, maxLookbackDate)
	if err != nil {
		log.Printf("Error getting series returns for scenario %v: %v", scenario.Name, err)
		return nil, err
	}

	log.Printf("Getting statistical resources for scenario %v (time: %v)", scenario.Name, time.Since(start))
	statisticalResources, err := GetStatisticalResources(seriesReturns, settings)
	if err != nil {
		log.Printf("Error getting statistical resources for scenario %v: %v", scenario.Name, err)
		return nil, err
	}

	log.Printf("Running monte carlo simulation for scenario %v (time: %v)", scenario.Name, time.Since(start))
	res, err := sc.RunMonteCarloSimulation(statisticalResources, settings)
	if err != nil {
		log.Printf("Error running monte carlo simulation for scenario %v: %v", scenario.Name, err)
		return sc.markSimulationRunAsFailure(simulationRunId, err.Error())
	}

	if err := sc.PostgresConnection.UpdateSimulationRunAsSuccess(sc.Context, simulationRunId); err != nil {
		log.Printf("Error updating simulation run as success for scenario %v: %v", scenario.Name, err)
		return nil, err // not making this as failure here, if we cant update it to success, we most likely cant update it to failure either
	}

	log.Printf("Building simulation response for scenario %v (time: %v)", scenario.Name, time.Since(start))
	response := buildSimulationResponse(res, settings.Seed)

	log.Printf("Saving simulation result for scenario %v (time: %v)", scenario.Name, time.Since(start))
	if err := sc.PostgresConnection.InsertSimulationResult(sc.Context, simulationRunId, response); err != nil {
		log.Printf("Error saving simulation result for scenario %v: %v", scenario.Name, err)
		return nil, err
	}

	log.Printf("Simulation for scenario %v completed (time: %v)", scenario.Name, time.Since(start))
	return response, nil
}

func validateScenario(scenario *dm.Scenario) error {
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
		if _, ok := v[a.AssetId]; ok {
			return fmt.Errorf("duplicate assetId %d", a.AssetId)
		}
		v[a.AssetId] = true
	}

	return nil
}

func (sc *ServiceContext) markSimulationRunAsFailure(runId int32, errorMessage string) (*dm.SimulationResponse, error) {
	return nil, sc.PostgresConnection.UpdateSimulationRunAsFailure(sc.Context, runId, errorMessage)
}

// randomScenarioPathsCount is how many extra paths (not used by expected exemplars) we expose for the chart.
const randomScenarioPathsCount = 5

func buildSimulationResponse(results []*SimulationResult, seed int64) *dm.SimulationResponse {
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
	exemplars, reserved := selectExemplarPaths(results)
	rng := rand.New(rand.NewSource(seed))
	extra := selectRoleSamplePaths(rng, results, reserved, randomScenarioPathsCount)
	samplePaths := make([]dm.SamplePath, 0, len(exemplars)+len(extra))
	samplePaths = append(samplePaths, exemplars...)
	samplePaths = append(samplePaths, extra...)
	summary := calculateSummaryStats(results)

	return &dm.SimulationResponse{
		RiskMetrics: riskMetrics,
		SamplePaths: samplePaths,
		Summary:     summary,
	}
}

func calculateRiskMetrics(results []*SimulationResult) dm.SimulationRiskMetrics {
	n := len(results)

	totalReturns := make([]float64, n)
	annualizedReturns := make([]float64, n)
	maxDrawdowns := make([]float64, n)

	for i, res := range results {
		totalReturns[i] = res.TotalReturn
		annualizedReturns[i] = res.AnnualizedReturn
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

	// MeanFinalValue / MedianFinalValue store annualized return (decimal, e.g. 0.06 == 6%/yr) for API consumers showing %.
	meanAnn := stat.Mean(annualizedReturns, nil)
	medianAnn := stat.Quantile(0.50, stat.Empirical, annualizedReturns, nil)

	return dm.SimulationRiskMetrics{
		VaR95:             var95,
		VaR99:             var99,
		CVaR95:            cvar95,
		CVaR99:            cvar99,
		ProbabilityOfLoss: probabilityOfLoss,
		MaxDrawdownP95:    maxDrawdownP95,
		MeanFinalValue:    meanAnn,
		MedianFinalValue:  medianAnn,
	}
}

func selectExemplarPaths(results []*SimulationResult) ([]dm.SamplePath, map[int]struct{}) {
	n := len(results)
	reserved := make(map[int]struct{})

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

	out := make([]dm.SamplePath, 0, len(percentiles)+2)
	for _, p := range percentiles {
		idx := int(p.percentile * float64(n-1))
		reserved[idx] = struct{}{}
		out = append(out, dm.SamplePath{
			Role:       dm.PathRolePercentile,
			Percentile: p.percentile,
			Values:     results[idx].PathValues,
			Label:      p.label,
		})
	}

	maxDrawdownIdx := 0
	maxDrawdownValue := results[0].MaxDrawdown
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

	reserved[maxDrawdownIdx] = struct{}{}
	out = append(out, dm.SamplePath{
		Role:       dm.PathRoleMaxDrawdown,
		Percentile: -1,
		Values:     results[maxDrawdownIdx].PathValues,
		Label:      "Maximum Drawdown",
	})

	reserved[maxVolatilityIdx] = struct{}{}
	out = append(out, dm.SamplePath{
		Role:       dm.PathRoleMaxVolatility,
		Percentile: -1,
		Values:     results[maxVolatilityIdx].PathValues,
		Label:      "Highest Volatility",
	})

	return out, reserved
}

func selectRoleSamplePaths(rng *rand.Rand, results []*SimulationResult, reserved map[int]struct{}, count int) []dm.SamplePath {
	n := len(results)
	if n == 0 || count <= 0 {
		return nil
	}
	// Only paths whose final value sits between the 5th and 95th percentile outcomes (by sorted final value).
	lowIdx := int(0.05 * float64(n-1))
	highIdx := int(0.95 * float64(n-1))
	candidates := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if _, used := reserved[i]; used {
			continue
		}
		if i < lowIdx || i > highIdx {
			continue
		}
		candidates = append(candidates, i)
	}
	if len(candidates) == 0 {
		return nil
	}
	if count > len(candidates) {
		count = len(candidates)
	}
	order := rng.Perm(len(candidates))
	out := make([]dm.SamplePath, 0, count)
	for k := 0; k < count; k++ {
		idx := candidates[order[k]]
		out = append(out, dm.SamplePath{
			Role:       dm.PathRoleSample,
			Percentile: -1,
			Label:      fmt.Sprintf("Sample %d", k+1),
			Values:     results[idx].PathValues,
		})
	}
	return out
}

func calculateSummaryStats(results []*SimulationResult) dm.SimulationStats {
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

	return dm.SimulationStats{
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
