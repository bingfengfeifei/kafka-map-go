package controller

import (
	"net/http"
	"time"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/service"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
	"github.com/gin-gonic/gin"
)

type AccountController struct {
	userService *service.UserService
	tokenCache  *util.TokenCache
}

func NewAccountController(userService *service.UserService, tokenCache *util.TokenCache) *AccountController {
	return &AccountController{
		userService: userService,
		tokenCache:  tokenCache,
	}
}

// Login handles user login
func (c *AccountController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Authenticate user
	user, err := c.userService.Authenticate(req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}

	// Generate token
	token, err := c.tokenCache.GenerateToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate token",
		})
		return
	}

	// Store token
	c.tokenCache.Set(token, user.Username)

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Login successful",
		Data: dto.LoginResponse{
			Token:    token,
			Username: user.Username,
		},
	})
}

// Logout handles user logout
func (c *AccountController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("X-Auth-Token")
	if token != "" {
		c.tokenCache.Delete(token)
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Logout successful",
	})
}

// GetInfo retrieves current user information
func (c *AccountController) GetInfo(ctx *gin.Context) {
	username, exists := ctx.Get("username")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	user, err := c.userService.GetUserByUsername(username.(string))
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.Response{
			Code:    http.StatusNotFound,
			Message: "User not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data: map[string]interface{}{
			"username":  user.Username,
			"createdAt": user.CreatedAt.Format(time.RFC3339),
		},
	})
}

// ChangePassword handles password change
func (c *AccountController) ChangePassword(ctx *gin.Context) {
	username, exists := ctx.Get("username")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	err := c.userService.ChangePassword(username.(string), req.OldPassword, req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Password changed successfully",
	})
}
