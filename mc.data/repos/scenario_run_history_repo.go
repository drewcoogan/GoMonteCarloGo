package repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	q "mc.data/queries"
)

func (pg *Postgres) InsertScenarioRunHistory(ctx context.Context, scenarioId int32) (int32, error) {
	sql := q.Get(q.QueryHelper.Insert.ScenarioRun)
	args := pgx.NamedArgs{
		"scenario_id": scenarioId,
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

func (pg *Postgres) updateScenarioRun(ctx context.Context, args pgx.NamedArgs) error {
	sql := q.Get(q.QueryHelper.Update.ScenarioRun)
	if _, err := pg.db.Exec(ctx, sql, args); err != nil {
		return fmt.Errorf("error updating scenario run: %w", err)
	}
	return nil
}
