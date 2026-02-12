SELECT
    configuration_id,
    asset_id,
    "weight"
FROM scenario_configuration_component
WHERE configuration_id = @id