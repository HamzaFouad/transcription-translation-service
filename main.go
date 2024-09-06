package main

import (
	"fmt"
	"log"
	"transcriptions-translation-service/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	router := gin.Default()

	log.Println("Starting server on port", cfg.Port)
	err := router.Run(":" + cfg.Port)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
