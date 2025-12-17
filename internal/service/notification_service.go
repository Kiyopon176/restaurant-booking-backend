package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationEmail NotificationType = "email"
	NotificationSMS   NotificationType = "sms"
	NotificationPush  NotificationType = "push"
)

type Notification struct {
	ID        uuid.UUID
	Type      NotificationType
	Recipient string
	Subject   string
	Message   string
	CreatedAt time.Time
}

type NotificationService struct {
	notifications chan Notification
	workers       int
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
	sent          int
	failed        int
}

func NewNotificationService(workers int, bufferSize int) *NotificationService {
	ctx, cancel := context.WithCancel(context.Background())

	ns := &NotificationService{
		notifications: make(chan Notification, bufferSize),
		workers:       workers,
		ctx:           ctx,
		cancel:        cancel,
		sent:          0,
		failed:        0,
	}

	for i := 0; i < workers; i++ {
		ns.wg.Add(1)
		go ns.worker(i)
	}

	log.Printf("Notification service started with %d workers", workers)
	return ns
}

func (ns *NotificationService) worker(id int) {
	defer ns.wg.Done()

	log.Printf("Notification worker %d started", id)

	for {
		select {
		case <-ns.ctx.Done():
			log.Printf("Notification worker %d stopping", id)
			return
		case notification, ok := <-ns.notifications:
			if !ok {
				log.Printf("Notification worker %d: channel closed", id)
				return
			}

			if err := ns.sendNotification(notification); err != nil {
				log.Printf("Worker %d: Failed to send notification %s: %v", id, notification.ID, err)
				ns.incrementFailed()
			} else {
				log.Printf("Worker %d: Successfully sent %s notification to %s",
					id, notification.Type, notification.Recipient)
				ns.incrementSent()
			}
		}
	}
}

func (ns *NotificationService) sendNotification(n Notification) error {

	time.Sleep(100 * time.Millisecond)

	if time.Now().UnixNano()%10 == 0 {
		return fmt.Errorf("simulated network error")
	}

	log.Printf("Sent %s notification to %s: %s", n.Type, n.Recipient, n.Message)
	return nil
}

func (ns *NotificationService) Send(notification Notification) error {
	select {
	case <-ns.ctx.Done():
		return fmt.Errorf("notification service is shutting down")
	case ns.notifications <- notification:
		log.Printf("Notification %s queued for sending", notification.ID)
		return nil
	default:
		return fmt.Errorf("notification queue is full")
	}
}

func (ns *NotificationService) SendEmail(recipient, subject, message string) error {
	notification := Notification{
		ID:        uuid.New(),
		Type:      NotificationEmail,
		Recipient: recipient,
		Subject:   subject,
		Message:   message,
		CreatedAt: time.Now(),
	}
	return ns.Send(notification)
}

func (ns *NotificationService) SendSMS(recipient, message string) error {
	notification := Notification{
		ID:        uuid.New(),
		Type:      NotificationSMS,
		Recipient: recipient,
		Message:   message,
		CreatedAt: time.Now(),
	}
	return ns.Send(notification)
}

func (ns *NotificationService) SendBulk(notifications []Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	log.Printf("Sending %d notifications in bulk", len(notifications))

	go func() {
		for _, n := range notifications {
			if err := ns.Send(n); err != nil {
				log.Printf("Failed to queue notification %s: %v", n.ID, err)
			}
		}
	}()

	return nil
}

func (ns *NotificationService) GetStats() (sent int, failed int) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.sent, ns.failed
}

func (ns *NotificationService) incrementSent() {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.sent++
}

func (ns *NotificationService) incrementFailed() {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.failed++
}

func (ns *NotificationService) Shutdown() {
	log.Println("Shutting down notification service...")

	ns.cancel()

	close(ns.notifications)

	ns.wg.Wait()

	sent, failed := ns.GetStats()
	log.Printf("Notification service shutdown complete. Sent: %d, Failed: %d", sent, failed)
}
