UPDATE 
    scenario_configuration
SET 
    deleted_at = CURRENT_TIMESTAMP
WHERE 
    id = @id
    AND deleted_at IS NULL