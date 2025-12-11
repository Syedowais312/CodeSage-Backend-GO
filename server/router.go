package server

import (
	"codesage/config"
	"codesage/github"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.POST("/github/webhook", func(c *gin.Context) {
		github.HandleWebhook(c, cfg)
	})

	r.GET("/auth/github/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "redirect to GitHub OAuth here"})
	})
	r.GET("/auth/github/callback", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "handle GitHub OAuth callback here"})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "CodeSage is running"})
	})
	return r
}
