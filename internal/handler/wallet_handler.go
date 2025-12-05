package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/service"
)

type WalletHandler struct {
	Service *service.WalletService
}

func NewWalletHandler(s *service.WalletService) *WalletHandler {
	return &WalletHandler{Service: s}
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
	// user id from auth middleware — here we accept query param for demo
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_user_id"})
		return
	}
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id"})
		return
	}
	w, err := h.Service.GetOrCreateWallet(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wallet": w})
}

func (h *WalletHandler) Deposit(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id"`
		Amount      int    `json:"amount"`
		Description string `json:"description"`
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
	if err := h.Service.Deposit(uid, req.Amount, req.Description); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, _ := h.Service.GetOrCreateWallet(uid)
	c.JSON(http.StatusOK, gin.H{"wallet": w})
}

func (h *WalletHandler) Withdraw(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id"`
		Amount      int    `json:"amount"`
		Description string `json:"description"`
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
	if err := h.Service.Withdraw(uid, req.Amount, req.Description); err != nil {
		if err.Error() == "insufficient_balance" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient_balance"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	w, _ := h.Service.GetOrCreateWallet(uid)
	c.JSON(http.StatusOK, gin.H{"wallet": w})
}
