package repos

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	dm "mc.data/models"
	q "mc.data/queries"
)

func (pg *Postgres) InsertSimulationResult(ctx context.Context, simulationRunId int32, simulationResponse *dm.SimulationResponse) error {
	bytes, err := json.Marshal(simulationResponse)
	if err != nil {
		return fmt.Errorf("error marshalling simulation response: %w", err)
	}

	sql := q.Get(q.QueryHelper.Insert.SimulationResult)
	args := pgx.NamedArgs{"simulation_run_id": simulationRunId, "result": bytes}
	if _, err := pg.db.Exec(ctx, sql, args); err != nil {
		return fmt.Errorf("error inserting simulation result: %w", err)
	}
	return nil
}

func (pg *Postgres) GetSimulationResult(ctx context.Context, simulationRunId int32) (*dm.SimulationResponse, error) {
	sql := q.Get(q.QueryHelper.Select.SimulationResult)
	args := pgx.NamedArgs{"simulation_run_id": simulationRunId}

	var bytes []byte
	if err := pg.db.QueryRow(ctx, sql, args).Scan(&bytes); err != nil {
		return nil, fmt.Errorf("error getting simulation result: %w", err)
	}

	var simulationResponse dm.SimulationResponse
	if err := json.Unmarshal(bytes, &simulationResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling simulation result: %w", err)
	}

	return &simulationResponse, nil
}
