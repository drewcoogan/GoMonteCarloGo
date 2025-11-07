# GoMonteCarloGo

To run service:
    cd mc.service
    # create a .env file based on env.example
    # e.g. copy and edit: cp env.example .env
    # then set THIRD_PARTY_API_KEY in .env
    go run main.go

To run web:
    cd mc.web
    npm start

Environment variables:
    mc.service expects THIRD_PARTY_API_KEY to be set. For local dev, create mc.service/.env (see mc.service/env.example). In production, inject the variable via your hosting environment or a secrets manager.

To run test(s):
    cd *folder with "_test.go" file*
    to run all tests: go test
    to run verbosely: go test -v
    to run single test: go test -run *test func*

To start postgresql:
    Install via cmd: brew install postgresql@16
    To run via cmd: brew services start postgresql@16
    To stop via cmd: brew services stop postgresql@16