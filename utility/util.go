package utility

import (
	"blog/response"
	"blog/service"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// SendSuccessResponse sends a successful response
func SendSuccessResponse(c fiber.Ctx, statusCode int, message string, data interface{}) error {
	resp := response.APIResponse{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
	return c.Status(statusCode).JSON(resp)
}

// SendErrorResponse sends an error response
func SendErrorResponse(c fiber.Ctx, statusCode int, message string, err error) error {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	resp := response.APIResponse{
		StatusCode: statusCode,
		Message:    message,
		Error:      errMsg,
	}
	return c.Status(statusCode).JSON(resp)
}

// GetStatusCodeFromError extracts status code from ServiceError
func GetStatusCodeFromError(err error) int {
	if serviceErr, ok := err.(*service.ServiceError); ok {
		return serviceErr.StatusCode
	}
	return 500
}

// GetPagination extracts pagination parameters from query
func GetPagination(c fiber.Ctx) (limit, offset int) {
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")

	limit, _ = strconv.Atoi(limitStr)
	offset, _ = strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return
}
