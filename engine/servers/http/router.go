package http

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (e *Engine) initRoutes() {

	e.router.Use(gin.Recovery())

	// prometheus metrics
	e.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	e.router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	suites := e.router.Group("/suites")
	{
		suites.POST("/", e.HandlerSuiteRun)
	}
}
