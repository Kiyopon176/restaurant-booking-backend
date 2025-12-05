package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
	"github.com/Kiyopon176/restaurant-booking-backend/internal/service"
)

type PaymentHandler struct {
	Service *service.PaymentService
}

func NewPaymentHandler(s *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{Service: s}
}

func (h *PaymentHandler) CreateWalletPayment(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id"`
		Amount    int    `json:"amount"`
		BookingID string `json:"booking_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id"})
		return
	}
	var bid *uuid.UUID
	if req.BookingID != "" {
		tmp, err := uuid.Parse(req.BookingID)
		if err == nil {
			bid = &tmp
		}
	}
	p, err := h.Service.CreatePayment(uid, req.Amount, domain.PaymentMethodWallet, bid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payment": p})
}

func (h *PaymentHandler) CreateHalykPayment(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id"`
		Amount    int    `json:"amount"`
		BookingID string `json:"booking_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id"})
		return
	}
	var bid *uuid.UUID
	if req.BookingID != "" {
		tmp, err := uuid.Parse(req.BookingID)
		if err == nil {
			bid = &tmp
		}
	}
	p, err := h.Service.CreatePayment(uid, req.Amount, domain.PaymentMethodHalyk, bid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	url, err := h.Service.CreateHalykPayment(p.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payment": p, "external_payment_url": url})
}

func (h *PaymentHandler) HalykWebhook(c *gin.Context) {
	var req struct {
		ExternalPaymentID string `json:"external_payment_id"`
		Status            string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	success := req.Status == "success" || req.Status == "completed" || req.Status == "paid"
	if err := h.Service.ProcessExternalPaymentCallback(req.ExternalPaymentID, success); err != nil {
		// log but return ok to provider (idempotent & safe)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *PaymentHandler) KaspiWebhook(c *gin.Context) {
	h.HalykWebhook(c)
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}
	if err := h.Service.RefundPayment(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "refunded"})
}
