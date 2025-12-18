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
	NotificationSvc *service.NotificationService
	BookingSvc      *service.BookingService
}

func SetupConcurrentServices(
	refreshTokenRepo repository.RefreshTokenRepository,
	bookingRepo repository.BookingRepository,
	tableRepo repository.TableRepository,
	restaurantRepo repository.RestaurantRepository,
) *ConcurrentServices {
	log.Println("Setting up concurrent services...")

	notificationSvc := service.NewNotificationService(5, 100)

	bookingSvc := service.NewBookingService(
		bookingRepo,
		tableRepo,
		restaurantRepo,
		notificationSvc,
	)

	log.Println("All concurrent services initialized successfully")

	return &ConcurrentServices{
		NotificationSvc: notificationSvc,
		BookingSvc:      bookingSvc,
	}
}

func StartGracefulShutdown(services *ConcurrentServices) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("\nReceived signal: %v. Starting graceful shutdown...", sig)

		log.Println("Stopping notification service...")
		services.NotificationSvc.Shutdown()

		log.Println("All services stopped gracefully")
		os.Exit(0)
	}()
}

func DemoConcurrentFeatures(services *ConcurrentServices) {
	log.Println("\nDemonstrating concurrent features...")

	go func() {
		time.Sleep(2 * time.Second)
		log.Println("\nDemo 1: Sending bulk notifications...")

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

		log.Println("Demo 1: Notifications queued")
	}()

	go func() {
		time.Sleep(5 * time.Second)
		sent, failed := services.NotificationSvc.GetStats()
		log.Printf("\nDemo 2: Current Stats - Sent: %d, Failed: %d", sent, failed)
	}()
}
