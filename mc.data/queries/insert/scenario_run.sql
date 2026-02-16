WITH inserted_run AS (
    INSERT INTO scenario_run_history 
        (scenario_id, configuration_id, start_time_utc)
    VALUES 
        (@scenario_id, @configuration_id, CURRENT_TIMESTAMP)
    RETURNING id
)

INSERT INTO scenario_run_history_component 
    (run_id, asset_id, "weight")
SELECT 
    ir.id, scc.asset_id, scc."weight"
FROM scenario_configuration_component scc
CROSS JOIN inserted_run ir
WHERE scc.configuration_id = @configuration_id
RETURNING ir.id