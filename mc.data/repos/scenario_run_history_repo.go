package repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	dm "mc.data/models"
	q "mc.data/queries"
)

func (pg *Postgres) InsertScenarioRunHistory(ctx context.Context, scenarioId int32, scenarioRunHistory dm.ScenarioRunHistory) (int32, error) {
	sql := q.Get(q.QueryHelper.Insert.ScenarioRun)
	args := pgx.NamedArgs{
		"scenario_id":             scenarioId,
		"max_lookback":            scenarioRunHistory.MaxLookback,
		"distribution_type":       scenarioRunHistory.DistributionType,
		"simulation_unit_of_time": scenarioRunHistory.SimulationUnitOfTime,
		"simulation_duration":     scenarioRunHistory.SimulationDuration,
		"iterations":              scenarioRunHistory.Iterations,
		"seed":                    scenarioRunHistory.Seed,
		"degrees_of_freedom":      scenarioRunHistory.DegreesOfFreedom,
	}

	var run_id int32
	if err := pg.db.QueryRow(ctx, sql, args).Scan(&run_id); err != nil {
		return 0, fmt.Errorf("error inserting scenario run history: %w", err)
	}

	return run_id, nil
}

func (pg *Postgres) UpdateScenarioRunAsFailure(ctx context.Context, run_id int32, error_message string) error {
	clean_error_message := strings.TrimSpace(error_message)
	if clean_error_message == "" {
		return fmt.Errorf("error message is required if scenario run is failing, occureed in %d", run_id)
	}

	return pg.updateScenarioRun(ctx, pgx.NamedArgs{
		"id":            run_id,
		"error_message": clean_error_message,
	})
}

func (pg *Postgres) UpdateScenarioRunAsSuccess(ctx context.Context, run_id int32) error {
	return pg.updateScenarioRun(ctx, pgx.NamedArgs{
		"id":            run_id,
		"error_message": nil,
	})
}

func (pg *Postgres) GetScenarioRunHistories(ctx context.Context, scenario_id int32, top_n int) ([]*dm.ScenarioRun, error) {
	sql := q.Get(q.QueryHelper.Select.ScenarioRunHistories)
	args := pgx.NamedArgs{"scenario_id": scenario_id, "top_n": top_n}
	runs, err := Query[dm.ScenarioRunHistory](ctx, pg, sql, args)
	if err != nil {
		return nil, fmt.Errorf("unable to get scenario run histories: %w", err)
	}
	if len(runs) == 0 {
		return []*dm.ScenarioRun{}, nil
	}

	runIds := make([]int32, len(runs))
	for i, r := range runs {
		runIds[i] = r.Id
	}

	sql = q.Get(q.QueryHelper.Select.ScenarioRunHistoryComponentsByRunIds)
	args = pgx.NamedArgs{"run_ids": runIds}
	components, err := Query[dm.ScenarioRunHistoryComponent](ctx, pg, sql, args)
	if err != nil {
		return nil, fmt.Errorf("unable to get scenario run history components: %w", err)
	}

	componentLookup := make(map[int32][]dm.ScenarioRunHistoryComponent)
	for _, c := range components {
		componentLookup[c.RunId] = append(componentLookup[c.RunId], *c)
	}

	res := make([]*dm.ScenarioRun, 0, len(runs))
	for _, r := range runs {
		res = append(res, &dm.ScenarioRun{
			ScenarioRunHistory: *r,
			Components:         componentLookup[r.Id],
		})
	}
	return res, nil
}

func (pg *Postgres) updateScenarioRun(ctx context.Context, args pgx.NamedArgs) error {
	sql := q.Get(q.QueryHelper.Update.ScenarioRun)
	if _, err := pg.db.Exec(ctx, sql, args); err != nil {
		return fmt.Errorf("error updating scenario run: %w", err)
	}
	return nil
}
