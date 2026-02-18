module mc.service

go 1.25.1

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/joho/godotenv v1.5.1
	mc.data v0.0.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.6 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
)

require (
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/exp v0.0.0-20251125195548-87e1e737ad39 // indirect
	golang.org/x/sync v0.18.0
	golang.org/x/text v0.31.0 // indirect
	gonum.org/v1/gonum v0.16.0
)

replace mc.data => ../mc.data
