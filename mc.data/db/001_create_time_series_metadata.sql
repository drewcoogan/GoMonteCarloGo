CREATE TABLE IF NOT EXISTS time_series_metadata (
    id SERIAL PRIMARY KEY,
    information VARCHAR(255),
    symbol VARCHAR(50) NOT NULL,
    last_refreshed DATE NOT NULL,
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