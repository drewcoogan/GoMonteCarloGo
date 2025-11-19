package core

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	ex "mc.data/extensions"
)

const (
	DefaultAddr = ":8080"
)

func GetHttpServer(sc ServiceContext) *http.Server {
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
	engine.POST("/api/syncStockData", func (c *gin.Context) { syncStockData(c, sc) })

	// these two methods were used to figure out how this all works.
	engine.GET("/api/test/addByGet", addByGet)
	engine.POST("/api/test/addByPost", addByPost)

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

type SyncStockDataRequest struct {
	Symbol string `json:"symbol"`
}

func syncStockData(c *gin.Context, sc ServiceContext) {
	var req SyncStockDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	lut, err := sc.SyncSymbolTimeSeriesData(req.Symbol)
	if err != nil {
		if lut.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"date": ex.FmtShort(lut),
			"message": err.Error(),
		})
		return
	}

	// Get the updated metadata to return the last refreshed date
	md, err := sc.PostgresConnection.GetMetaDataBySymbol(sc.Context, req.Symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error getting metadata: %v", err)})
		return
	}

	if md == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "metadata not found after sync"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"date": ex.FmtShort(md.LastRefreshed)})
}


// Testing endpoints below to ensure functionality

type NumbersToSum struct {
	Number1 int `json:"number1"`
	Number2 int `json:"number2"`
}

// AddByGet adds two numbers via a GET request
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

// AddByPost adds two numbers via a POST request
func addByPost(c *gin.Context) {
	var nums NumbersToSum
	if err := c.ShouldBindJSON(&nums); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result := nums.Number1 + nums.Number2
	c.JSON(http.StatusOK, gin.H{"result": result})
}