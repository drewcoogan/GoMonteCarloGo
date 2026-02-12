WITH price_data AS (
    SELECT 
        t.source_id,
        t.timestamp,
        t.adjusted_close,
        LAG(t.adjusted_close) OVER (PARTITION BY t.source_id ORDER BY t.timestamp) AS prev_close
    FROM av_time_series_data t
    WHERE t.source_id = ANY(@source_ids)
        AND t.timestamp >= @max_lookback
)

SELECT 
    source_id,
    timestamp,
    LN(adjusted_close / prev_close) AS log_return
FROM price_data
WHERE prev_close IS NOT NULL
ORDER BY source_id, timestamp DESC