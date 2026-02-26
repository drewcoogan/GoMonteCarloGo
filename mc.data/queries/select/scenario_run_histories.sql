SELECT
    id,
    scenario_id,
    "name",
    floated_weight,
    distribution_type,
    simulation_unit_of_time,
    simulation_duration,
    max_lookback,
    iterations,
    seed,
    degrees_of_freedom,
    error_message,
    start_time_utc,
    end_time_utc
FROM scenario_run_history
WHERE scenario_id = @scenario_id
ORDER BY start_time_utc DESC
LIMIT @top_n
