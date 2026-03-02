SELECT
    run_id,
    asset_id,
    "weight"
FROM simulation_run_history_component
WHERE run_id = ANY(@run_ids)
ORDER BY run_id, asset_id
