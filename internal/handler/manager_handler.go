package handler

import (
	"errors"
	"net/http"
	"restaurant-booking/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ManagerHandler struct {
	managerService service.ManagerService
}

func NewManagerHandler(managerService service.ManagerService) *ManagerHandler {
	return &ManagerHandler{managerService: managerService}
}

func (h *ManagerHandler) AddManager(c *gin.Context) {

	restaurantIDStr := c.Param("id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	ownerIDStr := c.Query("owner_id")
	if ownerIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "owner_id is required"})
		return
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid owner id"})
		return
	}

	var req AddManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := service.AddManagerRequest{
		UserID: req.UserID,
	}

	manager, err := h.managerService.AddManager(c.Request.Context(), restaurantID, ownerID, serviceReq)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "restaurant not found"})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized: not the owner"})
		case errors.Is(err, service.ErrManagerAlreadyExists):
			c.JSON(http.StatusConflict, ErrorResponse{Error: "user is already a manager"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, manager)
}

func (h *ManagerHandler) RemoveManager(c *gin.Context) {

	restaurantIDStr := c.Param("id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	ownerIDStr := c.Query("owner_id")
	if ownerIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "owner_id is required"})
		return
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid owner id"})
		return
	}

	err = h.managerService.RemoveManager(c.Request.Context(), restaurantID, ownerID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "restaurant not found"})
		case errors.Is(err, service.ErrManagerNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "manager not found"})
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized: not the owner"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ManagerHandler) GetManagers(c *gin.Context) {

	restaurantIDStr := c.Param("id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	managers, err := h.managerService.GetManagers(c.Request.Context(), restaurantID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "restaurant not found"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, managers)
}

type AddManagerRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}
