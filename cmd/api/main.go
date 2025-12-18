package main

import (
	"fmt"
	"restaurant-booking/internal/config"
	"restaurant-booking/internal/database"
	"restaurant-booking/internal/handler"
	"restaurant-booking/internal/middleware"
	"restaurant-booking/internal/repository"
	"restaurant-booking/internal/service"
	"restaurant-booking/pkg/jwt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/restaurant-booking/pkg/logger"
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
	log, _ := logger.New("debug")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", zap.Error(err))
	}

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Info("Successfully connected to database!")

	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessExpire, cfg.JWTRefreshExpire)

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)
	tableRepo := repository.NewTableRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	restaurantManagerRepo := repository.NewRestaurantManagerRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtManager, log)
	userService := service.NewUserService(userRepo, log)
	restaurantService := service.NewRestaurantService(restaurantRepo, db, log)
	walletService := service.NewWalletService(walletRepo, db, log)
	paymentService := service.NewPaymentService(paymentRepo, walletService, db, log)

	managerService := service.NewManagerService(restaurantManagerRepo, restaurantRepo, userRepo, log)

	authHandler := handler.NewAuthHandler(authService, userService)
	userHandler := handler.NewUserHandler(userRepo)
	restaurantHandler := handler.NewRestaurantHandler(restaurantService)
	tableHandler := handler.NewTableHandler(tableRepo)
	bookingHandler := handler.NewBookingHandler(bookingRepo, tableRepo)
	reviewHandler := handler.NewReviewHandler(reviewRepo, restaurantRepo)
	managerHandler := handler.NewManagerHandler(managerService)
	walletHandler := handler.NewWalletHandler(walletService)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)

	concurrentServices := SetupConcurrentServices(
		refreshTokenRepo,
		bookingRepo,
		tableRepo,
		restaurantRepo,
	)

	StartGracefulShutdown(concurrentServices)

	concurrentDemoHandler := handler.NewConcurrentDemoHandler(
		concurrentServices.NotificationSvc,
		concurrentServices.BookingSvc,
	)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	api := r.Group("/api")
	{

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authMiddleware.Authenticate(), authHandler.Logout)
			auth.GET("/me", authMiddleware.Authenticate(), authHandler.GetMe)
		}

		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)

			users.GET("/:id", userHandler.GetUser)
			users.GET("/:id/bookings", bookingHandler.GetUserBookings)
			users.GET("/:id/reviews", reviewHandler.GetUserReviews)
		}

		restaurants := api.Group("/restaurants")
		{
			restaurants.POST("", restaurantHandler.CreateRestaurant)
			restaurants.GET("", restaurantHandler.ListRestaurants)

			restaurants.GET("/:id/tables", tableHandler.GetRestaurantTables)
			restaurants.GET("/:id/bookings", bookingHandler.GetRestaurantBookings)
			restaurants.GET("/:id/reviews", reviewHandler.GetRestaurantReviews)

			restaurants.POST("/:id/managers", managerHandler.AddManager)
			restaurants.GET("/:id/managers", managerHandler.GetManagers)
			restaurants.DELETE("/:id/managers/:user_id", managerHandler.RemoveManager)

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

		wallet := api.Group("/wallet")
		{
			wallet.GET("", walletHandler.GetWallet)
			wallet.GET("/transactions", walletHandler.GetTransactions)
			wallet.POST("/deposit", walletHandler.Deposit)
			wallet.POST("/withdraw", walletHandler.Withdraw)
		}

		payments := api.Group("/payments")
		{
			payments.GET("", paymentHandler.GetUserPayments)
			payments.POST("/wallet", paymentHandler.CreateWalletPayment)
			payments.POST("/halyk", paymentHandler.CreateHalykPayment)
			payments.POST("/kaspi", paymentHandler.CreateKaspiPayment)
			payments.POST("/webhook/halyk", paymentHandler.HalykWebhook)
			payments.POST("/webhook/kaspi", paymentHandler.KaspiWebhook)
			payments.POST("/:id/refund", paymentHandler.RefundPayment)
		}

		demo := api.Group("/demo")
		{
			demo.POST("/bulk-notifications", concurrentDemoHandler.SendBulkNotifications)
			demo.GET("/notification-stats", concurrentDemoHandler.GetNotificationStats)
			demo.POST("/check-availability", concurrentDemoHandler.CheckTablesAvailability)
			demo.GET("/booking-stats/:restaurant_id", concurrentDemoHandler.GetBookingStats)
			demo.POST("/search-tables", concurrentDemoHandler.SearchAvailableTables)
		}
	}

	fmt.Printf("Server running on http://localhost:%s\n", cfg.Port)
	fmt.Printf("Swagger UI: http://localhost:%s/swagger/index.html\n", cfg.Port)
	fmt.Printf("Health check: http://localhost:%s/health\n", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", zap.Error(err))
	}
}
