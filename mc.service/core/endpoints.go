package core

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type NumbersToSum struct {
	Number1 int `json:"number1"`
	Number2 int `json:"number2"`
}

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

	engine.GET("/api/ping", func (c *gin.Context) { ping(c, sc) })
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

func ping(c *gin.Context, sc ServiceContext) {
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