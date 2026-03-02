package repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	dm "mc.data/models"
	q "mc.data/queries"
)

func (pg *Postgres) InsertSimulationRunHistory(ctx context.Context, scenarioId int32, simulationRunHistory dm.SimulationRunHistory) (int32, error) {
	sql := q.Get(q.QueryHelper.Insert.SimulationRunHistory)
	args := pgx.NamedArgs{
		"scenario_id":             scenarioId,
		"max_lookback":            simulationRunHistory.MaxLookback,
		"distribution_type":       simulationRunHistory.DistributionType,
		"simulation_unit_of_time": simulationRunHistory.SimulationUnitOfTime,
		"simulation_duration":     simulationRunHistory.SimulationDuration,
		"iterations":              simulationRunHistory.Iterations,
		"seed":                    simulationRunHistory.Seed,
		"degrees_of_freedom":      simulationRunHistory.DegreesOfFreedom,
	}

	var run_id int32
	if err := pg.db.QueryRow(ctx, sql, args).Scan(&run_id); err != nil {
		return 0, fmt.Errorf("error inserting simulation run history: %w", err)
	}

	return run_id, nil
}

func (pg *Postgres) UpdateSimulationRunAsFailure(ctx context.Context, run_id int32, error_message string) error {
	clean_error_message := strings.TrimSpace(error_message)
	if clean_error_message == "" {
		return fmt.Errorf("error message is required if simulation run is failing, occureed in %d", run_id)
	}

	return pg.updateSimulationRun(ctx, pgx.NamedArgs{
		"id":            run_id,
		"error_message": clean_error_message,
	})
}

func (pg *Postgres) UpdateSimulationRunAsSuccess(ctx context.Context, run_id int32) error {
	return pg.updateSimulationRun(ctx, pgx.NamedArgs{
		"id":            run_id,
		"error_message": nil,
	})
}

func (pg *Postgres) GetSimulationRunHistories(ctx context.Context, scenario_id int32, top_n int) ([]*dm.SimulationRun, error) {
	sql := q.Get(q.QueryHelper.Select.SimulationRunHistoriesByScenarioId)
	args := pgx.NamedArgs{"scenario_id": scenario_id, "top_n": top_n}
	runs, err := Query[dm.SimulationRunHistory](ctx, pg, sql, args)
	if err != nil {
		return nil, fmt.Errorf("unable to get simulation run histories: %w", err)
	}
	if len(runs) == 0 {
		return []*dm.SimulationRun{}, nil
	}

	runIds := make([]int32, len(runs))
	for i, r := range runs {
		runIds[i] = r.Id
	}

	sql = q.Get(q.QueryHelper.Select.SimulationRunHistoryComponentsByRunIds)
	args = pgx.NamedArgs{"run_ids": runIds}
	components, err := Query[dm.SimulationRunHistoryComponent](ctx, pg, sql, args)
	if err != nil {
		return nil, fmt.Errorf("unable to get simulation run history components: %w", err)
	}

	componentLookup := make(map[int32][]dm.SimulationRunHistoryComponent)
	for _, c := range components {
		componentLookup[c.RunId] = append(componentLookup[c.RunId], *c)
	}

	res := make([]*dm.SimulationRun, 0, len(runs))
	for _, r := range runs {
		res = append(res, &dm.SimulationRun{
			SimulationRunHistory: *r,
			Components:           componentLookup[r.Id],
		})
	}
	return res, nil
}

func (pg *Postgres) updateSimulationRun(ctx context.Context, args pgx.NamedArgs) error {
	sql := q.Get(q.QueryHelper.Update.SimulationRunHistory)
	if _, err := pg.db.Exec(ctx, sql, args); err != nil {
		return fmt.Errorf("error updating simulation run: %w", err)
	}
	return nil
}
