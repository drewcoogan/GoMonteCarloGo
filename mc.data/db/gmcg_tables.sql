--CREATE DATABASE GoMonteCarloGo;
--DROP TABLE av_time_series_data;
--DROP TABLE av_time_series_metadata;

-- create meta data table
CREATE TABLE IF NOT EXISTS av_time_series_metadata (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    last_refreshed DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uq_time_series_metadata_symbol UNIQUE (symbol)
);

CREATE OR REPLACE FUNCTION update_av_time_series_metadata_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_av_time_series_metadata_updated_at
    BEFORE UPDATE ON av_time_series_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_av_time_series_metadata_updated_at();

-- create time series data
CREATE TABLE IF NOT EXISTS av_time_series_data (
    source_id INTEGER NOT NULL,
    "timestamp" DATE NOT NULL,
    "open" NUMERIC(20, 4) NOT NULL,
    high NUMERIC(20, 4) NOT NULL,
    low NUMERIC(20, 4) NOT NULL,
    "close" NUMERIC(20, 4) NOT NULL,
    volume NUMERIC(20, 0) NOT NULL,
    adjusted_close NUMERIC(20, 4) NOT NULL,
    dividend_amount NUMERIC(20, 4) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uq_source_timestamp UNIQUE (source_id, timestamp),
    CONSTRAINT fk_time_series_data_metadata FOREIGN KEY (source_id)
        REFERENCES av_time_series_metadata(id)
        ON DELETE CASCADE -- what does CASCADE do? 
);

CREATE INDEX IF NOT EXISTS idx_time_series_source_timestamp ON av_time_series_data(source_id, timestamp DESC);

-- create table to store scenario meta data
CREATE TABLE IF NOT EXISTS scenario_configuration (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(100) NOT NULL,
    floated_weight BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- create table to store scenario components
CREATE TABLE IF NOT EXISTS scenario_configuration_component (
    id SERIAL PRIMARY KEY,
    configuration_id INTEGER NOT NULL,
    asset_id INTEGER NOT NULL,
    "weight" NUMERIC(8, 6) NOT NULL, -- is this the proper precision?

    CONSTRAINT uq_scenario_configuration_component UNIQUE (configuration_id, asset_id),
    
    CONSTRAINT fk_scenario_configuration FOREIGN KEY (configuration_id)
        REFERENCES scenario_configuration(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_av_time_series_metadata FOREIGN KEY (asset_id)
        REFERENCES av_time_series_metadata(id)
);

CREATE INDEX IF NOT EXISTS idx_scenario_configuration_component_configuration_id
    ON scenario_configuration_component(configuration_id);