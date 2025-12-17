package handler

import (
	"context"
	"net/http"
	"restaurant-booking/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ConcurrentDemoHandler struct {
	notificationSvc *service.NotificationService
	bookingSvc      *service.BookingService
}

func NewConcurrentDemoHandler(
	notificationSvc *service.NotificationService,
	bookingSvc *service.BookingService,
) *ConcurrentDemoHandler {
	return &ConcurrentDemoHandler{
		notificationSvc: notificationSvc,
		bookingSvc:      bookingSvc,
	}
}

func (h *ConcurrentDemoHandler) SendBulkNotifications(c *gin.Context) {
	var req BulkNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	notifications := make([]service.Notification, len(req.Recipients))
	for i, recipient := range req.Recipients {
		notifications[i] = service.Notification{
			ID:        uuid.New(),
			Type:      service.NotificationEmail,
			Recipient: recipient,
			Subject:   req.Subject,
			Message:   req.Message,
			CreatedAt: time.Now(),
		}
	}

	if err := h.notificationSvc.SendBulk(notifications); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Notifications queued for sending",
	})
}

func (h *ConcurrentDemoHandler) GetNotificationStats(c *gin.Context) {
	sent, failed := h.notificationSvc.GetStats()

	c.JSON(http.StatusOK, NotificationStatsResponse{
		Sent:   sent,
		Failed: failed,
	})
}

func (h *ConcurrentDemoHandler) CheckTablesAvailability(c *gin.Context) {
	var req CheckAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	results := h.bookingSvc.CheckMultipleTablesAvailability(
		ctx,
		req.TableIDs,
		req.StartTime,
		req.EndTime,
	)

	c.JSON(http.StatusOK, ConcurrentAvailabilityResponse{
		Availability: results,
	})
}

func (h *ConcurrentDemoHandler) GetBookingStats(c *gin.Context) {
	restaurantIDStr := c.Param("restaurant_id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid restaurant ID"})
		return
	}

	ctx := c.Request.Context()
	stats, err := h.bookingSvc.GetBookingStatistics(ctx, restaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, BookingStatsResponse{
		RestaurantID: restaurantID,
		Stats:        stats,
	})
}

func (h *ConcurrentDemoHandler) SearchAvailableTables(c *gin.Context) {
	var req SearchTablesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	results := h.bookingSvc.SearchAvailableTablesParallel(
		ctx,
		req.RestaurantIDs,
		req.StartTime,
		req.EndTime,
		req.GuestCount,
	)

	c.JSON(http.StatusOK, SearchTablesResponse{
		Results: results,
	})
}

type BulkNotificationRequest struct {
	Recipients []string `json:"recipients" binding:"required"`
	Subject    string   `json:"subject" binding:"required"`
	Message    string   `json:"message" binding:"required"`
}

type NotificationStatsResponse struct {
	Sent   int `json:"sent"`
	Failed int `json:"failed"`
}

type CheckAvailabilityRequest struct {
	TableIDs  []uuid.UUID `json:"table_ids" binding:"required"`
	StartTime time.Time   `json:"start_time" binding:"required"`
	EndTime   time.Time   `json:"end_time" binding:"required"`
}

type ConcurrentAvailabilityResponse struct {
	Availability map[uuid.UUID]bool `json:"availability"`
}

type BookingStatsResponse struct {
	RestaurantID uuid.UUID      `json:"restaurant_id"`
	Stats        map[string]int `json:"stats"`
}

type SearchTablesRequest struct {
	RestaurantIDs []uuid.UUID `json:"restaurant_ids" binding:"required"`
	StartTime     time.Time   `json:"start_time" binding:"required"`
	EndTime       time.Time   `json:"end_time" binding:"required"`
	GuestCount    int         `json:"guest_count" binding:"required"`
}

type SearchTablesResponse struct {
	Results map[uuid.UUID][]uuid.UUID `json:"results"`
}
