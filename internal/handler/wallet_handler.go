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

// @Summary Get user wallet
// @Description Get or create wallet for a user
// @Tags Wallet
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {object} domain.Wallet
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/wallet [get]
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

// @Summary Deposit to wallet
// @Description Add funds to user's wallet
// @Tags Wallet
// @Accept json
// @Produce json
// @Param request body DepositRequest true "Deposit request"
// @Success 200 {object} domain.Wallet
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/wallet/deposit [post]
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

// @Summary Withdraw from wallet
// @Description Withdraw funds from user's wallet
// @Tags Wallet
// @Accept json
// @Produce json
// @Param request body WithdrawRequest true "Withdraw request"
// @Success 200 {object} domain.Wallet
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/wallet/withdraw [post]
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

// @Summary Get wallet transactions
// @Description Get transaction history for a user's wallet
// @Tags Wallet
// @Produce json
// @Param user_id query string true "User ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.WalletTransaction
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/wallet/transactions [get]
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
