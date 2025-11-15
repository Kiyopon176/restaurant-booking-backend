package handler

import (
	"net/http"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BookingHandler struct {
	bookingRepo repository.BookingRepository
	tableRepo   repository.TableRepository
}

func NewBookingHandler(bookingRepo repository.BookingRepository, tableRepo repository.TableRepository) *BookingHandler {
	return &BookingHandler{
		bookingRepo: bookingRepo,
		tableRepo:   tableRepo,
	}
}

// CreateBooking godoc
// @Summary Создать бронирование
// @Description Создать новое бронирование столика в ресторане
// @Tags bookings
// @Accept json
// @Produce json
// @Param booking body CreateBookingRequest true "Данные бронирования"
// @Success 201 {object} domain.Booking
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse "Столик недоступен"
// @Failure 500 {object} ErrorResponse
// @Router /api/bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	available, err := h.bookingRepo.CheckTableAvailability(
		c.Request.Context(),
		req.TableID,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	if !available {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "table is not available for the selected time"})
		return
	}

	booking := &domain.Booking{
		RestaurantID: req.RestaurantID,
		TableID:      req.TableID,
		UserID:       req.UserID,
		BookingDate:  req.BookingDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		GuestsCount:  req.GuestsCount,
		SpecialNote:  req.SpecialNote,
		Status:       domain.BookingStatusPending,
	}

	if err := h.bookingRepo.Create(c.Request.Context(), booking); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// GetBooking godoc
// @Summary Получить бронирование
// @Description Получить информацию о бронировании по ID
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/bookings/{id} [get]
func (h *BookingHandler) GetBooking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid booking id"})
		return
	}

	booking, err := h.bookingRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "booking not found"})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// GetUserBookings godoc
// @Summary Получить бронирования пользователя
// @Description Получить все бронирования определенного пользователя
// @Tags bookings
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {array} domain.Booking
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/users/{user_id}/bookings [get]
func (h *BookingHandler) GetUserBookings(c *gin.Context) {
	idStr := c.Param("user_id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	bookings, err := h.bookingRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// GetRestaurantBookings godoc
// @Summary Получить бронирования ресторана
// @Description Получить все бронирования ресторана на определенную дату
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param date query string true "Дата в формате YYYY-MM-DD"
// @Success 200 {array} domain.Booking
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/restaurants/{id}/bookings [get]
func (h *BookingHandler) GetRestaurantBookings(c *gin.Context) {
	idStr := c.Param("id")
	restaurantID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	dateStr := c.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
		return
	}

	bookings, err := h.bookingRepo.GetByRestaurantID(c.Request.Context(), restaurantID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// UpdateBookingStatus godoc
// @Summary Обновить статус бронирования
// @Description Обновить статус бронирования (pending, confirmed, cancelled, completed, no_show)
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param status body UpdateBookingStatusRequest true "Новый статус"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/bookings/{id}/status [put]
func (h *BookingHandler) UpdateBookingStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid booking id"})
		return
	}

	booking, err := h.bookingRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "booking not found"})
		return
	}

	var req UpdateBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	booking.Status = req.Status

	if err := h.bookingRepo.Update(c.Request.Context(), booking); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// CancelBooking godoc
// @Summary Отменить бронирование
// @Description Отменить бронирование (установить статус cancelled)
// @Tags bookings
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/bookings/{id}/cancel [put]
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid booking id"})
		return
	}

	booking, err := h.bookingRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "booking not found"})
		return
	}

	booking.Status = domain.BookingStatusCancelled

	if err := h.bookingRepo.Update(c.Request.Context(), booking); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// CheckTableAvailability godoc
// @Summary Проверить доступность столика
// @Description Проверить, доступен ли столик для бронирования на указанное время
// @Tags bookings
// @Accept json
// @Produce json
// @Param table_id query string true "Table ID"
// @Param start_time query string true "Время начала (RFC3339 format)"
// @Param end_time query string true "Время окончания (RFC3339 format)"
// @Success 200 {object} AvailabilityResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/bookings/check-availability [get]
func (h *BookingHandler) CheckTableAvailability(c *gin.Context) {
	tableIDStr := c.Query("table_id")
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid table id"})
		return
	}

	startTimeStr := c.Query("start_time")
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_time format, use RFC3339"})
		return
	}

	endTimeStr := c.Query("end_time")
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid end_time format, use RFC3339"})
		return
	}

	available, err := h.bookingRepo.CheckTableAvailability(c.Request.Context(), tableID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, AvailabilityResponse{
		Available: available,
		TableID:   tableID,
		StartTime: startTime,
		EndTime:   endTime,
	})
}

type CreateBookingRequest struct {
	RestaurantID uuid.UUID `json:"restaurant_id" binding:"required"`
	TableID      uuid.UUID `json:"table_id" binding:"required"`
	UserID       uuid.UUID `json:"user_id" binding:"required"`
	BookingDate  time.Time `json:"booking_date" binding:"required"`
	StartTime    time.Time `json:"start_time" binding:"required"`
	EndTime      time.Time `json:"end_time" binding:"required"`
	GuestsCount  int       `json:"guests_count" binding:"required,min=1"`
	SpecialNote  string    `json:"special_note"`
}

type UpdateBookingStatusRequest struct {
	Status domain.BookingStatus `json:"status" binding:"required"`
}

type AvailabilityResponse struct {
	Available bool      `json:"available"`
	TableID   uuid.UUID `json:"table_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}
