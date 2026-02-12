INSERT INTO scenario_configuration
    (name, floated_weight)
VALUES
    (@name, @floated_weight)
RETURNING id, created_at, updated_at