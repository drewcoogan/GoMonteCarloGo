package core

import (
	"context"

	r "mc.data/repos"
	av "mc.service/api/alpha_vantage"
)

type ServiceContext struct {
	Context            context.Context
	PostgresConnection r.Postgres
	AlphaVantageClient av.AlphaVantageClient
}
