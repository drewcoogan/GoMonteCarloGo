SELECT
    configuration_id,
    asset_id,
    "weight"
FROM scenario_configuration_component scc
JOIN scenario_configuration sc ON scc.configuration_id = sc.id
WHERE sc.deleted_at IS NULL
ORDER BY configuration_id