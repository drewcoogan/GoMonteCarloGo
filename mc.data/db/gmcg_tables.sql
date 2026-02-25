--CREATE DATABASE GoMonteCarloGo;
--DROP TABLE scenario_run_history_component;
--DROP TABLE scenario_run_history;
--DROP TABLE scenario_configuration_component;
--DROP TABLE scenario_configuration;
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
        ON DELETE CASCADE -- cascade will delete the data in this row if the key in metadata is deleted
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
-- column name configuration_id matches models and queries (db tag and SQL)
CREATE TABLE IF NOT EXISTS scenario_configuration_component (
    id SERIAL PRIMARY KEY, -- this is the primary key for the table for easy indexing
    configuration_id INTEGER NOT NULL, -- foreign key to scenario_configuration(id)
    asset_id INTEGER NOT NULL,
    "weight" NUMERIC(8, 6) NOT NULL,

    CONSTRAINT uq_scenario_configuration_component UNIQUE (configuration_id, asset_id),
    
    CONSTRAINT fk_scenario_configuration FOREIGN KEY (configuration_id)
        REFERENCES scenario_configuration(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_av_time_series_metadata FOREIGN KEY (asset_id)
        REFERENCES av_time_series_metadata(id)
);

CREATE INDEX IF NOT EXISTS idx_scenario_configuration_component_configuration_id
    ON scenario_configuration_component(configuration_id);

-- create table to store scenario runs
CREATE TABLE IF NOT EXISTS scenario_run_history (
    id SERIAL PRIMARY KEY,
    scenario_id INTEGER NOT NULL, -- id that will match off on which scenario is being ran
    "name" VARCHAR(100) NOT NULL,
    floated_weight BOOLEAN NOT NULL,
    distribution_type VARCHAR(50) NOT NULL DEFAULT '',
    simulation_unit_of_time VARCHAR(50) NOT NULL DEFAULT '',
    simulation_duration INTEGER NOT NULL DEFAULT 0,
    max_lookback DATE NOT NULL DEFAULT '1970-01-01', -- cutoff date for time series query (reference_time - lookback duration), computed on insert
    iterations INTEGER NOT NULL DEFAULT 0,
    seed BIGINT NOT NULL DEFAULT 0,
    degrees_of_freedom INTEGER NOT NULL DEFAULT 0,
    error_message TEXT DEFAULT NULL,
    start_time_utc TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    end_time_utc TIMESTAMPTZ DEFAULT NULL
);

-- create table to store scenario run components
CREATE TABLE IF NOT EXISTS scenario_run_history_component (
    id SERIAL PRIMARY KEY,
    run_id INTEGER NOT NULL,
    asset_id INTEGER NOT NULL,
    "weight" NUMERIC(8, 6) NOT NULL,
    -- TODO: verify if there are any mechanisms to delete asset id, if so, well want to keep ticker here also

    CONSTRAINT uq_scenario_run_component UNIQUE (run_id, asset_id),
    
    CONSTRAINT fk_scenario_run FOREIGN KEY (run_id)
        REFERENCES scenario_run_history(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_av_time_series_metadata FOREIGN KEY (asset_id)
        REFERENCES av_time_series_metadata(id) -- dont cascade, but will this be a problem? maybe, but we can make it not able to delete if a run exists, sounds like a user problem
);