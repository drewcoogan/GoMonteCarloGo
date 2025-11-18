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
	startupCtx, startupCancel := context.WithTimeout(context.Background(), 30 * time.Second)
    ctx, stop := signal.NotifyContext(startupCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer startupCancel() // does order matter? 

    
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded: %v", err)
	}

    avClient := av.GetClient(os.Getenv("ALPHAVANTAGE_API_KEY"))    
    postgresConnection, err := r.GetPostgresConnection(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer postgresConnection.Close()

	sc := c.ServiceContext{
		Context:            ctx,
		PostgresConnection: postgresConnection,
		AlphaVantageClient: avClient,
	}
    
    s := c.GetHttpServer(sc)

    go func() {
        log.Printf("Starting GMCG server on %s", s.Addr)
        if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    <-ctx.Done() // golang channel, this will (in theory) pause the code until the context is closed (ie, ctrl+C)
    log.Println("Received shutdown signal, shutting down gracefully...")
    
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer shutdownCancel()
    
    if err := s.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }
    
    log.Println("Server stopped successfully")
}

