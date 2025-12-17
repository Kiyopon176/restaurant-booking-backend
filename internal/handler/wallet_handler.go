package handler

import (
	"errors"
	"fmt"
	"net/http"
	"restaurant-booking/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletService service.WalletService
}

func NewWalletHandler(walletService service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
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

	wallet, err := h.walletService.GetOrCreateWallet(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

func (h *WalletHandler) Deposit(c *gin.Context) {
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
		return
	}

	if err := h.walletService.Deposit(c.Request.Context(), userID, req.Amount, req.Description); err != nil {
		if errors.Is(err, service.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	wallet, _ := h.walletService.GetOrCreateWallet(c.Request.Context(), userID)
	c.JSON(http.StatusOK, wallet)
}

func (h *WalletHandler) Withdraw(c *gin.Context) {
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
		return
	}

	if err := h.walletService.Withdraw(c.Request.Context(), userID, req.Amount, req.Description); err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientBalance):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "insufficient balance"})
		case errors.Is(err, service.ErrInvalidAmount):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	wallet, _ := h.walletService.GetOrCreateWallet(c.Request.Context(), userID)
	c.JSON(http.StatusOK, wallet)
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
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

	transactions, err := h.walletService.GetTransactions(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

type DepositRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Amount      int    `json:"amount" binding:"required,min=1"`
	Description string `json:"description"`
}

type WithdrawRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Amount      int    `json:"amount" binding:"required,min=1"`
	Description string `json:"description"`
}
