package main

import (
	_ "github.com/devxp-tech/demo-app/config"
	"github.com/devxp-tech/demo-app/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.New()
	server.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  true,
			"message": "Hello world for your app demo-app",
		})
	})
	server.GET("/health-check/liveness", controllers.HealthCheckLiveness)
	server.GET("/health-check/readiness", controllers.HealthCheckReadiness)
	server.Run()
}
