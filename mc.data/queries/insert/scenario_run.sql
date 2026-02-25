WITH inserted_run AS (
    INSERT INTO scenario_run_history 
        (scenario_id, 
        "name", 
        floated_weight, 
        distribution_type, 
        simulation_unit_of_time,
        simulation_duration, 
        max_lookback, 
        iterations, 
        seed, 
        degrees_of_freedom, 
        start_time_utc)
    SELECT 
        sc.id, 
        sc."name", 
        sc.floated_weight,
        @distribution_type, 
        @simulation_unit_of_time, 
        @simulation_duration,
        @max_lookback::date, 
        @iterations, 
        @seed, 
        @degrees_of_freedom, 
        CURRENT_TIMESTAMP
    FROM scenario_configuration sc
    WHERE sc.id = @scenario_id
    RETURNING id
),
inserted_components AS (
    INSERT INTO scenario_run_history_component 
        (run_id, asset_id, "weight")
    SELECT 
        ir.id, scc.asset_id, scc."weight"
    FROM scenario_configuration_component scc
    CROSS JOIN inserted_run ir
    WHERE scc.configuration_id = @scenario_id
)
SELECT id FROM inserted_run