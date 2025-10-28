package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NumbersToSum struct {
	Number1 int `json:"number1"`
	Number2 int `json:"number2"`
}

const (
	DefaultAddr = "localhost:8080"
)

func main() {
	router := gin.Default()
	router.GET("/api/ping", ping)
	router.GET("/api/addByGet", addByGet)
	router.POST("/api/addByPost", addByPost)
	log.Printf("Starting GMCG server on %s", DefaultAddr)
	http.ListenAndServe(DefaultAddr, router)
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
	
	res := number1 + number2
	log.Printf("Adding %d and %d to get %d", number1, number2, res)
	c.JSON(http.StatusOK, gin.H{"result": res})
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

