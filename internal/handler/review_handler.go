package handler

import (
	"fmt"
	"net/http"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReviewHandler struct {
	reviewRepo     repository.ReviewRepository
	restaurantRepo repository.RestaurantRepository
}

func NewReviewHandler(reviewRepo repository.ReviewRepository, restaurantRepo repository.RestaurantRepository) *ReviewHandler {
	return &ReviewHandler{
		reviewRepo:     reviewRepo,
		restaurantRepo: restaurantRepo,
	}
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	review := &domain.Review{
		RestaurantID: req.RestaurantID,
		UserID:       req.UserID,
		BookingID:    req.BookingID,
		Rating:       req.Rating,
		Comment:      req.Comment,
		IsVisible:    true,
	}

	if err := h.reviewRepo.Create(c.Request.Context(), review); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, review)
}

func (h *ReviewHandler) GetReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid review id"})
		return
	}

	review, err := h.reviewRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "review not found"})
		return
	}

	c.JSON(http.StatusOK, review)
}

func (h *ReviewHandler) GetRestaurantReviews(c *gin.Context) {
	idStr := c.Param("id")
	restaurantID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid restaurant id"})
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

	reviews, err := h.reviewRepo.GetByRestaurantID(c.Request.Context(), restaurantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func (h *ReviewHandler) GetUserReviews(c *gin.Context) {
	idStr := c.Param("user_id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	reviews, err := h.reviewRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid review id"})
		return
	}

	review, err := h.reviewRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "review not found"})
		return
	}

	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if req.Rating != nil {
		review.Rating = *req.Rating
	}
	if req.Comment != nil {
		review.Comment = *req.Comment
	}

	if err := h.reviewRepo.Update(c.Request.Context(), review); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, review)
}

func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid review id"})
		return
	}

	if err := h.reviewRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "review not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

type CreateReviewRequest struct {
	RestaurantID uuid.UUID  `json:"restaurant_id" binding:"required"`
	UserID       uuid.UUID  `json:"user_id" binding:"required"`
	BookingID    *uuid.UUID `json:"booking_id"`
	Rating       int        `json:"rating" binding:"required,min=1,max=5"`
	Comment      string     `json:"comment"`
}

type UpdateReviewRequest struct {
	Rating  *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	Comment *string `json:"comment"`
}
