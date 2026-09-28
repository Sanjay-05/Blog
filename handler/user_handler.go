package handler

import (
	"blog/config"
	"blog/helper"
	"blog/request"
	"blog/response"
	"blog/service"
	"blog/utility"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// UserHandler handles user-related requests
type UserHandler interface {
	Register(c fiber.Ctx) error
	Login(c fiber.Ctx) error
	GetProfile(c fiber.Ctx) error
	GetAllUsers(c fiber.Ctx) error
	UpdateUser(c fiber.Ctx) error
	SwitchUserRole(c fiber.Ctx) error 
	DeleteUser(c fiber.Ctx) error
	ForgetPasswordHandler(c fiber.Ctx) error
	ResetPassword(c fiber.Ctx) error
}

type userHandler struct {
	userService service.UserService
	cfg         *config.Configuration
	logger      *logrus.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(us service.UserService, cfg *config.Configuration, logger *logrus.Logger) UserHandler {
	return &userHandler{userService: us, cfg: cfg, logger: logger}
}

// Register handles user registration
func (h *userHandler) Register(c fiber.Ctx) error {
	var req request.RegisterUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("failed to bind registration request", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	if err := helper.ValidateJson(req); err != nil {
		h.logger.Warn("validation failed for registration", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	user, err := h.userService.RegisterUser(req)
	if err != nil {
		if err == service.ErrEmailExists || err == service.ErrUserExists {
			return utility.SendErrorResponse(c, http.StatusConflict, "registration failed", err)
		}
		h.logger.Error("Registration failed", err)
		return utility.SendErrorResponse(c, http.StatusInternalServerError, "registration failed", err)
	}

	h.logger.Info("user registered successfully", "user_id", user.ID)
	return utility.SendSuccessResponse(c, http.StatusCreated, "user registered successfully", nil)
}

// GetProfile handles getting user profile
func (h *userHandler) Login(c fiber.Ctx) error {
	var req request.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("failed to bind login request", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	if err := helper.ValidateJson(req); err != nil {
		h.logger.Error("validation failed for registration", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	token, err := h.userService.LoginUser(req, h.cfg)
	if err != nil {
		h.logger.Error("Login attempt failed", err)
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "login failed", err)
	}

	h.logger.Info("user login successfull", "token", token)
	return c.JSON(fiber.Map{"message": "login successfull", "token": token})
}

// GetProfile handles getting user profile
func (h *userHandler) GetProfile(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("unauthorized access to GetProfile")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.logger.Error("failed to get user profile", err)
		return utility.SendErrorResponse(c, http.StatusNotFound, "user not found", err)
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "user profile retrieved successfully", response.UserResponse{
		UserName: user.UserName,
		ID:       user.ID,
		Email:    user.Email,
		Role:     user.Role,
	})
}

// GetAllUsers handles getting all users (admin only)
func (h *userHandler) GetAllUsers(c fiber.Ctx) error {
	limit, offset := utility.GetPagination(c)

	users, total, err := h.userService.GetAllUsers(limit, offset)
	if err != nil {
		h.logger.Error("Failed to get users", err)
		statusCode := utility.GetStatusCodeFromError(err)
		return utility.SendErrorResponse(c, statusCode, "failed to get users", err)
	}

	if len(users) == 0 {
		return utility.SendSuccessResponse(c, http.StatusOK, "no users found", nil)
	}
	var userResponses []response.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, response.UserResponse{
			ID:       user.ID,
			UserName: user.UserName,
			Email:    user.Email,
			Role:     user.Role,
		})
	}
	return utility.SendSuccessResponse(c, http.StatusOK, "user fetched successfully", map[string]any{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"users":  userResponses,
	})
}

func (h *userHandler) UpdateUser(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	var req request.UpdateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request", err)
	}

	if err := helper.ValidateJson(req); err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	err := h.userService.UpdateUser(userID, req)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		return utility.SendErrorResponse(c, statusCode, "failed to update user", err)
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "user updated successfully", nil)
}

func (h *userHandler) DeleteUser(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	role, _ := c.Locals("role").(string)

	var targetUserID uuid.UUID

	if role == "admin" {
		type DeleteUserRequest struct {
			UserID uuid.UUID `json:"user_id"`
		}
		var req DeleteUserRequest
		if err := c.Bind().JSON(&req); err != nil {
			return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
		}
		if req.UserID == uuid.Nil {
			return utility.SendErrorResponse(c, http.StatusBadRequest, "user_id is required for admin", errors.New("user_id required"))
		}
		targetUserID = req.UserID
	} else {
		targetUserID = userID
	}

	if role != "admin" && targetUserID != userID {
		return utility.SendErrorResponse(c, http.StatusForbidden, "unauthorized", errors.New("insufficient permissions"))
	}
	if err := h.userService.DeleteUser(targetUserID); err != nil {
		h.logger.Error("Failed to delete user", err)
		return utility.SendErrorResponse(c, http.StatusInternalServerError, "failed to delete user", err)
	}

	h.logger.Info("User deleted successfully", "user_id", targetUserID)
	return utility.SendSuccessResponse(c, http.StatusOK, "user deleted successfully", nil)
}

// RequestPasswordReset handles sending OTP to user email
func (h *userHandler) ForgetPasswordHandler(c fiber.Ctx) error {
	var req request.ForgetPasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("Failed to parse forget password request", err)
		return utility.SendErrorResponse(c, fiber.StatusBadRequest, "invalid request payload", err)
	}

	if err := h.userService.RequestResetPassword(req.UserName); err != nil {
		h.logger.Error("Failed to send OTP", err)
		return utility.SendErrorResponse(c, fiber.StatusUnauthorized, "failed to send OTP", err)
	}

	return utility.SendSuccessResponse(c, fiber.StatusOK, "OTP sent to registered email", nil)
}

func (h *userHandler) ResetPassword(c fiber.Ctx) error {
	var req request.ResetPasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("Failed to parse reset password request", err)
		return utility.SendErrorResponse(c, fiber.StatusBadRequest, "invalid request payload", err)
	}

	if err := h.userService.ResetPassword(&req); err != nil {
		h.logger.Error("Failed to reset password", err)
		return utility.SendErrorResponse(c, fiber.StatusUnauthorized, "failed to reset password", err)
	}

	return utility.SendSuccessResponse(c, fiber.StatusOK, "Password reset successfully", nil)
}

// SwitchUserRole handles role update (admin only, cannot update self)
func (h *userHandler) SwitchUserRole(c fiber.Ctx) error {
	currentUserID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	role, ok := c.Locals("role").(string)
	if !ok || role != "admin" {
		return utility.SendErrorResponse(c, http.StatusForbidden, "admin access required", errors.New("insufficient permissions"))
	}

	var req request.SwitchRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("Failed to parse role switch request", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	if err := helper.ValidateJson(req); err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	if req.UserID == currentUserID {
		return utility.SendErrorResponse(c, http.StatusForbidden, "cannot change your own role", errors.New("self role change not allowed"))
	}

	if err := h.userService.SwitchUserRole(req.UserID, req.Role); err != nil {
		h.logger.Error("Failed to switch user role", err)
		return utility.SendErrorResponse(c, http.StatusInternalServerError, "failed to switch role", err)
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "user role updated successfully", nil)
}