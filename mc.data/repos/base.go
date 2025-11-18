package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
}

func GetPostgresConnection(ctx context.Context, connectionString string) (Postgres, error) {
	config, err := pgxpool.ParseConfig(connectionString)

	if err != nil {
		panic(fmt.Sprintf("error parsing pgx connection string: %v", err))
	}

	config.MaxConns = 10
	config.MinConns = 2
	pool, err := pgxpool.NewWithConfig(context.Background(), config)

	if err != nil {
		panic(fmt.Sprintf("error making new pgx pool: %v", err))
	}

	return Postgres{pool}, nil
}

func (pg *Postgres) GetTransaction(ctx context.Context) (pgx.Tx, error) {
	return pg.db.Begin(ctx)
}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *Postgres) Close() {
	pg.db.Close()
}

func (pg *Postgres) BulkInsert(ctx context.Context, table_name string, columns []string, data [][]any, tx *pgx.Tx) (int64, error) {
	if tx == nil {
		return pg.db.CopyFrom(ctx, pgx.Identifier{table_name}, columns, pgx.CopyFromRows(data))
	}
	return (*tx).CopyFrom(ctx, pgx.Identifier{table_name}, columns, pgx.CopyFromRows(data))
}

func Query[T any](ctx context.Context, pg *Postgres, query string, args pgx.NamedArgs) ([]*T, error) {
    rows, err := pg.db.Query(ctx, query, args)
    if err != nil {
        return nil, fmt.Errorf("unable to query: %w", err)
    }
    defer rows.Close()
    
	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, fmt.Errorf("error occured while collecting rows in query: %w", err)
	}

    result := make([]*T, len(res))
    for i := range res {
        result[i] = &res[i]
    }
    
    return result, nil
}

func QuerySingle[T any](ctx context.Context, pg *Postgres, query string, args pgx.NamedArgs) (*T, error) {
	res, err := Query[T](ctx, pg, query, args)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("no results found")
	}
	if len(res) > 1 {
		return nil, fmt.Errorf("multiple results found")
	}

	return res[0], nil
}