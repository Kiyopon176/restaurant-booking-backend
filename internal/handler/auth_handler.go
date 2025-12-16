package handler

import (
	"errors"
	"log"
	"net/http"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService service.AuthService
	userService service.UserService
}

func NewAuthHandler(authService service.AuthService, userService service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

type RegisterRequest struct {
	Email    string          `json:"email" binding:"required"`
	Password string          `json:"password" binding:"required"`
	Name     string          `json:"name" binding:"required"`
	Phone    *string         `json:"phone"`
	Role     domain.UserRole `json:"role" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
}

type UserResponse struct {
	ID        uuid.UUID       `json:"id"`
	Email     string          `json:"email"`
	Name      string          `json:"name"`
	Phone     *string         `json:"phone,omitempty"`
	Role      domain.UserRole `json:"role"`
	CreatedAt string          `json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, accessToken, refreshToken, err := h.authService.Register(
		req.Email,
		req.Password,
		req.Name,
		req.Phone,
		req.Role,
	)
	if err != nil {
		log.Printf("Register error: %v", err)
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid email format"})
		case errors.Is(err, service.ErrInvalidPassword):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Password must be at least 8 characters"})
		case errors.Is(err, service.ErrEmailExists):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Email already exists"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	accessToken, refreshToken, user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		log.Printf("Login error: %v", err)
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		log.Printf("Refresh token error: %v", err)
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
		case errors.Is(err, service.ErrExpiredRefreshToken):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Refresh token has expired"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		log.Printf("Logout error: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Token not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Logged out successfully"})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		log.Printf("GetMe error: %v", err)
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": toUserResponse(user),
	})
}

func toUserResponse(user *domain.User) *UserResponse {
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	phonePtr := &user.Phone
	if user.Phone == "" {
		phonePtr = nil
	}
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      name,
		Phone:     phonePtr,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func extractToken(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}
	return ""
}
