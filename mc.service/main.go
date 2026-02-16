package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	r "mc.data/repos"
	av "mc.service/api/alpha_vantage"
	c "mc.service/core"
)

func main() {
    // initialize context and signal handler, listen for interrupt and term signals
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
    
    // load in environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded: %v", err)
	}

    // get alpha vantage client
    avClient := av.GetClient(os.Getenv("ALPHAVANTAGE_API_KEY"))

    // get postgres connection
    postgresConnection, err := r.GetPostgresConnection(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer postgresConnection.Close()

    // if we need to have any other connections, we can add them here
    // redis, queue, etc.

	sc := c.ServiceContext{
		Context:            ctx,
		PostgresConnection: postgresConnection,
		AlphaVantageClient: avClient,
	}
    
    // get http server, makes all of the endpoints and routes
    s := c.GetHttpServer(sc)

    // start http server in goroutine
    go func() {
        log.Printf("Starting GMCG server on %s", s.Addr)
        if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    // golang channel, will wait here until the context is closed (ie, ctrl+C)
    <-ctx.Done()
    log.Println("Received shutdown signal, shutting down gracefully...")
    
    // this gives the server 10 seconds to shutdown gracefully
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer shutdownCancel()
    
    if err := s.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }
    
    log.Println("Server stopped successfully")
}

