UPDATE 
    scenario_configuration
SET 
    "name" = @name,
    floated_weight = @floated_weight,
    updated_at = CURRENT_TIMESTAMP
WHERE 
    id = @id
    AND deleted_at IS NULL
RETURNING 
    id,
    "name",
    floated_weight,
    created_at,
    updated_at