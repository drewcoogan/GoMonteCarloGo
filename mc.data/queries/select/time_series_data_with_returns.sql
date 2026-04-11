SELECT
    atsd.source_id,
    atsd."timestamp",
    atsd."open",
    atsd.high,
    atsd.low,
    atsd."close",
    atsd.volume,
    atsd.adjusted_close,
    atsd.dividend_amount,
    CASE
        WHEN lag_1 IS NULL OR lag_1 = 0 THEN NULL
        ELSE atsd.adjusted_close / lag_1 - 1
    END AS daily_return,
    CASE
        WHEN lag_5 IS NULL OR lag_5 = 0 THEN NULL
        ELSE atsd.adjusted_close / lag_5 - 1
    END AS rolling_5d_return
FROM (
    SELECT
        atsd_inner.source_id,
        atsd_inner."timestamp",
        atsd_inner."open",
        atsd_inner.high,
        atsd_inner.low,
        atsd_inner."close",
        atsd_inner.volume,
        atsd_inner.adjusted_close,
        atsd_inner.dividend_amount,
        LAG(atsd_inner.adjusted_close) OVER (PARTITION BY atsd_inner.source_id ORDER BY atsd_inner."timestamp") AS lag_1,
        LAG(atsd_inner.adjusted_close, 5) OVER (PARTITION BY atsd_inner.source_id ORDER BY atsd_inner."timestamp") AS lag_5
    FROM av_time_series_data atsd_inner
    JOIN av_time_series_metadata atsm_inner ON atsd_inner.source_id = atsm_inner.id
    WHERE atsm_inner.symbol = @symbol
) atsd
ORDER BY atsd."timestamp" ASC
