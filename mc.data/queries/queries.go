package queries

import (
	"embed"
	"fmt"
)

//go:embed delete/*.sql insert/*.sql select/*.sql update/*.sql
var Files embed.FS

// ^^^ the go:embed directive is used to embed the files in the queries package
// meaning on compile time it will convert the files to binary data and embed it in the queries package

type DeleteQueries struct {
	ScenarioConfiguration string
}

type InsertQueries struct {
	Metadata              string
	ScenarioConfiguration string
}

type SelectQueries struct {
	AllMetaData                        string
	AllScenarioConfigurationComponents string
	AllScenarioConfigurations          string
	MetaDataBySymbol                   string
	MostRecentTimestampBySymbol        string
	ScenarioConfigurationById          string
	ScenarioConfigurationComponentById string
	TimeSeriesData                     string
	TimeSeriesReturns                  string
}

type UpdateQueries struct {
	LastRefreshedDate     string
	ScenarioConfiguration string
}

type QueryHelperStruct struct {
	Delete DeleteQueries
	Insert InsertQueries
	Select SelectQueries
	Update UpdateQueries
}

var QueryHelper = QueryHelperStruct{
	Delete: DeleteQueries{
		ScenarioConfiguration: "delete/scenario_configuration.sql",
	},
	Insert: InsertQueries{
		Metadata:              "insert/metadata.sql",
		ScenarioConfiguration: "insert/scenario_configuration.sql",
	},
	Select: SelectQueries{
		AllMetaData:                        "select/all_meta_data.sql",
		AllScenarioConfigurationComponents: "select/all_scenario_configuration_components.sql",
		AllScenarioConfigurations:          "select/all_scenario_configurations.sql",
		MetaDataBySymbol:                   "select/meta_data_by_symbol.sql",
		MostRecentTimestampBySymbol:        "select/most_recent_timestamp_by_symbol.sql",
		ScenarioConfigurationById:          "select/scenario_configuration_by_id.sql",
		ScenarioConfigurationComponentById: "select/scenario_configuration_component_by_id.sql",
		TimeSeriesData:                     "select/time_series_data.sql",
		TimeSeriesReturns:                  "select/time_series_returns.sql",
	},
	Update: UpdateQueries{
		LastRefreshedDate:     "update/last_refreshed_date.sql",
		ScenarioConfiguration: "update/scenario_configuration.sql",
	},
}

func Get(path string) string {
	content, err := Files.ReadFile(path)

	if err != nil {
		panic(fmt.Errorf("error reading query file: %w", err))
	}

	return string(content)
}
