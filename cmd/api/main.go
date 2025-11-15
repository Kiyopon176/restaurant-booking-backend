package main

import (
	"fmt"
	"log"
	"restaurant-booking/internal/config"
	"restaurant-booking/internal/database"
	"restaurant-booking/internal/handler"
	"restaurant-booking/internal/repository"
	"restaurant-booking/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "restaurant-booking/docs"
)

// @title Restaurant Booking API
// @version 1.0
// @description API для системы бронирования столиков в ресторанах
// @host localhost:8088
// @BasePath /
// @schemes http
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := database.New(cfg.DB)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Successfully connected to database!")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)
	tableRepo := repository.NewTableRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	restaurantManagerRepo := repository.NewRestaurantManagerRepository(db)

	// Initialize services
	restaurantService := service.NewRestaurantService(restaurantRepo, db)
	tableService := service.NewTableService(tableRepo, restaurantRepo, db)
	managerService := service.NewManagerService(restaurantManagerRepo, restaurantRepo, userRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userRepo)
	restaurantHandler := handler.NewRestaurantHandler(restaurantService)
	tableHandler := handler.NewTableHandler(tableRepo)
	bookingHandler := handler.NewBookingHandler(bookingRepo, tableRepo)
	reviewHandler := handler.NewReviewHandler(reviewRepo, restaurantRepo)
	managerHandler := handler.NewManagerHandler(managerService)

	r := gin.Default()

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
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.GET("/:id/bookings", bookingHandler.GetUserBookings)
			users.GET("/:id/reviews", reviewHandler.GetUserReviews)
		}

		restaurants := api.Group("/restaurants")
		{
			restaurants.POST("", restaurantHandler.CreateRestaurant)
			restaurants.GET("", restaurantHandler.ListRestaurants)
			restaurants.GET("/search", restaurantHandler.SearchRestaurants)

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

		tables := api.Group("/tables")
		{
			tables.POST("", tableHandler.CreateTable)
			tables.GET("/available", tableHandler.GetAvailableTables)
			tables.GET("/:id", tableHandler.GetTable)
			tables.PUT("/:id", tableHandler.UpdateTable)
			tables.DELETE("/:id", tableHandler.DeleteTable)
		}

		bookings := api.Group("/bookings")
		{
			bookings.POST("", bookingHandler.CreateBooking)
			bookings.GET("/check-availability", bookingHandler.CheckTableAvailability)
			bookings.GET("/:id", bookingHandler.GetBooking)
			bookings.PATCH("/:id/status", bookingHandler.UpdateBookingStatus)
			bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
		}

		reviews := api.Group("/reviews")
		{
			reviews.POST("", reviewHandler.CreateReview)
			reviews.GET("/:id", reviewHandler.GetReview)
			reviews.PUT("/:id", reviewHandler.UpdateReview)
			reviews.DELETE("/:id", reviewHandler.DeleteReview)
		}
	}

	fmt.Printf("Server running on http://localhost:%s\n", cfg.Server.Port)
	fmt.Printf("Swagger UI: http://localhost:%s/swagger/index.html\n", cfg.Server.Port)
	fmt.Printf("Health check: http://localhost:%s/health\n", cfg.Server.Port)

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
