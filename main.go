package main

import (
	"fmt"
	"log"
	"net/http"
	"transcriptions-translation-service/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	log.Println("Starting server on port", cfg.Port)
	err := router.Run(":" + cfg.Port)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
