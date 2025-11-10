package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	r "mc.data/repos"
	"mc.service/api"
)

type NumbersToSum struct {
	Number1 int `json:"number1"`
	Number2 int `json:"number2"`
}

const (
	DefaultAddr = ":8080"
)

var (
	postgresConnection *r.Postgres
	avClient           api.AlphaVantageClient
)

func main() {
	startupCtx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
    ctx, stop := signal.NotifyContext(startupCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer cancel() // does order matter? 

    
    loadEnv()
    avClient = api.GetClient(os.Getenv("ALPHAVANTAGE_API_KEY"))    
    postgresConnection, err := r.GetPostgresConnection(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer postgresConnection.Close()
    
    s := establishApiEndpoints()

    go func() {
        log.Printf("Starting GMCG server on %s", s.Addr)
        if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    <-ctx.Done() // golang channel, this will (in theory) pause the code until the context is closed (ie, ctrl+C)
    log.Println("Received shutdown signal, shutting down gracefully...")
    
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()
    
    if err := s.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }
    
    log.Println("Server stopped successfully")
}
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded: %v", err)
	}
}

func establishApiEndpoints() *http.Server {
	engine := gin.Default()

	engine.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

	engine.GET("/api/ping", ping)
	engine.GET("/api/addByGet", addByGet)
	engine.POST("/api/addByPost", addByPost)

	server := &http.Server{
		Addr:           DefaultAddr,
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return server;
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func addByGet(c *gin.Context) {
	number1Str := c.Query("number1")
	number2Str := c.Query("number2")
	
	number1, err1 := strconv.Atoi(number1Str)
	number2, err2 := strconv.Atoi(number2Str)
	
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid numbers"})
		return
	}
	
	result := number1 + number2
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func addByPost(c *gin.Context) {
	var nums NumbersToSum
	if err := c.ShouldBindJSON(&nums); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result := nums.Number1 + nums.Number2
	c.JSON(http.StatusOK, gin.H{"result": result})
}

