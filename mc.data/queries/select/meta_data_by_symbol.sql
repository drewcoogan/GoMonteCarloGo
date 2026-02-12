SELECT 
    id, 
    symbol, 
    last_refreshed
FROM av_time_series_metadata 
WHERE symbol = @symbol