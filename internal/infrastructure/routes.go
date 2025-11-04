package infrastructure

import (
	httpadapter "goshop/internal/adapters/input/http"
	"goshop/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type Handlers struct {
	UserHandler     *httpadapter.UserHandler
	AddressHandler  *httpadapter.AddressHandler
	CategoryHandler *httpadapter.CategoryHandler
	ProductHandler  *httpadapter.ProductHandler
	CartHandler     *httpadapter.CartHandler
	OrderHandler    *httpadapter.OrderHandler
	ReviewHandler   *httpadapter.ReviewHandler
	OAuthHandler    *httpadapter.OAuthHandler
}

func RegisterRoutes(router *gin.Engine, handlers *Handlers, jwtSecret string, logger *zap.Logger) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
	router.POST("/products/append", handlers.ProductHandler.SaveProductImg)

	router.GET("/reviews", handlers.ReviewHandler.GetReviews)
	router.GET("/reviews/:id", handlers.ReviewHandler.GetReviewByID)
	router.GET("/reviews/stats/:productId", handlers.ReviewHandler.GetProductReviewStats)

	router.GET("/auth/google/login", handlers.OAuthHandler.GoogleLogin)
	router.GET("/auth/google/callback", handlers.OAuthHandler.GoogleCallback)

	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTMiddleware(jwtSecret, logger))
	{
		protected.PUT("/profile/avatar", handlers.UserHandler.UploadAvatar)
		protected.GET("/avatar", handlers.UserHandler.GetAvatar)

		protected.GET("/profile", handlers.UserHandler.GetProfile)
		protected.PUT("/profile", handlers.UserHandler.UpdateProfile)

		protected.GET("/addresses", handlers.AddressHandler.GetUserAddresses)
		protected.POST("/addresses", handlers.AddressHandler.CreateAddress)
		protected.GET("/addresses/:id", handlers.AddressHandler.GetAddressByID)
		protected.PUT("/addresses/:id", handlers.AddressHandler.UpdateAddress)
		protected.DELETE("/addresses/:id", handlers.AddressHandler.DeleteAddress)

		protected.POST("/reviews", handlers.ReviewHandler.CreateReview)
		protected.PUT("/reviews/:id", handlers.ReviewHandler.UpdateReview)
		protected.DELETE("/reviews/:id", handlers.ReviewHandler.DeleteReview)
	}

	admin := router.Group("/admin")
	admin.Use(middleware.JWTMiddleware(jwtSecret, logger))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.POST("/categories", handlers.CategoryHandler.CreateCategory)
		admin.PUT("/categories/:id", handlers.CategoryHandler.UpdateCategory)
		admin.DELETE("/categories/:id", handlers.CategoryHandler.DeleteCategory)

		admin.POST("/products/:id/images", handlers.ProductHandler.SaveProductImg)
		admin.DELETE("/products/:id/images/:img_id", handlers.ProductHandler.DeleteProductImg)

		admin.POST("/products", handlers.ProductHandler.CreateProduct)
		admin.GET("/products", handlers.ProductHandler.GetProducts)
		admin.PUT("/products/:id", handlers.ProductHandler.UpdateProduct)
		admin.DELETE("/products/:id", handlers.ProductHandler.DeleteProduct)
		admin.PATCH("/products/:id/toggle", handlers.ProductHandler.ToggleProductStatus)
	}
}
