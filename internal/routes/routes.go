package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"goshop/internal/handler/address"
	"goshop/internal/handler/cart"
	"goshop/internal/handler/category"
	"goshop/internal/handler/product"
	"goshop/internal/handler/user"
	"goshop/internal/middleware"
)

type Handlers struct {
	UserHandler     *user.UserHandler
	AddressHandler  *address.AddressHandler
	CategoryHandler *category.CategoryHandler
	ProductHandler  *product.ProductHandler
	CartHandler     *cart.CartHandler
}

func RegisterRoutes(router *gin.Engine, handlers *Handlers, jwtSecret string, logger *zap.Logger) {

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.POST("/auth/register", handlers.UserHandler.Register)
	router.POST("/auth/login", handlers.UserHandler.Login)

	router.GET("/categories", handlers.CategoryHandler.GetAllCategories)
	router.GET("/categories/:id", handlers.CategoryHandler.GetCategoryByID)

	router.GET("/products", handlers.ProductHandler.GetProducts)
	router.GET("/products/:id", handlers.ProductHandler.GetProductByID)
	router.GET("/products/category/:id", handlers.ProductHandler.GetProductsByCategory)

	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTMiddleware(jwtSecret, logger))
	{

		protected.GET("/profile", handlers.UserHandler.GetProfile)
		protected.PUT("/profile", handlers.UserHandler.UpdateProfile)

		// Addresses
		protected.GET("/addresses", handlers.AddressHandler.GetUserAddresses)
		protected.POST("/addresses", handlers.AddressHandler.CreateAddress)
		protected.GET("/addresses/:id", handlers.AddressHandler.GetAddressByID)
		protected.PUT("/addresses/:id", handlers.AddressHandler.UpdateAddress)
		protected.DELETE("/addresses/:id", handlers.AddressHandler.DeleteAddress)
	}

	admin := router.Group("/admin")
	admin.Use(middleware.JWTMiddleware(jwtSecret, logger))
	admin.Use(middleware.AdminMiddleware())
	{

		admin.POST("/categories", handlers.CategoryHandler.CreateCategory)
		admin.PUT("/categories/:id", handlers.CategoryHandler.UpdateCategory)
		admin.DELETE("/categories/:id", handlers.CategoryHandler.DeleteCategory)

		admin.POST("/products", handlers.ProductHandler.CreateProduct)
		admin.GET("/products", handlers.ProductHandler.GetProducts)
		admin.PUT("/products/:id", handlers.ProductHandler.UpdateProduct)
		admin.DELETE("/products/:id", handlers.ProductHandler.DeleteProduct)
		admin.PATCH("/products/:id/toggle", handlers.ProductHandler.ToggleProductStatus)
	}
}
