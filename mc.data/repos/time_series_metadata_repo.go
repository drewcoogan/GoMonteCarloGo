package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"mc.data/models"
)

func (pg *Postgres) GetMetaDataBySymbol(ctx context.Context, symbol string) (*models.TimeSeriesMetadata, error) {
	query := `SELECT information, symbol, last_refreshed, output_size, time_zone 
		FROM time_series_metadata 
		WHERE symbol = @symbol`
	
	args := pgx.NamedArgs{
		"symbol": symbol,
	}

	res, err := QuerySingle[models.TimeSeriesMetadata](ctx, pg, query, args)
	if err != nil {
		return nil, fmt.Errorf("unable to query metadata by symbol (%s): %w", symbol, err)
	}
	return res, nil
}

func (pg *Postgres) InsertNewMetaData(ctx context.Context, metadata *models.TimeSeriesMetadata) (int64, error) {
	query := `INSERT INTO time_series_metadata (information, symbol, last_refreshed, output_size, time_zone) 
		VALUES (@information, @symbol, @last_refreshed, @output_size, @time_zone) 
		RETURNING id`

	args := pgx.NamedArgs{
		"information": metadata.Information,
		"symbol": metadata.Symbol,
		"last_refreshed": metadata.LastRefreshed,
		"output_size": metadata.OutputSize,
		"time_zone": metadata.TimeZone,
	}

	res, err := pg.QueryRow(ctx, query, args)
	if err != nil {
		return 0, fmt.Errorf("error inserting new metadata: %w", err)
	}

	return res, nil
}