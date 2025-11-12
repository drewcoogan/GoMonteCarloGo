package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	time_series_data     = "time_series_data"
	time_series_metadata = "time_series_metadata"
)

type Postgres struct {
	db *pgxpool.Pool
}

func GetPostgresConnection(ctx context.Context, connectionString string) (*Postgres, error) {
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

	return &Postgres{pool}, nil
}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *Postgres) Close() {
	pg.db.Close()
}

func (pg *Postgres) Execute(ctx context.Context, query string, args pgx.NamedArgs) (rowsAffected int64, err error) {
	commandTag, err := pg.db.Exec(ctx, query, args)
	if err != nil {
	  return
	}
	rowsAffected = commandTag.RowsAffected()
	return
}

// this seems silly and overly complex, dont need to abstract this.
func (pg *Postgres) QueryRow(ctx context.Context, query string, args pgx.NamedArgs) (id int64, err error) {
    err = pg.db.QueryRow(ctx, query, args).Scan(&id)
	return
}

func (pg *Postgres) BulkInsert(ctx context.Context, table_name string, columns []string, data [][]any) (int64, error) {
	return pg.db.CopyFrom(ctx, pgx.Identifier{table_name}, columns, pgx.CopyFromRows(data))
}

func Query[T any](ctx context.Context, pg *Postgres, query string, args pgx.NamedArgs) ([]T, error) {
    rows, err := pg.db.Query(ctx, query, args)
    if err != nil {
        return nil, fmt.Errorf("unable to query: %w", err)
    }
    defer rows.Close()
    
    return pgx.CollectRows(rows, pgx.RowToStructByName[T]) // return pointer to array? could be a lot of data
}

func QuerySingle[T any](ctx context.Context, pg *Postgres, query string, args pgx.NamedArgs) (*T, error) {
	res, err := Query[T](ctx, pg, query, args)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	return &res[0], nil
}