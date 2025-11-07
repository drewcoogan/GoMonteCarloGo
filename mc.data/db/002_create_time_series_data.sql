CREATE TABLE IF NOT EXISTS time_series_data (
    source_id INTEGER NOT NULL,
    timestamp DATE NOT NULL,
    open NUMERIC(20, 4) NOT NULL,
    high NUMERIC(20, 4) NOT NULL,
    low NUMERIC(20, 4) NOT NULL,
    close NUMERIC(20, 4) NOT NULL,
    adjusted_close NUMERIC(20, 4),
    volume NUMERIC(20, 0),
    dividend_amount NUMERIC(20, 4),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uq_source_timestamp UNIQUE (source_id, timestamp),
    CONSTRAINT fk_time_series_data_metadata FOREIGN KEY (source_id)
        REFERENCES time_series_metadata(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_time_series_source_timestamp ON time_series_data(source_id, timestamp DESC);