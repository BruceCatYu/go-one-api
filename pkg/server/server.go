package server

import (
	"github.com/BruceCatYu/go-one-api/pkg/config"
	"github.com/gin-gonic/gin"
	"strings"
)

func Auth(c *gin.Context) {
	authTokens := strings.Split(c.GetHeader("Authorization"), " ")
	if len(authTokens) != 2 || authTokens[0] != "Bearer" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}
	if authTokens[1] != config.GetServerConfig().Key {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}
}

func StartServer() {
	r := gin.Default()
	r.Use(Auth)

	v1 := r.Group("/v1")
	{
		chat := v1.Group("/chat")
		{
			chat.POST("completions", chatCompletions)
		}
	}
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// Start server
	r.Run(":" + config.GetServerConfig().Port)
}
