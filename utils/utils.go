package utils

import (
	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, status int, errMsg string) {
	c.JSON(status, gin.H{"error": errMsg})
	c.Abort()
}
