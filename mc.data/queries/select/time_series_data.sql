SELECT 
    atsd.source_id,
    atsd."timestamp", 
    atsd."open", 
    atsd.high, 
    atsd.low, 
    atsd."close", 
    atsd.volume, 
    atsd.adjusted_close, 
    atsd.dividend_amount
FROM av_time_series_data atsd 
JOIN av_time_series_metadata atsm ON atsd.source_id = atsm.id
WHERE atsm.symbol = @symbol
ORDER BY atsd."timestamp" DESC