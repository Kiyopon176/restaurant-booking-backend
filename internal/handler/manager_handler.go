package handler

import (
	"errors"
	"net/http"
	"restaurant-booking/internal/domain"
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

// AddManager godoc
// @Summary Добавить менеджера к ресторану
// @Description Добавить пользователя в качестве менеджера ресторана (только владелец)
// @Tags managers
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param owner_id query string true "Owner ID"
// @Param request body AddManagerRequest true "Данные менеджера"
// @Success 201 {object} domain.RestaurantManager
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/restaurants/{id}/managers [post]
func (h *ManagerHandler) AddManager(c *gin.Context) {
	// Get restaurant ID from path
	restaurantIDStr := c.Param("id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	// Get owner ID from query (in real app, this would come from JWT token)
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

	// Parse request body
	var req AddManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Call service
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

// RemoveManager godoc
// @Summary Удалить менеджера из ресторана
// @Description Удалить менеджера из ресторана (только владелец)
// @Tags managers
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param user_id path string true "User ID"
// @Param owner_id query string true "Owner ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/restaurants/{id}/managers/{user_id} [delete]
func (h *ManagerHandler) RemoveManager(c *gin.Context) {
	// Get restaurant ID from path
	restaurantIDStr := c.Param("id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	// Get user ID from path
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	// Get owner ID from query (in real app, this would come from JWT token)
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

	// Call service
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

// GetManagers godoc
// @Summary Получить список менеджеров ресторана
// @Description Получить всех менеджеров ресторана
// @Tags managers
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Success 200 {array} domain.RestaurantManager
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/restaurants/{id}/managers [get]
func (h *ManagerHandler) GetManagers(c *gin.Context) {
	// Get restaurant ID from path
	restaurantIDStr := c.Param("id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	// Call service
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

// Request types
type AddManagerRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}
