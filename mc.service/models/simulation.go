package models

import (
	"time"

	dm "mc.data/models"
)

// SimulationSettingsResources will be the resources for the simulation settings, rest will be simple numbers provided by user
type SimulationSettingsResources struct {
	DistributionType     map[string]int `json:"distributionType"`     // standar normal, student t
	SimulationUnitOfTime map[string]int `json:"simulationUnitOfTime"` // daily, weekly, monthly, quarterly, yearly
	SimulationDuration   map[string]int `json:"simulationDuration"`   // number of units of time to simulate
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
		DistributionType:     distType,
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
	DistributionType     int `json:"distributionType"`     // standar normal, student t
	SimulationUnitOfTime int `json:"simulationUnitOfTime"` // daily, weekly, monthly, quarterly, yearly
	SimulationDuration   int `json:"simulationDuration"`   // number of units of time to simulate

	MaxLookback time.Duration `json:"maxLookback"`
	Iterations  int           `json:"iterations"`
	Seed        int64         `json:"seed"`

	DegreesOfFreedom int `json:"degreesOfFreedom"` // degrees of freedom for student t distribution
}

func MapSimulationRequestSettingsToSimulationRunHistory(settings SimulationRequestSettings, maxLookback time.Time) dm.SimulationRunHistory {
	return dm.SimulationRunHistory{
		DistributionType:     DistTypeToString(settings.DistributionType),
		SimulationUnitOfTime: SimulationUnitOfTimeToString(settings.SimulationUnitOfTime),
		SimulationDuration:   settings.SimulationDuration,
		MaxLookback:          maxLookback,
		Iterations:           settings.Iterations,
		Seed:                 settings.Seed,
		DegreesOfFreedom:     settings.DegreesOfFreedom,
	}
}
