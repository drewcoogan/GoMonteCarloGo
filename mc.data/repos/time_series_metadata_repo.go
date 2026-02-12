package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	m "mc.data/models"
	q "mc.data/queries"
)

func (pg *Postgres) GetMetaDataBySymbol(ctx context.Context, symbol string) (*m.TimeSeriesMetadata, error) {
	sql := q.Get(q.QueryHelper.Select.MetaDataBySymbol)
	args := pgx.NamedArgs{"symbol": symbol}

	if res, err := Query[m.TimeSeriesMetadata](ctx, pg, sql, args); err != nil {
		return nil, fmt.Errorf("unable to query metadata by symbol (%s): %w", symbol, err)
	} else if len(res) == 0 {
		return nil, nil
	} else {
		return res[0], nil
	}
}

func (pg *Postgres) InsertNewMetaData(ctx context.Context, metadata *m.TimeSeriesMetadata, tx pgx.Tx) (err error) {
	sql := q.Get(q.QueryHelper.Insert.Metadata)
	args := pgx.NamedArgs{
		"symbol":         metadata.Symbol,
		"last_refreshed": metadata.LastRefreshed,
	}

	if tx == nil {
		err = pg.db.QueryRow(ctx, sql, args).Scan(&metadata.Id)
	} else {
		err = tx.QueryRow(ctx, sql, args).Scan(&metadata.Id)
	}

	if err != nil {
		return fmt.Errorf("error inserting new metadata: %w", err)
	}

	return nil
}

func (pg *Postgres) UpdateLastRefreshedDate(ctx context.Context, symbol string, lastRefreshed time.Time, tx pgx.Tx) (err error) {
	sql := q.Get(q.QueryHelper.Update.LastRefreshedDate)
	args := pgx.NamedArgs{
		"symbol":         symbol,
		"last_refreshed": lastRefreshed,
	}

	if tx == nil {
		_, err = pg.db.Exec(ctx, sql, args)
	} else {
		_, err = tx.Exec(ctx, sql, args)
	}

	return
}

func (pg *Postgres) GetAllMetaData(ctx context.Context) ([]*m.TimeSeriesMetadata, error) {
	sql := q.Get(q.QueryHelper.Select.AllMetaData)
	args := pgx.NamedArgs{}

	if res, err := Query[m.TimeSeriesMetadata](ctx, pg, sql, args); err != nil {
		return nil, fmt.Errorf("unable to query metadata: %w", err)
	} else {
		return res, nil
	}
}
