package handler

import (
	"errors"
	"fmt"
	"net/http"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// @Summary Create wallet payment
// @Description Create payment using wallet balance
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "Payment request"
// @Success 200 {object} domain.Payment
// @Failure 400 {object} ErrorResponse
// @Router /api/payments/wallet [post]
func (h *PaymentHandler) CreateWalletPayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
		return
	}

	var bookingID *uuid.UUID
	if req.BookingID != "" {
		bid, err := uuid.Parse(req.BookingID)
		if err == nil {
			bookingID = &bid
		}
	}

	payment, err := h.paymentService.CreatePayment(
		c.Request.Context(),
		userID,
		req.Amount,
		domain.PaymentMethodWallet,
		bookingID,
	)

	if err != nil {
		if errors.Is(err, service.ErrInsufficientBalance) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "insufficient balance"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// @Summary Create Halyk Bank payment
// @Description Create payment via Halyk Bank (returns payment URL)
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "Payment request"
// @Success 200 {object} PaymentWithURLResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/payments/halyk [post]
func (h *PaymentHandler) CreateHalykPayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
		return
	}

	var bookingID *uuid.UUID
	if req.BookingID != "" {
		bid, err := uuid.Parse(req.BookingID)
		if err == nil {
			bookingID = &bid
		}
	}

	payment, err := h.paymentService.CreatePayment(
		c.Request.Context(),
		userID,
		req.Amount,
		domain.PaymentMethodHalyk,
		bookingID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	url, err := h.paymentService.CreateHalykPayment(c.Request.Context(), payment.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, PaymentWithURLResponse{
		Payment:            payment,
		ExternalPaymentURL: url,
	})
}

// @Summary Create Kaspi payment
// @Description Create payment via Kaspi (returns payment URL)
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "Payment request"
// @Success 200 {object} PaymentWithURLResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/payments/kaspi [post]
func (h *PaymentHandler) CreateKaspiPayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
		return
	}

	var bookingID *uuid.UUID
	if req.BookingID != "" {
		bid, err := uuid.Parse(req.BookingID)
		if err == nil {
			bookingID = &bid
		}
	}

	payment, err := h.paymentService.CreatePayment(
		c.Request.Context(),
		userID,
		req.Amount,
		domain.PaymentMethodKaspi,
		bookingID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	url, err := h.paymentService.CreateKaspiPayment(c.Request.Context(), payment.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, PaymentWithURLResponse{
		Payment:            payment,
		ExternalPaymentURL: url,
	})
}

// @Summary Halyk Bank webhook
// @Description Webhook endpoint for Halyk Bank payment notifications
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body WebhookRequest true "Webhook request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/payments/webhook/halyk [post]
func (h *PaymentHandler) HalykWebhook(c *gin.Context) {
	var req WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	success := req.Status == "success" || req.Status == "completed" || req.Status == "paid"

	if err := h.paymentService.ProcessExternalPaymentCallback(
		c.Request.Context(),
		req.ExternalPaymentID,
		success,
	); err != nil {

	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "ok"})
}

// @Summary Kaspi webhook
// @Description Webhook endpoint for Kaspi payment notifications
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body WebhookRequest true "Webhook request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/payments/webhook/kaspi [post]
func (h *PaymentHandler) KaspiWebhook(c *gin.Context) {
	h.HalykWebhook(c)
}

// @Summary Refund payment
// @Description Refund a completed payment
// @Tags Payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/payments/{id}/refund [post]
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid payment id"})
		return
	}

	if err := h.paymentService.RefundPayment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "refunded"})
}

// @Summary Get user payments
// @Description Get payment history for a user
// @Tags Payments
// @Produce json
// @Param user_id query string true "User ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Payment
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/payments [get]
func (h *PaymentHandler) GetUserPayments(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user_id is required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
		return
	}

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	payments, err := h.paymentService.GetPaymentsByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, payments)
}

type CreatePaymentRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	Amount    int    `json:"amount" binding:"required,min=1"`
	BookingID string `json:"booking_id"`
}

type WebhookRequest struct {
	ExternalPaymentID string `json:"external_payment_id" binding:"required"`
	Status            string `json:"status" binding:"required"`
}

type PaymentWithURLResponse struct {
	Payment            *domain.Payment `json:"payment"`
	ExternalPaymentURL string          `json:"external_payment_url"`
}
