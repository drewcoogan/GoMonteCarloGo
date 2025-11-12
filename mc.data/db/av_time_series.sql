-- create meta data table
CREATE TABLE IF NOT EXISTS av_time_series_metadata (
    id SERIAL PRIMARY KEY,
    information VARCHAR(255),
    symbol VARCHAR(50) NOT NULL,
    last_refreshed DATE NOT NULL,
    interval VARCHAR(50),
    output_size VARCHAR(10) CHECK (output_size IN ('compact', 'full')),
    time_zone VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uq_time_series_metadata_symbol UNIQUE (symbol)
);

CREATE OR REPLACE FUNCTION update_time_series_metadata_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_time_series_metadata_updated_at
    BEFORE UPDATE ON time_series_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_time_series_metadata_updated_at();

-- create time series data
CREATE TABLE IF NOT EXISTS av_time_series_data (
    source_id INTEGER NOT NULL,
    [timestamp] DATE NOT NULL,
    [open] NUMERIC(20, 4),
    high NUMERIC(20, 4),
    low NUMERIC(20, 4),
    [close] NUMERIC(20, 4),
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