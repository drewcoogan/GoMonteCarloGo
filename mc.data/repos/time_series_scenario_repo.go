package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	m "mc.data/models"
)

func (pg *Postgres) GetScenarios(ctx context.Context) ([]*m.Scenario, error) {
	scenarioQuery := `
		SELECT
			id,
			name,
			floated_weight,
			created_at,
			updated_at
		FROM scenario_configuration
		WHERE deleted_at IS NULL
	`

	scenarios, err := Query[m.ScenarioConfiguration](ctx, pg, scenarioQuery, pgx.NamedArgs{})
	if err != nil {
		return nil, fmt.Errorf("unable to get scenarios: %w", err)
	}

	componentQuery := `
		SELECT
			configuration_id,
			asset_id,
			weight
		FROM scenario_configuration_component scc
		JOIN scenario_configuration sc ON scc.configuration_id = sc.id
		WHERE sc.deleted_at IS NULL
		ORDER BY configuration_id
	`

	components, err := Query[m.ScenarioConfigurationComponent](ctx, pg, componentQuery, pgx.NamedArgs{})
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
	scenarioQuery := `
		SELECT
			id,
			name,
			floated_weight,
			created_at,
			updated_at
		FROM scenario_configuration
		WHERE id = @id
			AND deleted_at IS NULL
	`

	scenarios, err := Query[m.ScenarioConfiguration](ctx, pg, scenarioQuery, pgx.NamedArgs{"id": id})
	if err != nil {
		return nil, fmt.Errorf("unable to get scenario by id: %w", err)
	}
	if len(scenarios) == 0 {
		return nil, nil
	}

	componentQuery := `
		SELECT
			configuration_id,
			asset_id,
			weight
		FROM scenario_configuration_component
		WHERE configuration_id = @id
	`

	components, err := Query[m.ScenarioConfigurationComponent](ctx, pg, componentQuery, pgx.NamedArgs{"id": id})
	if err != nil {
		return nil, fmt.Errorf("unable to get scenario components by id: %w", err)
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

	insertScenarioQuery := `
		INSERT INTO scenario_configuration
			(name, floated_weight)
		VALUES
			(@name, @floated_weight)
		RETURNING id, created_at, updated_at
	`

	config := m.ScenarioConfiguration{
		Name:          ns.Name,
		FloatedWeight: ns.FloatedWeight,
	}

	args := pgx.NamedArgs{
		"name":           ns.Name,
		"floated_weight": ns.FloatedWeight,
	}

	if err := tx.QueryRow(ctx, insertScenarioQuery, args).Scan(&config.Id, &config.CreatedAt, &config.UpdatedAt); err != nil {
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

	columns := []string{"configuration_id", "asset_id", "weight"}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"scenario_configuration_component"}, columns, pgx.CopyFromRows(componentRows)); err != nil {
		return nil, fmt.Errorf("error inserting scenario components: %w", err)
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

	updateScenarioQuery := `
		UPDATE scenario_configuration
		SET name = @name,
			floated_weight = @floated_weight,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = @id
			AND deleted_at IS NULL
		RETURNING id, name, floated_weight, created_at, updated_at
	`

	updateArgs := pgx.NamedArgs{
		"id":             scenarioID,
		"name":           ns.Name,
		"floated_weight": ns.FloatedWeight,
	}

	var config m.ScenarioConfiguration
	if err := tx.QueryRow(ctx, updateScenarioQuery, updateArgs).Scan(
		&config.Id,
		&config.Name,
		&config.FloatedWeight,
		&config.CreatedAt,
		&config.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("scenario not found")
		}
		return nil, fmt.Errorf("error updating scenario configuration: %w", err)
	}

	deleteComponentsQuery := `DELETE FROM scenario_configuration_component WHERE configuration_id = @id`
	if _, err := tx.Exec(ctx, deleteComponentsQuery, pgx.NamedArgs{"id": scenarioID}); err != nil {
		return nil, fmt.Errorf("error deleting existing scenario components: %w", err)
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

	columns := []string{"configuration_id", "asset_id", "weight"}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"scenario_configuration_component"}, columns, pgx.CopyFromRows(componentRows)); err != nil {
		return nil, fmt.Errorf("error inserting scenario components: %w", err)
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
	deleteScenarioQuery := `
		UPDATE scenario_configuration
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = @id
			AND deleted_at IS NULL
	`

	tag, err := tx.Exec(ctx, deleteScenarioQuery, pgx.NamedArgs{"id": scenarioID})

	if err != nil {
		return fmt.Errorf("error deleting scenario: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("scenario not found")
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
