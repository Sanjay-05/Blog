package service

import (
	"blog/config"
	"blog/helper"
	"blog/models"
	"blog/repository"
	"blog/request"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	OtpStore = make(map[string]helper.OTPData)
	OtpMutex = &sync.Mutex{}
)

// UserService defines user business logic
type UserService interface {
	RegisterUser(req request.RegisterUserRequest) (*models.User, error)
	LoginUser(req request.LoginRequest, cfg *config.Configuration) (string, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetAllUsers(limit, offset int) ([]models.User, int64, error)
	UpdateUser(userID uuid.UUID, req request.UpdateUserRequest) error
	DeleteUser(userID uuid.UUID) error
	RequestResetPassword(userName string) error
	ResetPassword(req *request.ResetPasswordRequest) error
	SwitchUserRole(userID uuid.UUID, role string) error
}

type userService struct {
	repo   repository.UserRepository
	logger *logrus.Logger
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository, logger *logrus.Logger) UserService {
	return &userService{repo: repo, logger: logger}
}

// RegisterUser registers a new user
func (s *userService) RegisterUser(req request.RegisterUserRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.repo.GetUserByName(req.UserName)
	if err == nil && existingUser != nil {
		s.logger.Warn("User registration failed - username already exists", "username", req.UserName)
		return nil, ErrUserExists
	}

	existingEmail, err := s.repo.GetUserByEmail(req.Email)
	if err == nil && existingEmail != nil {
		s.logger.Warn("User registration failed - email already exists", "email", req.Email)
		return nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", err)
		return nil, err
	}

	user := &models.User{
		UserName: req.UserName,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
		Status:   "active",
	}
	if err := s.repo.CreateUser(user); err != nil {
		s.logger.Error("Failed to create user", err)
		return nil, err
	}
	s.logger.Info("User registered successfully", "user_id", user.ID)
	return user, nil
}

// LoginUser authenticates user
func (s *userService) LoginUser(req request.LoginRequest, cfg *config.Configuration) (string, error) {
	user, err := s.repo.GetUserByName(req.UserName)
	if err != nil {
		s.logger.Warn("Login failed - user not found", "username", req.UserName)
		return "", errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("Login failed - invalid password", "username", req.UserName)
		return "", errors.New("invalid username or password")
	}

	token, err := helper.GenerateToken(cfg, user.ID, user.Email, user.Role)

	if err != nil {
		s.logger.Error("Failed to generate token", err)
		return "", err
	}
	s.logger.Info("User login successful", "user_id", user.ID)
	return token, nil
}

// GetUserByID gets user by ID
func (s *userService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

// GetAllUsers gets all users
func (s *userService) GetAllUsers(limit, offset int) ([]models.User, int64, error) {
	return s.repo.GetAllUsers(limit, offset)
}

func (s *userService) UpdateUser(userID uuid.UUID, req request.UpdateUserRequest) error {

	updates := map[string]interface{}{}
	if req.Email != "" {
		existing, err := s.repo.GetUserByEmail(req.Email)
		if err == nil && existing.ID != userID {
			return errors.New("email already in use")
		}
		updates["email"] = req.Email
	}

	if req.UserName != "" {
		existing, err := s.repo.GetUserByName(req.UserName)
		if err == nil && existing.ID != userID {
			return errors.New("username already in use")
		}
		updates["user_name"] = req.UserName
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	return s.repo.UpdateUser(userID, updates)
}

func (s *userService) DeleteUser(userID uuid.UUID) error {
	if err := s.repo.DeleteUser(userID); err != nil {
		s.logger.Error("Failed to delete user", err)
		return err
	}

	s.logger.Info("User deleted successfully", "user_id", userID)
	return nil
}

func (s *userService) RequestResetPassword(userName string) error {
	user, err := s.repo.GetUserByName(userName)
	if err != nil {
		return errors.New("user not found")
	}

	otp, err := helper.GenerateOTP()
	if err != nil {
		return err
	}

	hashedOTP, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	OtpMutex.Lock()
	OtpStore[user.UserName] = helper.OTPData{
		OTP:     string(hashedOTP),
		Expires: time.Now().Add(5 * time.Minute),
	}
	OtpMutex.Unlock()
	s.logger.Infof("OTP for %s: %s (expires in 5 mins)", user.Email, otp)
	return nil
}

func (s *userService) ResetPassword(req *request.ResetPasswordRequest) error {
	if req.NewPassword != req.ConfirmNewPassword {
		return errors.New("passwords do not match")
	}

	user, err := s.repo.GetUserByName(req.UserName)
	if err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.NewPassword)); err == nil {
		return errors.New("new password cannot be same as old password")
	}

	OtpMutex.Lock()
	data, ok := OtpStore[user.UserName]
	OtpMutex.Unlock()
	if !ok || time.Now().After(data.Expires) {
		return errors.New("OTP expired or invalid")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(data.OTP), []byte(req.OTP)); err != nil {
		return errors.New("invalid OTP")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateUser(user.ID, map[string]interface{}{"password": string(hashedPassword)}); err != nil {
		return err
	}
	helper.StartOTPCleanup(OtpMutex, OtpStore)
	return nil
}

func (s *userService) SwitchUserRole(userID uuid.UUID, role string) error {
	if role != "admin" && role != "user" {
		return errors.New("invalid role")
	}

	err := s.repo.UpdateUser(userID, map[string]interface{}{
		"role": role,
	})
	if err != nil {
		s.logger.Error("Failed to update role", err)
		return err
	}

	s.logger.Info("User role updated", "user_id", userID, "role", role)
	return nil
}
