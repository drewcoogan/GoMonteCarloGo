package models

import (
	"fmt"
	"math"
	"strings"
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
	SimulationUnitOfTime int `json:"simulationUnitOfTime"` // daily, weekly, monthly, quarterly, yearly (derived when horizon is set)
	SimulationDuration   int `json:"simulationDuration"`   // weekly steps (derived when horizon is set)

	// MaxLookbackCount + MaxLookbackUnit are the preferred API; MaxLookback duration is computed before run.
	MaxLookbackCount int    `json:"maxLookbackCount"`
	MaxLookbackUnit  string `json:"maxLookbackUnit"` // months, years (max 5 years)

	// SimulationHorizonCount + SimulationHorizonUnit are request-only (months or years). They normalize to
	// SimulationUnitOfTime=weekly and SimulationDuration=weekly step count, which is what we persist.
	SimulationHorizonCount int    `json:"simulationHorizonCount"`
	SimulationHorizonUnit  string `json:"simulationHorizonUnit"` // months, years

	MaxLookback time.Duration `json:"maxLookback"` // legacy: nanoseconds in JSON; ignored when MaxLookbackCount > 0
	Iterations  int           `json:"iterations"`
	Seed        int64         `json:"seed"`

	DegreesOfFreedom int `json:"degreesOfFreedom"` // degrees of freedom for student t distribution
}

const maxWeeklyHorizonSteps = 520 // 10 years × 52 weeks/year

// HorizonToWeeklySteps converts a user horizon (months or years) to weekly Monte Carlo steps.
func HorizonToWeeklySteps(count int, unit string) (int, error) {
	if count <= 0 {
		return 0, fmt.Errorf("simulationHorizonCount must be positive")
	}
	u := strings.ToLower(strings.TrimSpace(unit))
	switch u {
	case "years":
		if count > 10 {
			return 0, fmt.Errorf("simulation horizon must be at most 10 years")
		}
		return count * Weekly, nil
	case "months":
		if count > 120 {
			return 0, fmt.Errorf("simulation horizon must be at most 120 months (10 years)")
		}
		steps := int(math.Round(float64(count) * float64(Weekly) / 12.0))
		if steps < 1 {
			return 1, nil
		}
		if steps > maxWeeklyHorizonSteps {
			return 0, fmt.Errorf("simulation horizon exceeds 10 years")
		}
		return steps, nil
	default:
		return 0, fmt.Errorf("simulationHorizonUnit must be months or years")
	}
}

// MaxLookbackToDuration converts a user lookback count + unit into a duration (approximate months/years).
// Lookback is capped at five years (60 months or 5 years).
func MaxLookbackToDuration(count int, unit string) (time.Duration, error) {
	if count <= 0 {
		return 0, fmt.Errorf("maxLookbackCount must be positive")
	}
	u := strings.ToLower(strings.TrimSpace(unit))
	switch u {
	case "months":
		if count > 60 {
			return 0, fmt.Errorf("max lookback must be at most 60 months (5 years)")
		}
		return time.Duration(count) * 30 * 24 * time.Hour, nil
	case "years":
		if count > 5 {
			return 0, fmt.Errorf("max lookback must be at most 5 years")
		}
		return time.Duration(count) * 365 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("maxLookbackUnit must be months or years")
	}
}

// NormalizeSimulationRequestSettings fills SimulationUnitOfTime, SimulationDuration, and MaxLookback from
// horizon / lookback fields when provided; otherwise legacy duration and maxLookback are left as sent.
func NormalizeSimulationRequestSettings(s *SimulationRequestSettings) error {
	if s.SimulationHorizonCount > 0 {
		d, err := HorizonToWeeklySteps(s.SimulationHorizonCount, s.SimulationHorizonUnit)
		if err != nil {
			return err
		}
		s.SimulationUnitOfTime = Weekly
		s.SimulationDuration = d
	}

	if s.MaxLookbackCount > 0 {
		d, err := MaxLookbackToDuration(s.MaxLookbackCount, s.MaxLookbackUnit)
		if err != nil {
			return err
		}
		s.MaxLookback = d
	}

	if s.SimulationDuration <= 0 {
		return fmt.Errorf("simulation duration must be positive")
	}
	if s.MaxLookback <= 0 {
		return fmt.Errorf("max lookback must be positive")
	}
	return nil
}

func MapSimulationRequestSettingsToSimulationRunHistory(settings SimulationRequestSettings, maxLookback time.Time) dm.SimulationRunHistory {
	return dm.SimulationRunHistory{
		DistributionType:     DistTypeToString(settings.DistributionType),
		SimulationUnitOfTime: SimulationUnitOfTimeToString(settings.SimulationUnitOfTime),
		SimulationDuration:   settings.SimulationDuration,
		MaxLookback:          maxLookback,
		MaxLookbackCount:     settings.MaxLookbackCount,
		MaxLookbackUnit:      settings.MaxLookbackUnit,
		Iterations:           settings.Iterations,
		Seed:                 settings.Seed,
		DegreesOfFreedom:     settings.DegreesOfFreedom,
	}
}
