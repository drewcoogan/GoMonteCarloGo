SELECT
    id,
    "name",
    floated_weight,
    created_at,
    updated_at
FROM scenario_configuration
WHERE deleted_at IS NULL