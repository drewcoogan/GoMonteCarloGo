package models

import (
	"time"

	dm "mc.data/models"
)

type ScenarioRequest struct {
	Name          string                     `json:"name"`
	FloatedWeight bool                       `json:"floatedWeight"`
	Components    []ScenarioComponentPayload `json:"components"`
}

type ScenarioResponse struct {
	Id            int32                      `json:"id"`
	Name          string                     `json:"name"`
	FloatedWeight bool                       `json:"floatedWeight"`
	CreatedAt     time.Time                  `json:"createdAt"`
	UpdatedAt     time.Time                  `json:"updatedAt"`
	Components    []ScenarioComponentPayload `json:"components"`
}

type ScenarioComponentPayload struct {
	AssetId int32   `json:"assetId"`
	Weight  float64 `json:"weight"`
}

func MapScenarioToResponse(scenario *dm.Scenario) ScenarioResponse {
	res := ScenarioResponse{
		Id:            scenario.Id,
		Name:          scenario.Name,
		FloatedWeight: scenario.FloatedWeight,
		CreatedAt:     scenario.CreatedAt,
		UpdatedAt:     scenario.UpdatedAt,
		Components:    make([]ScenarioComponentPayload, len(scenario.Components)),
	}

	for idx, component := range scenario.Components {
		res.Components[idx] = ScenarioComponentPayload{
			AssetId: component.AssetId,
			Weight:  component.Weight,
		}
	}

	return res
}

func MapScenarioRequestToDataModel(req ScenarioRequest) dm.Scenario {
	components := make([]dm.ScenarioConfigurationComponent, len(req.Components))
	for i, c := range req.Components {
		components[i] = dm.ScenarioConfigurationComponent{
			AssetId: c.AssetId,
			Weight:  c.Weight,
		}
	}

	return dm.Scenario{
		ScenarioConfiguration: dm.ScenarioConfiguration{
			Name:          req.Name,
			FloatedWeight: req.FloatedWeight,
		},
		Components: components,
	}
}
