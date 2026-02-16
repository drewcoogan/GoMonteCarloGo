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

func GetPostgresConnection(ctx context.Context, connectionString string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(connectionString)

	if err != nil {
		return nil, fmt.Errorf("error parsing pgx connection string: %w", err)
	}

	config.MaxConns = 10 // TODO: make this defined in the config string or env variable
	config.MinConns = 2  // TODO: make this defined in the config string or env variable
	pool, err := pgxpool.NewWithConfig(context.Background(), config)

	if err != nil {
		return nil, fmt.Errorf("error making new pgx pool: %w", err)
	}

	return &Postgres{db: pool}, nil
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

type PgxExecutor interface {
	Query(ctx context.Context, query string, args pgx.NamedArgs) (pgx.Rows, error)
}

// Query is a wrapper around the pgxpool.Query method, allowing for the use of a QueryExecutor interface
// We need this because the default pg.db expects a ...args rather than the pgx.NamedArgs type
func (pg *Postgres) Query(ctx context.Context, query string, args pgx.NamedArgs) (pgx.Rows, error) {
	return pg.db.Query(ctx, query, args)
}

func Query[T any](ctx context.Context, p PgxExecutor, sql string, args pgx.NamedArgs) ([]*T, error) {
	rows, err := p.Query(ctx, sql, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	result := make([]*T, len(res))
	for i := range res {
		result[i] = &res[i]
	}

	return result, nil
}

func QuerySingle[T any](ctx context.Context, p PgxExecutor, query string, args pgx.NamedArgs) (*T, error) {
	res, err := Query[T](ctx, p, query, args)

	if err != nil {
		return nil, err
	}

	if len(res) != 1 {
		return nil, fmt.Errorf("expected a single result but got %d results", len(res))
	}

	return res[0], nil
}
