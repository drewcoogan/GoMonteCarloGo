SELECT
    id,
    "name",
    floated_weight,
    created_at,
    updated_at
FROM scenario_configuration
WHERE id = @id
    AND deleted_at IS NULL