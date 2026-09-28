package middleware

import (
	"blog/config"
	"blog/models"
	"blog/utility"
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type Config struct {
	*config.Configuration
}

// JWTMiddleware validates JWT token
func JWTMiddleware(cfg *config.Configuration, logger *logrus.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Authorization header missing")
			return utility.SendErrorResponse(c, http.StatusUnauthorized, "missing authorization header", errors.New("authorization header required"))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format")
			return utility.SendErrorResponse(c, http.StatusUnauthorized, "invalid authorization format", errors.New("invalid authorization format"))
		}

		tokenString := parts[1]

		claims := &models.JWTCustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.ApiSecret), nil
		})

		if err != nil {
			logger.Warn("JWT parsing failed", err)
			return utility.SendErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
		}

		if !token.Valid {
			logger.Warn("Invalid JWT token")
			return utility.SendErrorResponse(c, http.StatusUnauthorized, "invalid token", errors.New("token is invalid"))
		}

		// Set user information in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// AdminMiddleware checks if user is admin
func AdminMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != "admin" {
			return utility.SendErrorResponse(c, http.StatusForbidden, "admin access required", errors.New("you don't have permission to access this resource"))
		}
		return c.Next()
	}
}
