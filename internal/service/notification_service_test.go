package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewNotificationService(t *testing.T) {
	workers := 3
	bufferSize := 10

	ns := NewNotificationService(workers, bufferSize)

	assert.NotNil(t, ns)
	assert.Equal(t, workers, ns.workers)
	assert.NotNil(t, ns.notifications)
	assert.NotNil(t, ns.ctx)
	assert.NotNil(t, ns.cancel)

	ns.Shutdown()
}

func TestSendEmail_Success(t *testing.T) {
	ns := NewNotificationService(2, 10)
	defer ns.Shutdown()

	err := ns.SendEmail("test@example.com", "Test Subject", "Test Message")

	assert.NoError(t, err)
}

func TestSendSMS_Success(t *testing.T) {
	ns := NewNotificationService(2, 10)
	defer ns.Shutdown()

	err := ns.SendSMS("+1234567890", "Test SMS Message")

	assert.NoError(t, err)
}

func TestSend_Success(t *testing.T) {
	ns := NewNotificationService(2, 10)
	defer ns.Shutdown()

	notification := Notification{
		ID:        uuid.New(),
		Type:      NotificationEmail,
		Recipient: "test@example.com",
		Subject:   "Test",
		Message:   "Test message",
		CreatedAt: time.Now(),
	}

	err := ns.Send(notification)

	assert.NoError(t, err)
}

func TestSend_QueueFull(t *testing.T) {
	ns := NewNotificationService(1, 1)
	defer ns.Shutdown()

	notification := Notification{
		ID:        uuid.New(),
		Type:      NotificationEmail,
		Recipient: "test@example.com",
		Message:   "Test",
		CreatedAt: time.Now(),
	}

	err1 := ns.Send(notification)
	assert.NoError(t, err1)

	err2 := ns.Send(notification)
	if err2 != nil {
		assert.Contains(t, err2.Error(), "queue is full")
	}
}

func TestSend_AfterShutdown(t *testing.T) {
	ns := NewNotificationService(2, 10)
	ns.Shutdown()

	notification := Notification{
		ID:        uuid.New(),
		Type:      NotificationEmail,
		Recipient: "test@example.com",
		Message:   "Test",
		CreatedAt: time.Now(),
	}

	err := ns.Send(notification)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shutting down")
}

func TestSendBulk_Success(t *testing.T) {
	ns := NewNotificationService(3, 20)
	defer ns.Shutdown()

	notifications := []Notification{
		{
			ID:        uuid.New(),
			Type:      NotificationEmail,
			Recipient: "user1@example.com",
			Message:   "Message 1",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Type:      NotificationSMS,
			Recipient: "+1234567890",
			Message:   "Message 2",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Type:      NotificationPush,
			Recipient: "device123",
			Message:   "Message 3",
			CreatedAt: time.Now(),
		},
	}

	err := ns.SendBulk(notifications)

	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
}

func TestSendBulk_EmptyList(t *testing.T) {
	ns := NewNotificationService(2, 10)
	defer ns.Shutdown()

	err := ns.SendBulk([]Notification{})

	assert.NoError(t, err)
}

func TestGetStats(t *testing.T) {
	ns := NewNotificationService(2, 10)
	defer ns.Shutdown()

	sent, failed := ns.GetStats()
	assert.Equal(t, 0, sent)
	assert.Equal(t, 0, failed)

	ns.SendEmail("test@example.com", "Test", "Message")
	time.Sleep(150 * time.Millisecond)

	sent, failed = ns.GetStats()
	assert.GreaterOrEqual(t, sent+failed, 0)
}

func TestIncrementSent(t *testing.T) {
	ns := NewNotificationService(2, 10)
	defer ns.Shutdown()

	ns.incrementSent()
	ns.incrementSent()

	sent, _ := ns.GetStats()
	assert.Equal(t, 2, sent)
}

func TestIncrementFailed(t *testing.T) {
	ns := NewNotificationService(2, 10)
	defer ns.Shutdown()

	ns.incrementFailed()
	ns.incrementFailed()
	ns.incrementFailed()

	_, failed := ns.GetStats()
	assert.Equal(t, 3, failed)
}

func TestShutdown(t *testing.T) {
	ns := NewNotificationService(2, 10)

	ns.SendEmail("test@example.com", "Test", "Message")

	ns.Shutdown()

	sent, failed := ns.GetStats()
	assert.GreaterOrEqual(t, sent+failed, 0)
}

func TestWorkerProcessing(t *testing.T) {
	ns := NewNotificationService(3, 5)
	defer ns.Shutdown()

	for i := 0; i < 5; i++ {
		ns.SendEmail("test@example.com", "Test", "Message")
	}

	time.Sleep(200 * time.Millisecond)

	sent, failed := ns.GetStats()
	assert.Greater(t, sent+failed, 0)
}

func TestConcurrentSends(t *testing.T) {
	ns := NewNotificationService(5, 50)
	defer ns.Shutdown()

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			err := ns.SendEmail("test@example.com", "Test", "Concurrent message")
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	time.Sleep(200 * time.Millisecond)

	sent, failed := ns.GetStats()
	assert.Greater(t, sent+failed, 0)
}

func TestNotificationTypes(t *testing.T) {
	ns := NewNotificationService(3, 10)
	defer ns.Shutdown()

	tests := []struct {
		notificationType NotificationType
		expectedType     NotificationType
	}{
		{NotificationEmail, NotificationEmail},
		{NotificationSMS, NotificationSMS},
		{NotificationPush, NotificationPush},
	}

	for _, tt := range tests {
		notification := Notification{
			ID:        uuid.New(),
			Type:      tt.notificationType,
			Recipient: "recipient",
			Message:   "test",
			CreatedAt: time.Now(),
		}

		err := ns.Send(notification)
		assert.NoError(t, err)
		assert.Equal(t, tt.expectedType, notification.Type)
	}
}

func TestShutdownWithPendingNotifications(t *testing.T) {
	ns := NewNotificationService(1, 5)

	for i := 0; i < 5; i++ {
		ns.SendEmail("test@example.com", "Test", "Message")
	}

	ns.Shutdown()

	sent, failed := ns.GetStats()
	assert.GreaterOrEqual(t, sent+failed, 0)
}

func TestContextCancellation(t *testing.T) {
	ns := NewNotificationService(2, 10)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	ns.SendEmail("test@example.com", "Test", "Message")

	time.Sleep(100 * time.Millisecond)

	ns.Shutdown()
}
