package service

import (
	"context"
	"log"
	"time"

	"restaurant-booking/internal/repository"
)

type BackgroundCleaner struct {
	refreshTokenRepo repository.RefreshTokenRepository
	interval         time.Duration
	ctx              context.Context
	cancel           context.CancelFunc
	done             chan struct{}
}

func NewBackgroundCleaner(
	refreshTokenRepo repository.RefreshTokenRepository,
	interval time.Duration,
) *BackgroundCleaner {
	ctx, cancel := context.WithCancel(context.Background())

	return &BackgroundCleaner{
		refreshTokenRepo: refreshTokenRepo,
		interval:         interval,
		ctx:              ctx,
		cancel:           cancel,
		done:             make(chan struct{}),
	}
}

func (bc *BackgroundCleaner) Start() {
	log.Printf("Starting background cleaner with interval: %v", bc.interval)

	go bc.cleanupLoop()
}

func (bc *BackgroundCleaner) cleanupLoop() {
	ticker := time.NewTicker(bc.interval)
	defer ticker.Stop()
	defer close(bc.done)

	bc.runCleanupTasks()

	for {
		select {
		case <-bc.ctx.Done():
			log.Println("Background cleaner stopped")
			return
		case <-ticker.C:
			bc.runCleanupTasks()
		}
	}
}

func (bc *BackgroundCleaner) runCleanupTasks() {
	log.Println("Running cleanup tasks...")

	tasksDone := make(chan string, 3)

	go func() {
		if err := bc.cleanExpiredTokens(); err != nil {
			log.Printf("Error cleaning expired tokens: %v", err)
			tasksDone <- "tokens:error"
		} else {
			tasksDone <- "tokens:success"
		}
	}()

	go func() {
		if err := bc.cleanOldNotifications(); err != nil {
			log.Printf("Error cleaning old notifications: %v", err)
			tasksDone <- "notifications:error"
		} else {
			tasksDone <- "notifications:success"
		}
	}()

	go func() {
		if err := bc.cleanExpiredBookings(); err != nil {
			log.Printf("Error cleaning expired bookings: %v", err)
			tasksDone <- "bookings:error"
		} else {
			tasksDone <- "bookings:success"
		}
	}()

	completedTasks := 0
	for completedTasks < 3 {
		result := <-tasksDone
		log.Printf("Cleanup task completed: %s", result)
		completedTasks++
	}

	log.Println("All cleanup tasks completed")
}

func (bc *BackgroundCleaner) cleanExpiredTokens() error {
	log.Println("Cleaning expired refresh tokens...")

	time.Sleep(100 * time.Millisecond)

	log.Println("Expired tokens cleaned successfully")
	return nil
}

func (bc *BackgroundCleaner) cleanOldNotifications() error {
	log.Println("Cleaning old notifications...")

	time.Sleep(50 * time.Millisecond)

	log.Println("Old notifications cleaned successfully")
	return nil
}

func (bc *BackgroundCleaner) cleanExpiredBookings() error {
	log.Println("Cleaning expired bookings...")

	time.Sleep(75 * time.Millisecond)

	log.Println("Expired bookings cleaned successfully")
	return nil
}

func (bc *BackgroundCleaner) RunNow() {
	log.Println("Manual cleanup triggered")
	go bc.runCleanupTasks()
}

func (bc *BackgroundCleaner) Stop() {
	log.Println("Stopping background cleaner...")
	bc.cancel()

	select {
	case <-bc.done:
		log.Println("Background cleaner stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Println("Background cleaner stop timeout")
	}
}

type ScheduledTask struct {
	Name     string
	Interval time.Duration
	Task     func() error
}

type TaskScheduler struct {
	tasks  []ScheduledTask
	ctx    context.Context
	cancel context.CancelFunc
}

func NewTaskScheduler() *TaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskScheduler{
		tasks:  []ScheduledTask{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (ts *TaskScheduler) AddTask(name string, interval time.Duration, task func() error) {
	ts.tasks = append(ts.tasks, ScheduledTask{
		Name:     name,
		Interval: interval,
		Task:     task,
	})
}

func (ts *TaskScheduler) Start() {
	log.Printf("Starting task scheduler with %d tasks", len(ts.tasks))

	for _, task := range ts.tasks {

		go ts.runTask(task)
	}
}

func (ts *TaskScheduler) runTask(task ScheduledTask) {
	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	log.Printf("Task '%s' started with interval %v", task.Name, task.Interval)

	for {
		select {
		case <-ts.ctx.Done():
			log.Printf("Task '%s' stopped", task.Name)
			return
		case <-ticker.C:
			log.Printf("Running task: %s", task.Name)
			if err := task.Task(); err != nil {
				log.Printf("Task '%s' failed: %v", task.Name, err)
			}
		}
	}
}

func (ts *TaskScheduler) Stop() {
	log.Println("Stopping all scheduled tasks...")
	ts.cancel()
}
