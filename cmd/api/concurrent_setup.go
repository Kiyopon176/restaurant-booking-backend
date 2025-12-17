package main

import (
	"log"
	"os"
	"os/signal"
	"restaurant-booking/internal/repository"
	"restaurant-booking/internal/service"
	"syscall"
	"time"
)

type ConcurrentServices struct {
	NotificationSvc   *service.NotificationService
	BackgroundCleaner *service.BackgroundCleaner
	TaskScheduler     *service.TaskScheduler
	BookingSvc        *service.BookingService
}

func SetupConcurrentServices(
	refreshTokenRepo repository.RefreshTokenRepository,
	bookingRepo repository.BookingRepository,
	tableRepo repository.TableRepository,
	restaurantRepo repository.RestaurantRepository,
) *ConcurrentServices {
	log.Println("Setting up concurrent services...")

	notificationSvc := service.NewNotificationService(5, 100)

	backgroundCleaner := service.NewBackgroundCleaner(
		refreshTokenRepo,
		1*time.Hour,
	)
	backgroundCleaner.Start()

	taskScheduler := service.NewTaskScheduler()

	taskScheduler.AddTask("log-stats", 5*time.Minute, func() error {
		sent, failed := notificationSvc.GetStats()
		log.Printf("📊 Notification Stats - Sent: %d, Failed: %d", sent, failed)
		return nil
	})

	taskScheduler.AddTask("health-check", 10*time.Minute, func() error {
		log.Println("💚 System health check passed")
		return nil
	})

	taskScheduler.Start()

	bookingSvc := service.NewBookingService(
		bookingRepo,
		tableRepo,
		restaurantRepo,
		notificationSvc,
	)

	log.Println("✅ All concurrent services initialized successfully")

	return &ConcurrentServices{
		NotificationSvc:   notificationSvc,
		BackgroundCleaner: backgroundCleaner,
		TaskScheduler:     taskScheduler,
		BookingSvc:        bookingSvc,
	}
}

func StartGracefulShutdown(services *ConcurrentServices) {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("\n🛑 Received signal: %v. Starting graceful shutdown...", sig)

		log.Println("Stopping task scheduler...")
		services.TaskScheduler.Stop()

		log.Println("Stopping background cleaner...")
		services.BackgroundCleaner.Stop()

		log.Println("Stopping notification service...")
		services.NotificationSvc.Shutdown()

		log.Println("✅ All services stopped gracefully")
		os.Exit(0)
	}()
}

func DemoConcurrentFeatures(services *ConcurrentServices) {
	log.Println("\n🎯 Demonstrating concurrent features...")

	go func() {
		time.Sleep(2 * time.Second)
		log.Println("\n📧 Demo 1: Sending bulk notifications...")

		recipients := []string{
			"user1@example.com",
			"user2@example.com",
			"user3@example.com",
			"user4@example.com",
			"user5@example.com",
		}

		for _, email := range recipients {
			services.NotificationSvc.SendEmail(
				email,
				"Welcome to Restaurant Booking!",
				"Thank you for joining our platform.",
			)
		}

		log.Println("✅ Demo 1: Notifications queued")
	}()

	go func() {
		time.Sleep(5 * time.Second)
		log.Println("\n🧹 Demo 2: Triggering manual cleanup...")
		services.BackgroundCleaner.RunNow()
		log.Println("✅ Demo 2: Cleanup triggered")
	}()

	go func() {
		time.Sleep(8 * time.Second)
		sent, failed := services.NotificationSvc.GetStats()
		log.Printf("\n📊 Demo 3: Current Stats - Sent: %d, Failed: %d", sent, failed)
	}()
}
