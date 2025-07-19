package routes

import (
	"github.com/gin-gonic/gin"
	"goshop/internal/handler"
	"goshop/internal/middleware"
)

type Handlers struct {
	UserHandler    *handler.UserHandler
	AddressHandler *handler.AddressHandler
}

func RegisterRoutes(router *gin.Engine, handlers *Handlers, jwtSecret string) {

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.POST("/register", handlers.UserHandler.Register)
	router.POST("/login", handlers.UserHandler.Login)

	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTMiddleware(jwtSecret))
	{
		protected.GET("/profile", handlers.UserHandler.GetProfile)
		protected.PUT("/profile", handlers.UserHandler.UpdateProfile)

		protected.GET("/addresses", handlers.AddressHandler.GetUserAddresses)            // Все адреса пользователя
		protected.POST("/addresses", handlers.AddressHandler.CreateAddress)              // Создать адрес
		protected.GET("/addresses/:addressID", handlers.AddressHandler.GetAddressByID)   // Получить конкретный адрес
		protected.PUT("/addresses/:addressID", handlers.AddressHandler.UpdateAddress)    // Обновить адрес
		protected.DELETE("/addresses/:addressID", handlers.AddressHandler.DeleteAddress) // Удалить адрес
	}
}
