# GoMonteCarloGo

The purpose of this project was to learn Golang, relearn one of my favorite academic topics (Monte Carlo Simulations), build a PostgreSQL database from scratch, and design a front end system. It's probably not great code--it's probably the equivalent of "a face only a mother could love," but that's my cross to bear. 

The structure is loosely set up how I manage .NET applications professionally, not necessarily how I've seen other Golang-based repos. Keeping that constant helped me think through the architecture, but I understand I will eventually need to get over long singular files.

## Running the Service
```bash
cd mc.service
# Create a .env file based on env.example
# e.g. copy and edit: cp env.example .env
# then set THIRD_PARTY_API_KEY in .env
go run main.go
```

## Running the Web Frontend
```bash
cd mc.web/frontend
npm start
```

## Environment Variables

`mc.service` expects `THIRD_PARTY_API_KEY` to be set. 

- **Local development**: Create `mc.service/.env` (see `mc.service/env.example`)
- **Production**: Inject the variable via your hosting environment or a secrets manager

## Running Tests
```bash
cd *folder with "_test.go" file*

# Run all tests
go test

# Run verbosely
go test -v

# Run a single test
go test -run *test_func*
```

## Updating Dependencies

If there are updates in other packages, force update them by running:
```bash
cd mc.service

# Directly update the dependency
go get -u mc.data

# Update the consuming module
go mod tidy
```

## PostgreSQL Setup

### Installation
```bash
brew install postgresql@16
```

### Commands
```bash
# Start PostgreSQL
brew services start postgresql@16

# Stop PostgreSQL
brew services stop postgresql@16

# Restart PostgreSQL
brew services restart postgresql@16
```

## Useful Commands
```bash
# Format files to Go standards
go fmt
```

## Go Notes

See [GONOTES.md](GONOTES.md) for lessons learned and Go-specific patterns encountered while building this project.