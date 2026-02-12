SELECT 
    MAX(atsd.timestamp)
FROM av_time_series_data atsd 
JOIN av_time_series_metadata atsm ON atsd.source_id = atsm.id
WHERE atsm.symbol = @symbol