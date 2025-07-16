package routes

import (
	"github.com/gin-gonic/gin"
	"goshop/internal/handler"
)

func RegisterRoutes(router *gin.Engine, handler *handler.AuthHandler) {
	v1 := router.Group("/api/v1")

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1.POST("/register", handler.Register)
	v1.POST("/login", handler.Login)

}
