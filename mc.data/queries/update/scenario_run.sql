UPDATE 
    scenario_run_history
SET 
    error_message = @error_message,
    end_time_utc = CURRENT_TIMESTAMP
WHERE 
    id = @id