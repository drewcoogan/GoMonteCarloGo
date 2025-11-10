package repos

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"mc.data/models"
)

func Test_Base_CanGetConnectionAndPing(t *testing.T) {
	ctx := context.Background()
	pg := getConnection(t, ctx)
	err := pg.Ping(ctx)

	if err != nil {
		t.Errorf("error pinging postgres database: %s", err)
	}
}

func Test_TimeSeriesMetaDataRepo_CanInsert(t *testing.T) {
	ex := models.TimeSeriesMetadata{
		Information:   null.StringFrom("TEST INFO"),
		Symbol:        null.StringFrom("_TEST"),
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
		TimeZone:      null.StringFrom("testtimezone"),
	}

	ctx := context.Background()
	pg := getConnection(t, ctx)

	_, err := pg.InsertNewMetaData(ctx, &ex)

	if err != nil {
		
	}
}

func getConnection(t *testing.T, ctx context.Context) *Postgres {
	t.Helper()
	err := godotenv.Load("../.env");
	if err != nil {
		t.Fatalf("error loading environment: %s", err)
	}

	connectionString := os.Getenv("DATABASE_URL")
	res, err := GetPostgresConnection(ctx, connectionString)

	if err != nil {
		t.Fatalf("error getting postgres connection: %s", err)
	}

	// on test resolving, this will close connections, even if an error is thrown
	t.Cleanup(func() {
		res.Close()
	})

	return res
}

func (pg *Postgres) deleteTestTimeSeriesData(t *testing.T, ctx context.Context, id int64) {
	t.Helper()
	sql := `DELETE FROM time_series_data WHERE source_id = @source_id;
			DELETE FROM time_series_metadata WHERE id = @source_id`
			args := pgx.NamedArgs{"source_id": id}
	pg.Execute(ctx, sql, args)
}