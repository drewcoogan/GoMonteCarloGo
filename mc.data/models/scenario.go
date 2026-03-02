package models

import (
	"time"
)

// Scenario is a single entity to store the scenario configuration and multiple components
type Scenario struct {
	ScenarioConfiguration
	Components []ScenarioConfigurationComponent
}

// ScenarioConfiguration is the configuration for a scenario.
// Will keep configuation data for a scenario, probably need to add more here down the road
type ScenarioConfiguration struct {
	Id            int32     `db:"id"`
	Name          string    `db:"name"`
	FloatedWeight bool      `db:"floated_weight"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// ScenarioConfigurationComponent is a component of a scenario configuration, like ticker and associated weight
type ScenarioConfigurationComponent struct {
	ConfigurationId int32   `db:"configuration_id"`
	AssetId         int32   `db:"asset_id"`
	Weight          float64 `db:"weight"`
}

// SimulationRun is a single entity to store the simulation run history and multiple components at the time of run
type SimulationRun struct {
	SimulationRunHistory
	Components []SimulationRunHistoryComponent
}

// SimulationRunHistory is the history of a simulation run (when a scenario is executed), will keep the run id, scenario id, error message, start time, and end time
// If I ever get to a point where I expand to users, will track user ids here as well, or any other relevant info.
type SimulationRunHistory struct {
	Id                   int32     `db:"id" json:"id"`
	ScenarioId           int32     `db:"scenario_id"` // foreign key to scenario configuration
	Name                 string    `db:"name" json:"name"`
	FloatedWeight        bool      `db:"floated_weight" json:"floatedWeight"`
	DistributionType     string    `db:"distribution_type" json:"distributionType"`
	SimulationUnitOfTime string    `db:"simulation_unit_of_time" json:"simulationUnitOfTime"`
	SimulationDuration   int       `db:"simulation_duration" json:"simulationDuration"` // will be in units of simulation_unit_of_time
	MaxLookback          time.Time `db:"max_lookback" json:"maxLookback"`               // cutoff date for time series query (reference_time - lookback duration), computed on insert
	Iterations           int       `db:"iterations" json:"iterations"`
	Seed                 int64     `db:"seed" json:"seed"`
	DegreesOfFreedom     int       `db:"degrees_of_freedom" json:"degreesOfFreedom"`
	ErrorMessage         string    `db:"error_message" json:"errorMessage"`
	StartTimeUtc         time.Time `db:"start_time_utc" json:"startTimeUtc"`
	EndTimeUtc           time.Time `db:"end_time_utc" json:"endTimeUtc"`
}

// TODO: need to add asset details here, like symbol, name, etc.
type SimulationRunHistoryComponent struct {
	RunId   int32   `db:"run_id"`
	AssetId int32   `db:"asset_id" json:"assetId"`
	Weight  float64 `db:"weight" json:"weight"`
}
