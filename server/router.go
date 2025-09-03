package server

import (
	"codesage/config"
	"codesage/github"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.POST("/webhook", func(c *gin.Context) {
		github.HandleWebhook(c, cfg)
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "CodeSage is running"})
	})
	return r
}
