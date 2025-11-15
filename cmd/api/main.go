package main

import (
	"fmt"
	"log"
	"restaurant-booking/internal/config"
	"restaurant-booking/internal/database"
	"restaurant-booking/internal/handler"
	"restaurant-booking/internal/middleware"
	"restaurant-booking/internal/repository"
	"restaurant-booking/internal/service"
	"restaurant-booking/pkg/jwt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "restaurant-booking/docs"
)

// @title Restaurant Booking API
// @version 1.0
// @description API для системы бронирования столиков в ресторанах
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Successfully connected to database!")

	// Initialize JWT manager
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessExpire, cfg.JWTRefreshExpire)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)
	tableRepo := repository.NewTableRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	restaurantManagerRepo := repository.NewRestaurantManagerRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtManager)
	userService := service.NewUserService(userRepo)
	restaurantService := service.NewRestaurantService(restaurantRepo, db)
	// tableService := service.NewTableService(tableRepo, restaurantRepo, db) // Not used yet
	managerService := service.NewManagerService(restaurantManagerRepo, restaurantRepo, userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, userService)
	userHandler := handler.NewUserHandler(userRepo)
	restaurantHandler := handler.NewRestaurantHandler(restaurantService)
	tableHandler := handler.NewTableHandler(tableRepo)
	bookingHandler := handler.NewBookingHandler(bookingRepo, tableRepo)
	reviewHandler := handler.NewReviewHandler(reviewRepo, restaurantRepo)
	managerHandler := handler.NewManagerHandler(managerService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	api := r.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authMiddleware.Authenticate(), authHandler.Logout)
			auth.GET("/me", authMiddleware.Authenticate(), authHandler.GetMe)
		}

		// User routes
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			// users.GET("", userHandler.ListUsers) // Commented out - not implemented
			users.GET("/:id", userHandler.GetUser)
			users.GET("/:id/bookings", bookingHandler.GetUserBookings)
			users.GET("/:id/reviews", reviewHandler.GetUserReviews)
		}

		// Restaurant routes
		restaurants := api.Group("/restaurants")
		{
			restaurants.POST("", restaurantHandler.CreateRestaurant)
			restaurants.GET("", restaurantHandler.ListRestaurants)
			// restaurants.GET("/search", restaurantHandler.SearchRestaurants) // Not implemented yet

			restaurants.GET("/:id/tables", tableHandler.GetRestaurantTables)
			restaurants.GET("/:id/bookings", bookingHandler.GetRestaurantBookings)
			restaurants.GET("/:id/reviews", reviewHandler.GetRestaurantReviews)

			// Manager routes
			restaurants.POST("/:id/managers", managerHandler.AddManager)
			restaurants.GET("/:id/managers", managerHandler.GetManagers)
			restaurants.DELETE("/:id/managers/:user_id", managerHandler.RemoveManager)

			// Image routes
			restaurants.POST("/:id/images", restaurantHandler.AddImage)
			restaurants.DELETE("/:id/images/:image_id", restaurantHandler.DeleteImage)

			restaurants.GET("/:id", restaurantHandler.GetRestaurant)
			restaurants.PUT("/:id", restaurantHandler.UpdateRestaurant)
			restaurants.DELETE("/:id", restaurantHandler.DeleteRestaurant)
		}

		// Table routes
		tables := api.Group("/tables")
		{
			tables.POST("", tableHandler.CreateTable)
			tables.GET("/available", tableHandler.GetAvailableTables)
			tables.GET("/:id", tableHandler.GetTable)
			tables.PUT("/:id", tableHandler.UpdateTable)
			tables.DELETE("/:id", tableHandler.DeleteTable)
		}

		// Booking routes
		bookings := api.Group("/bookings")
		{
			bookings.POST("", bookingHandler.CreateBooking)
			bookings.GET("/check-availability", bookingHandler.CheckTableAvailability)
			bookings.GET("/:id", bookingHandler.GetBooking)
			bookings.PATCH("/:id/status", bookingHandler.UpdateBookingStatus)
			bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
		}

		// Review routes
		reviews := api.Group("/reviews")
		{
			reviews.POST("", reviewHandler.CreateReview)
			reviews.GET("/:id", reviewHandler.GetReview)
			reviews.PUT("/:id", reviewHandler.UpdateReview)
			reviews.DELETE("/:id", reviewHandler.DeleteReview)
		}
	}

	fmt.Printf("Server running on http://localhost:%s\n", cfg.Port)
	fmt.Printf("Swagger UI: http://localhost:%s/swagger/index.html\n", cfg.Port)
	fmt.Printf("Health check: http://localhost:%s/health\n", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
