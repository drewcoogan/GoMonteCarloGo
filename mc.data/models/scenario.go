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

// ScenarioRun is a single entity to store the scenario run history and multiple components at the time of run
type ScenarioRun struct {
	ScenarioRunHistory
	Components []ScenarioRunHistoryComponent
}

// ScenarioRunHistory is the history of a scenario run, will keep the run id, scenario id, configuration id, error message, start time, and end time
// If I ever get to a point where I expand to users, will track user ids here as well, or any other relevant info.
// TODO: add the other fields used in the simulation, it will be selected on the fly by the user, but we want to keep track of the parameters used.
type ScenarioRunHistory struct {
	Id            int32     `db:"id"`
	ScenarioId    int32     `db:"scenario_id"` // foreign key to scenario configuration
	Name          string    `db:"name"`
	FloatedWeight bool      `db:"floated_weight"`
	ErrorMessage  string    `db:"error_message"`
	StartTimeUtc  time.Time `db:"start_time_utc"`
	EndTimeUtc    time.Time `db:"end_time_utc"`
}

type ScenarioRunHistoryComponent struct {
	RunId   int32   `db:"run_id"`
	AssetId int32   `db:"asset_id"`
	Weight  float64 `db:"weight"`
}
