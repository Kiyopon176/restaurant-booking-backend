package handler

import (
	"net/http"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// CreateUser godoc
// @Summary      Создать пользователя
// @Description  Регистрация нового пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user body CreateUserRequest true "Данные пользователя"
// @Success      201 {object} domain.User
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user := &domain.User{
		Email:     req.Email,
		Password:  req.Password, // TODO: Хешировать пароль!
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	user.Password = ""

	c.JSON(http.StatusCreated, user)
}

// GetUser godoc
// @Summary      Получить пользователя
// @Description  Получить информацию о пользователе по ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} domain.User
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}

	user, err := h.userRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		return
	}

	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// ListUsers godoc
// @Summary      Список пользователей
// @Description  Получить список всех пользователей
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        limit query int false "Лимит" default(10)
// @Param        offset query int false "Оффсет" default(0)
// @Success      200 {array} domain.User
// @Failure      500 {object} ErrorResponse
// @Router       /api/users [get]
/*
func (h *UserHandler) ListUsers(c *gin.Context) {
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	users, err := h.userRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	for _, user := range users {
		user.Password = ""
	}

	c.JSON(http.StatusOK, users)
}
*/

type CreateUserRequest struct {
	Email     string          `json:"email" binding:"required,email"`
	Password  string          `json:"password" binding:"required,min=6"`
	FirstName string          `json:"first_name" binding:"required"`
	LastName  string          `json:"last_name" binding:"required"`
	Phone     string          `json:"phone" binding:"required"`
	Role      domain.UserRole `json:"role" binding:"required"`
}
