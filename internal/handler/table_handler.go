package handler

import (
	"fmt"
	"net/http"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TableHandler struct {
	tableRepo repository.TableRepository
}

func NewTableHandler(tableRepo repository.TableRepository) *TableHandler {
	return &TableHandler{tableRepo: tableRepo}
}

func (h *TableHandler) CreateTable(c *gin.Context) {
	var req CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	table := &domain.Table{
		RestaurantID: req.RestaurantID,
		TableNumber:  req.TableNumber,
		MinCapacity:  req.MinCapacity,
		MaxCapacity:  req.MaxCapacity,
		LocationType: req.LocationType,
		XPosition:    req.XPosition,
		YPosition:    req.YPosition,
	}

	if err := h.tableRepo.Create(c.Request.Context(), table); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, table)
}

func (h *TableHandler) GetTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid table id"})
		return
	}

	table, err := h.tableRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "table not found"})
		return
	}

	c.JSON(http.StatusOK, table)
}

func (h *TableHandler) GetRestaurantTables(c *gin.Context) {
	idStr := c.Param("id")
	restaurantID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	tables, err := h.tableRepo.GetByRestaurantID(c.Request.Context(), restaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tables)
}

func (h *TableHandler) GetAvailableTables(c *gin.Context) {
	restaurantIDStr := c.Query("restaurant_id")
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
		return
	}

	minCapacity := 1
	if mc := c.Query("min_capacity"); mc != "" {
		fmt.Sscanf(mc, "%d", &minCapacity)
	}

	tables, err := h.tableRepo.GetAvailableTables(c.Request.Context(), restaurantID, minCapacity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, tables)
}

func (h *TableHandler) UpdateTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid table id"})
		return
	}

	table, err := h.tableRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "table not found"})
		return
	}

	var req UpdateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if req.IsActive != nil {
		table.IsActive = *req.IsActive
	}
	if req.LocationType != nil {
		table.LocationType = *req.LocationType
	}
	if req.XPosition != nil {
		table.XPosition = req.XPosition
	}
	if req.YPosition != nil {
		table.YPosition = req.YPosition
	}

	if err := h.tableRepo.Update(c.Request.Context(), table); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, table)
}

func (h *TableHandler) DeleteTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid table id"})
		return
	}

	if err := h.tableRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "table not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

type CreateTableRequest struct {
	RestaurantID uuid.UUID           `json:"restaurant_id" binding:"required"`
	TableNumber  string              `json:"table_number" binding:"required"`
	MinCapacity  int                 `json:"min_capacity" binding:"required,min=1"`
	MaxCapacity  int                 `json:"max_capacity" binding:"required,min=1"`
	LocationType domain.LocationType `json:"location_type" binding:"required"`
	XPosition    *int                `json:"x_position"`
	YPosition    *int                `json:"y_position"`
}

type UpdateTableRequest struct {
	IsActive     *bool                `json:"is_active"`
	LocationType *domain.LocationType `json:"location_type"`
	XPosition    *int                 `json:"x_position"`
	YPosition    *int                 `json:"y_position"`
}
