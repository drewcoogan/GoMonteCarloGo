UPDATE 
    av_time_series_metadata
SET 
    last_refreshed = @last_refreshed
WHERE 
    symbol = @symbol