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

type RestaurantHandler struct {
	restaurantService service.RestaurantService
}

func NewRestaurantHandler(restaurantService service.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{restaurantService: restaurantService}
}

// CreateRestaurant godoc
// @Summary Создать ресторан
// @Description Создать новый ресторан
// @Tags restaurants
// @Accept json
// @Produce json
// @Param restaurant body CreateRestaurantRequest true "Данные ресторана"
// @Success 201 {object} domain.Restaurant
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/restaurants [post]
func (h *RestaurantHandler) CreateRestaurant(c *gin.Context) {
	var req CreateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := service.CreateRestaurantRequest{
		Name:                req.Name,
		Address:             req.Address,
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		Description:         req.Description,
		Phone:               req.Phone,
		Instagram:           req.Instagram,
		Website:             req.Website,
		CuisineType:         req.CuisineType,
		AveragePrice:        req.AveragePrice,
		MaxCombinableTables: req.MaxCombinableTables,
		WorkingHours:        req.WorkingHours,
	}

	restaurant, err := h.restaurantService.CreateRestaurant(c.Request.Context(), req.OwnerID, serviceReq)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRestaurantName):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "restaurant name cannot be empty"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, restaurant)
}

// GetRestaurant godoc
// @Summary Получить ресторан
// @Description Получить информацию о ресторане по ID
// @Tags restaurants
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Success 200 {object} domain.Restaurant
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/restaurants/{id} [get]
func (h *RestaurantHandler) GetRestaurant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	restaurant, err := h.restaurantService.GetRestaurant(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "restaurant not found"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, restaurant)
}

// ListRestaurants godoc
// @Summary Список ресторанов
// @Description Получить список всех ресторанов с пагинацией
// @Tags restaurants
// @Accept json
// @Produce json
// @Param limit query int false "Количество записей" default(10)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} domain.Restaurant
// @Failure 500 {object} ErrorResponse
// @Router /api/restaurants [get]
func (h *RestaurantHandler) ListRestaurants(c *gin.Context) {
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	restaurants, err := h.restaurantService.GetRestaurants(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, restaurants)
}

// UpdateRestaurant godoc
// @Summary Обновить ресторан
// @Description Обновить информацию о ресторане
// @Tags restaurants
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param owner_id query string true "Owner ID"
// @Param restaurant body UpdateRestaurantRequest true "Данные для обновления"
// @Success 200 {object} domain.Restaurant
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/restaurants/{id} [put]
func (h *RestaurantHandler) UpdateRestaurant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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

	var req UpdateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	serviceReq := service.UpdateRestaurantRequest{
		Name:        req.Name,
		Address:     req.Address,
		Description: req.Description,
		Phone:       req.Phone,
		IsActive:    req.IsActive,
	}

	restaurant, err := h.restaurantService.UpdateRestaurant(c.Request.Context(), id, ownerID, serviceReq)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "restaurant not found"})
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized: not the owner"})
		case errors.Is(err, service.ErrInvalidRestaurantName):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "restaurant name cannot be empty"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, restaurant)
}

// DeleteRestaurant godoc
// @Summary Удалить ресторан (soft delete)
// @Description Удалить ресторан по ID (устанавливает is_active = false)
// @Tags restaurants
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param owner_id query string true "Owner ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/restaurants/{id} [delete]
func (h *RestaurantHandler) DeleteRestaurant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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

	err = h.restaurantService.DeleteRestaurant(c.Request.Context(), id, ownerID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "restaurant not found"})
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized: not the owner"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// AddImage godoc
// @Summary Добавить изображение к ресторану
// @Description Загрузить изображение для ресторана (mock upload)
// @Tags restaurants
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param owner_id query string true "Owner ID"
// @Param image formData file true "Файл изображения"
// @Param is_main formData boolean false "Является ли главным изображением" default(false)
// @Success 201 {object} domain.RestaurantImage
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/restaurants/{id}/images [post]
func (h *RestaurantHandler) AddImage(c *gin.Context) {
	idStr := c.Param("id")
	restaurantID, err := uuid.Parse(idStr)
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

	// Get file from form
	_, err = c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "image file is required"})
		return
	}

	// Get is_main flag
	isMain := c.DefaultPostForm("is_main", "false") == "true"

	// Mock upload - in real app, upload to Cloudinary
	mockURL := "https://via.placeholder.com/400"
	mockPublicID := "mock-id"

	serviceReq := service.AddImageRequest{
		CloudinaryURL:      mockURL,
		CloudinaryPublicID: mockPublicID,
		IsMain:             isMain,
	}

	image, err := h.restaurantService.AddImage(c.Request.Context(), restaurantID, ownerID, serviceReq)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "restaurant not found"})
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized: not the owner"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, image)
}

// DeleteImage godoc
// @Summary Удалить изображение ресторана
// @Description Удалить изображение ресторана по ID
// @Tags restaurants
// @Accept json
// @Produce json
// @Param id path string true "Restaurant ID"
// @Param image_id path string true "Image ID"
// @Param owner_id query string true "Owner ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/restaurants/{id}/images/{image_id} [delete]
func (h *RestaurantHandler) DeleteImage(c *gin.Context) {
	idStr := c.Param("id")
	restaurantID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	imageIDStr := c.Param("image_id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid image id"})
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

	err = h.restaurantService.DeleteImage(c.Request.Context(), imageID, restaurantID, ownerID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRestaurantNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "restaurant not found"})
		case errors.Is(err, service.ErrImageNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "image not found"})
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized: not the owner"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

type CreateRestaurantRequest struct {
	OwnerID             uuid.UUID           `json:"owner_id" binding:"required"`
	Name                string              `json:"name" binding:"required"`
	Address             string              `json:"address" binding:"required"`
	Latitude            *float64            `json:"latitude"`
	Longitude           *float64            `json:"longitude"`
	Description         string              `json:"description"`
	Phone               string              `json:"phone" binding:"required"`
	Instagram           *string             `json:"instagram"`
	Website             *string             `json:"website"`
	CuisineType         domain.CuisineType  `json:"cuisine_type" binding:"required"`
	AveragePrice        int                 `json:"average_price" binding:"required"`
	MaxCombinableTables int                 `json:"max_combinable_tables" binding:"required"`
	WorkingHours        domain.WorkingHours `json:"working_hours" binding:"required"`
}

type UpdateRestaurantRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Phone       *string `json:"phone"`
	Address     *string `json:"address"`
	IsActive    *bool   `json:"is_active"`
}
