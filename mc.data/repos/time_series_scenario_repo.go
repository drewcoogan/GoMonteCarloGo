package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	m "mc.data/models"
	q "mc.data/queries"
)

func (pg *Postgres) GetScenarios(ctx context.Context) ([]*m.Scenario, error) {
	sql := q.Get(q.QueryHelper.Select.AllScenarioConfigurations)
	args := pgx.NamedArgs{}
	scenarios, err := Query[m.ScenarioConfiguration](ctx, pg, sql, args)
	if err != nil {
		return nil, fmt.Errorf("unable to get scenarios: %w", err)
	}

	sql = q.Get(q.QueryHelper.Select.AllScenarioConfigurationComponents)
	args = pgx.NamedArgs{}
	components, err := Query[m.ScenarioConfigurationComponent](ctx, pg, sql, args)
	if err != nil {
		return nil, fmt.Errorf("unable to get scenario components: %w", err)
	}

	scenarioComponentsLookup := make(map[int32][]m.ScenarioConfigurationComponent)
	for _, v := range components {
		if scenarioComponentsLookup[v.ConfigurationId] == nil {
			scenarioComponentsLookup[v.ConfigurationId] = []m.ScenarioConfigurationComponent{}
		}
		scenarioComponentsLookup[v.ConfigurationId] = append(scenarioComponentsLookup[v.ConfigurationId], *v)
	}

	res := make([]*m.Scenario, 0, len(scenarios))
	for _, v := range scenarios {
		res = append(res, &m.Scenario{
			ScenarioConfiguration: *v,
			Components:            scenarioComponentsLookup[v.Id],
		})
	}

	return res, nil
}

func (pg *Postgres) GetScenarioByID(ctx context.Context, id int32) (*m.Scenario, error) {
	sql := q.Get(q.QueryHelper.Select.ScenarioConfigurationById)
	args := pgx.NamedArgs{"id": id}
	scenarios, err := Query[m.ScenarioConfiguration](ctx, pg, sql, args)
	if err != nil {
		return nil, fmt.Errorf("unable to get scenario by id (%d): %w", id, err)
	}

	if len(scenarios) == 0 {
		return nil, nil
	}

	sql = q.Get(q.QueryHelper.Select.ScenarioConfigurationComponentById)
	args = pgx.NamedArgs{"id": id}
	components, err := Query[m.ScenarioConfigurationComponent](ctx, pg, sql, args)
	if err != nil {
		return nil, fmt.Errorf("unable to get scenario components by id (%d): %w", id, err)
	}

	scenario := &m.Scenario{
		ScenarioConfiguration: *scenarios[0],
		Components:            make([]m.ScenarioConfigurationComponent, 0, len(components)),
	}

	for _, v := range components {
		scenario.Components = append(scenario.Components, *v)
	}

	return scenario, nil
}

func (pg *Postgres) InsertNewScenarioTx(ctx context.Context, ns m.NewScenario, tx pgx.Tx) (*m.Scenario, error) {
	if ns.Name == "" {
		return nil, fmt.Errorf("scenario name is required")
	}
	if len(ns.Components) == 0 {
		return nil, fmt.Errorf("scenario must include at least one component")
	}

	config := m.ScenarioConfiguration{
		Name:          ns.Name,
		FloatedWeight: ns.FloatedWeight,
	}

	sql := q.Get(q.QueryHelper.Insert.ScenarioConfiguration)
	args := pgx.NamedArgs{"name": ns.Name, "floated_weight": ns.FloatedWeight}
	if err := tx.QueryRow(ctx, sql, args).Scan(
		&config.Id,
		&config.CreatedAt,
		&config.UpdatedAt); err != nil {
		return nil, fmt.Errorf("error inserting scenario configuration: %w", err)
	}

	componentRows := make([][]any, len(ns.Components))
	components := make([]m.ScenarioConfigurationComponent, len(ns.Components))
	for i, c := range ns.Components {
		componentRows[i] = []any{config.Id, c.AssetId, c.Weight}
		components[i] = m.ScenarioConfigurationComponent{
			ConfigurationId: config.Id,
			AssetId:         c.AssetId,
			Weight:          c.Weight,
		}
	}

	table_name := pgx.Identifier{"scenario_configuration_component"}
	columns := []string{"configuration_id", "asset_id", "weight"}
	rows := pgx.CopyFromRows(componentRows)
	if _, err := tx.CopyFrom(ctx, table_name, columns, rows); err != nil {
		return nil, fmt.Errorf("error inserting scenario components (%d): %w", config.Id, err)
	}

	return &m.Scenario{
		ScenarioConfiguration: config,
		Components:            components,
	}, nil
}

func (pg *Postgres) InsertNewScenario(ctx context.Context, ns m.NewScenario) (*m.Scenario, error) {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	scenario, err := pg.InsertNewScenarioTx(ctx, ns, tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing scenario insert: %w", err)
	}

	return scenario, nil
}

func (pg *Postgres) UpdateExistingScenarioTx(ctx context.Context, scenarioID int32, ns m.NewScenario, tx pgx.Tx) (*m.Scenario, error) {
	if ns.Name == "" {
		return nil, fmt.Errorf("scenario name is required")
	}

	if len(ns.Components) == 0 {
		return nil, fmt.Errorf("scenario must include at least one component")
	}

	// update scenario configuration
	sql := q.Get(q.QueryHelper.Update.ScenarioConfiguration)
	args := pgx.NamedArgs{
		"id":             scenarioID,
		"name":           ns.Name,
		"floated_weight": ns.FloatedWeight,
	}

	var config m.ScenarioConfiguration
	if err := tx.QueryRow(ctx, sql, args).Scan(
		&config.Id,
		&config.Name,
		&config.FloatedWeight,
		&config.CreatedAt,
		&config.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("scenario (%d) not found", scenarioID)
		}
		return nil, fmt.Errorf("error updating scenario configuration (%d): %w", scenarioID, err)
	}

	// delete existing scenario components
	sql = q.Get(q.QueryHelper.Delete.ScenarioConfigurationComponentByConfigurationId)
	args = pgx.NamedArgs{"id": scenarioID}
	if _, err := tx.Exec(ctx, sql, args); err != nil {
		return nil, fmt.Errorf("error deleting existing scenario components (%d): %w", scenarioID, err)
	}

	componentRows := make([][]any, len(ns.Components))
	components := make([]m.ScenarioConfigurationComponent, len(ns.Components))
	for i, c := range ns.Components {
		componentRows[i] = []any{config.Id, c.AssetId, c.Weight}
		components[i] = m.ScenarioConfigurationComponent{
			ConfigurationId: config.Id,
			AssetId:         c.AssetId,
			Weight:          c.Weight,
		}
	}

	// insert new scenario components
	table_name := pgx.Identifier{"scenario_configuration_component"}
	columns := []string{"configuration_id", "asset_id", "weight"}
	rows := pgx.CopyFromRows(componentRows)
	if _, err := tx.CopyFrom(ctx, table_name, columns, rows); err != nil {
		return nil, fmt.Errorf("error inserting scenario components (%d): %w", scenarioID, err)
	}

	return &m.Scenario{
		ScenarioConfiguration: config,
		Components:            components,
	}, nil
}

func (pg *Postgres) UpdateExistingScenario(ctx context.Context, scenarioID int32, ns m.NewScenario) (*m.Scenario, error) {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	updated, err := pg.UpdateExistingScenarioTx(ctx, scenarioID, ns, tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing scenario update: %w", err)
	}

	return updated, nil
}

func (pg *Postgres) DeleteScenarioTx(ctx context.Context, scenarioID int32, tx pgx.Tx) error {
	sql := q.Get(q.QueryHelper.Delete.ScenarioConfiguration)
	args := pgx.NamedArgs{"id": scenarioID}
	tag, err := tx.Exec(ctx, sql, args)
	if err != nil {
		return fmt.Errorf("error deleting scenario (%d): %w", scenarioID, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("scenario not found (%d)", scenarioID)
	}

	return nil
}

func (pg *Postgres) DeleteScenario(ctx context.Context, scenarioID int32) error {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := pg.DeleteScenarioTx(ctx, scenarioID, tx); err != nil {
		return fmt.Errorf("error deleting scenario: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing scenario delete: %w", err)
	}

	return nil
}
