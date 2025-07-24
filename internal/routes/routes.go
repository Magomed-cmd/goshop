package routes

import (
	"github.com/gin-gonic/gin"
	"goshop/internal/handler/address"
	"goshop/internal/handler/category"
	"goshop/internal/handler/user"
	"goshop/internal/middleware"
)

type Handlers struct {
	UserHandler     *user.UserHandler
	AddressHandler  *address.AddressHandler
	CategoryHandler *category.CategoryHandler
}

func RegisterRoutes(router *gin.Engine, handlers *Handlers, jwtSecret string) {

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.POST("/register", handlers.UserHandler.Register)
	router.POST("/login", handlers.UserHandler.Login)

	router.GET("/categories", handlers.CategoryHandler.GetAllCategories)
	router.GET("/categories/:id", handlers.CategoryHandler.GetCategoryByID)

	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTMiddleware(jwtSecret))
	{
		protected.GET("/profile", handlers.UserHandler.GetProfile)
		protected.PUT("/profile", handlers.UserHandler.UpdateProfile)

		protected.GET("/addresses", handlers.AddressHandler.GetUserAddresses)
		protected.POST("/addresses", handlers.AddressHandler.CreateAddress)
		protected.GET("/addresses/:id", handlers.AddressHandler.GetAddressByID)
		protected.PUT("/addresses/:id", handlers.AddressHandler.UpdateAddress)
		protected.DELETE("/addresses/:id", handlers.AddressHandler.DeleteAddress)
	}

	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.JWTMiddleware(jwtSecret))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.POST("/categories", handlers.CategoryHandler.CreateCategory)
		admin.PUT("/categories/:id", handlers.CategoryHandler.UpdateCategory)
		admin.DELETE("/categories/:id", handlers.CategoryHandler.DeleteCategory)

		// TODO: добить эти эндпоинты, но пока они не обязательны
		// admin.GET("/users", handlers.UserHandler.GetAllUsers)
		// admin.DELETE("/users/:id", handlers.UserHandler.DeleteUser)
		// admin.GET("/orders", handlers.OrderHandler.GetAllOrders)
	}
}
