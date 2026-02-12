INSERT INTO av_time_series_metadata 
    (symbol, last_refreshed) 
VALUES 
    (@symbol, @last_refreshed) 
RETURNING id