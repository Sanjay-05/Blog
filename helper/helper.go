package helper

import (
	"blog/config"
	"blog/models"
	"blog/response"
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type OTPData struct {
	OTP     string
	Expires time.Time
}

func ValidateJson(data interface{}) error {
	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return NewCustomError(400, "nil request body")
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return NewCustomError(400, "expected struct")
	}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		validation := field.Tag.Get("validate")
		if validation == "" {
			continue
		}

		rules := strings.Split(validation, ",")

		for _, rule := range rules {
			switch {
			case rule == "required":
				if value.IsZero() {
					return NewCustomError(400, field.Name+" is required")
				}

			case strings.HasPrefix(rule, "min="):
				minStr := strings.TrimPrefix(rule, "min=")
				minVal, err := strconv.Atoi(minStr)
				if err != nil {
					return NewCustomError(500, "invalid min tag")
				}

				if value.Kind() == reflect.String && len(value.String()) < minVal {
					return NewCustomError(400, field.Name+" must be at least "+minStr+" characters")
				}

			case strings.HasPrefix(rule, "max="):
				maxStr := strings.TrimPrefix(rule, "max=")
				maxVal, err := strconv.Atoi(maxStr)
				if err != nil {
					return NewCustomError(500, "invalid max tag")
				}

				if value.Kind() == reflect.String && len(value.String()) > maxVal {
					return NewCustomError(400, field.Name+" must be at most "+maxStr+" characters")
				}
			}
		}
	}
	return nil
}

type CustomError struct {
	Code    int
	Message string
}

func NewCustomError(code int, message string) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
	}
}

func (e *CustomError) Error() string {
	return e.Message
}

// GenerateToken generates JWT token
func GenerateToken(cfg *config.Configuration, userID uuid.UUID, email, role string) (string, error) {
	claims := &models.JWTCustomClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "blog-api",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.ApiSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func PostToResponse(post *models.Post) response.PostResponse {
	return response.PostResponse{
		ID:         post.ID,
		Title:      post.Title,
		Slug:       post.Slug,
		Content:    post.Content,
		Status:     post.Status,
		UserID:     &post.UserID,
		CategoryID: &post.CategoryID,
		CreatedAt:  post.CreatedAt,
	}
}

func PostDetailResponse(post *models.Post) response.PostResponse {
	return response.PostResponse{
		ID:        post.ID,
		Title:     post.Title,
		Slug:      post.Slug,
		Content:   post.Content,
		Status:    post.Status,
		CreatedAt: post.CreatedAt,
		User: &response.UserResponse{
			ID:       post.UserID,
			UserName: post.User.UserName,
		},
		Category: &response.CategoryResponse{
			ID:           post.CategoryID,
			CategoryName: post.Category.CategoryName,
			Description:  post.Category.Description,
		},
	}
}

func CommentToResponse(comment *models.Comment) response.CommentResponse {
	return response.CommentResponse{
		ID:      comment.ID,
		PostID:  comment.PostID,
		Comment: comment.Comment,
		UserID:  comment.UserID,
		User: response.UserResponse{
			ID:       comment.UserID,
			UserName: comment.User.UserName,
		},
		CreatedAt: comment.CreatedAt,
	}
}

func CalculateTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	years := int(diff.Hours() / (24 * 365))
	months := int(diff.Hours() / (24 * 30))
	days := int(diff.Hours() / 24)

	if years > 0 {
		return fmt.Sprintf("%d years - %d months ago", years, months%12)
	} else if months > 0 {
		return fmt.Sprintf("%d months - %d days ago", months, days%30)
	} else if days > 0 {
		return fmt.Sprintf("%d days ago", days)
	}
	return "just now"
}

// helper function to generate random number for otp to reset password
func GenerateOTP() (string, error) {
	max := big.NewInt(1000000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n), nil
}

// helper function to cleanup the otp
func StartOTPCleanup(otpMutex *sync.Mutex, otpStore map[string]OTPData) {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			otpMutex.Lock()
			for k, v := range otpStore {
				if time.Now().After(v.Expires) {
					delete(otpStore, k)
				}
			}
			otpMutex.Unlock()
		}
	}()
}
