package utils

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, status int, errMsg string) {
	c.JSON(status, gin.H{"error": errMsg})
	c.Abort()
}

func SerializeToString(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("serialization error: %w", err)
	}
	return string(bytes), nil
}

func DeserializeFromString(input string, output interface{}) error {
	if err := json.Unmarshal([]byte(input), output); err != nil {
		return fmt.Errorf("deserialization error: %w", err)
	}
	return nil
}

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	})

	return router
}
