package core

import (
	"fmt"
	"math"

	m "mc.data/models"
)

// TODO: this is where we can add a queue to only run one scenario at a time
// can probably send and manage the queue after its validated and scenario is good
func (sc *ServiceContext) RunScenario(scenarioID int32, settings SimulationSettings) ([]*SimulationResult, error) {
	scenario, err := sc.PostgresConnection.GetScenarioByID(sc.Context, scenarioID)
	if err != nil {
		return nil, err
	}

	runId, err := sc.PostgresConnection.InsertScenarioRunHistory(sc.Context, scenario.Id)
	if err != nil {
		return nil, err
	}

	if err := validateScenario(scenario); err != nil {
		return sc.markScenarioRunAsFailure(runId, err.Error())
	}

	res, err := sc.RunEquityMonteCarloWithCovarianceMartix(scenario, settings)
	if err != nil {
		return sc.markScenarioRunAsFailure(runId, err.Error())
	}

	if err := sc.PostgresConnection.UpdateScenarioRunAsSuccess(sc.Context, runId); err != nil {
		return nil, err // TODO: should we mark as failure here? do we care?
	}

	return res, nil
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

func (sc *ServiceContext) markScenarioRunAsFailure(runId int32, errorMessage string) ([]*SimulationResult, error) {
	return nil, sc.PostgresConnection.UpdateScenarioRunAsFailure(sc.Context, runId, errorMessage)
}
