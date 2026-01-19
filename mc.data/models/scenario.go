package models

import (
	"time"
)

type NewScenario struct {
	Name          string
	FloatedWeight bool
	Components    []NewComponent
}

type NewComponent struct {
	AssetId int32
	Weight  float64
}

type Scenario struct {
	ScenarioConfiguration
	Components []ScenarioConfigurationComponent
}

type ScenarioConfiguration struct {
	Id            int32     `db:"id"`
	Name          string    `db:"name"`
	FloatedWeight bool      `db:"floated_weight"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type ScenarioConfigurationComponent struct {
	ConfigurationId int32   `db:"configuration_id"`
	AssetId         int32   `db:"asset_id"`
	Weight          float64 `db:"weight"`
}
