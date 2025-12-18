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

// @Summary Send bulk notifications
// @Description Send notifications to multiple recipients concurrently
// @Tags Demo - Concurrent Features
// @Accept json
// @Produce json
// @Param request body BulkNotificationRequest true "Bulk notification request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/demo/bulk-notifications [post]
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

// @Summary Get notification statistics
// @Description Get statistics of sent and failed notifications
// @Tags Demo - Concurrent Features
// @Produce json
// @Success 200 {object} NotificationStatsResponse
// @Router /api/demo/notification-stats [get]
func (h *ConcurrentDemoHandler) GetNotificationStats(c *gin.Context) {
	sent, failed := h.notificationSvc.GetStats()

	c.JSON(http.StatusOK, NotificationStatsResponse{
		Sent:   sent,
		Failed: failed,
	})
}

// @Summary Check multiple tables availability
// @Description Check availability of multiple tables concurrently
// @Tags Demo - Concurrent Features
// @Accept json
// @Produce json
// @Param request body CheckAvailabilityRequest true "Check availability request"
// @Success 200 {object} ConcurrentAvailabilityResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/demo/check-availability [post]
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

// @Summary Get booking statistics
// @Description Get booking statistics for a restaurant (calculated in parallel)
// @Tags Demo - Concurrent Features
// @Produce json
// @Param restaurant_id path string true "Restaurant ID"
// @Success 200 {object} BookingStatsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/demo/booking-stats/{restaurant_id} [get]
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

// @Summary Search available tables across restaurants
// @Description Search for available tables across multiple restaurants in parallel
// @Tags Demo - Concurrent Features
// @Accept json
// @Produce json
// @Param request body SearchTablesRequest true "Search tables request"
// @Success 200 {object} SearchTablesResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/demo/search-tables [post]
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
