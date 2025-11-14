package repos

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	m "mc.data/models"
)

var id int64

func Test_Base_CanGetConnectionAndPing(t *testing.T) {
	ctx := context.Background()
	pg := getConnection(t, ctx)
	err := pg.Ping(ctx)

	if err != nil {
		t.Errorf("error pinging postgres database: %s", err)
	}
}

func Test_TimeSeriesMetaDataRepo_CanInsert(t *testing.T) {
	symbol := "_TEST"
	
	testMetaData := m.TimeSeriesMetadata{
		Symbol:        symbol,
		LastRefreshed: time.Date(2025, time.October, 31, 0, 0, 0, 0, time.UTC),
	}

	ctx := context.Background()
	pg := getConnection(t, ctx)
	
	var err error
	id, err = pg.InsertNewMetaData(ctx, &testMetaData)
	if err != nil {
		t.Fatalf("error inserting new meta data: %s", err)
	}

	res, err := pg.GetMetaDataBySymbol(ctx, symbol)

	if err != nil {
		t.Fatalf("error getting meta data by symbol, %s", err)
	}

	if testMetaData.Symbol != res.Symbol {
		t.Fatalf("symbols did not match, inserted %s, got back %s", testMetaData.Symbol, res.Symbol)
	}

	if testMetaData.LastRefreshed != res.LastRefreshed {
		t.Fatalf("last refreshed time did not match, inserted %s, got back %s", testMetaData.LastRefreshed.Format(time.RFC3339), res.LastRefreshed.Format(time.RFC3339))
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
		res.deleteTestTimeSeriesData(t, ctx)
		res.Close()
	})

	return res
}

func (pg *Postgres) deleteTestTimeSeriesData(t *testing.T, ctx context.Context) {
	t.Helper()

	sql := `DELETE FROM time_series_data WHERE source_id = @source_id;
			DELETE FROM time_series_metadata WHERE id = @source_id`
	pg.Execute(ctx, sql, pgx.NamedArgs{"source_id": id})
}