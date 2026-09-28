package service

import "errors"

var (
	ErrCategoryExists = errors.New("category with this name already exists")
	ErrUserExists     = errors.New("user name already exists")
	ErrEmailExists    = errors.New("email already exists")
)

// ServiceError defines a custom error with HTTP status code
type ServiceError struct {
	Message    string
	StatusCode int
	Err        error
}

func (e *ServiceError) Error() string {
	return e.Message
}

// NewServiceError creates a new service error
func NewServiceError(message string, statusCode int, err error) *ServiceError {
	return &ServiceError{
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common service errors
var (
	ErrNotFound            = NewServiceError("resource not found", 404, errors.New("not found"))
	ErrBadRequest          = NewServiceError("invalid request", 400, errors.New("bad request"))
	ErrDuplicate           = NewServiceError("resource already exists", 409, errors.New("conflict"))
	ErrUnauthorized        = NewServiceError("unauthorized", 401, errors.New("unauthorized"))
	ErrForbidden           = NewServiceError("forbidden", 403, errors.New("forbidden"))
	ErrInternalServer      = NewServiceError("internal server error", 500, errors.New("internal server error"))
	ErrValidation          = NewServiceError("validation failed", 400, errors.New("validation failed"))
	ErrForeignKeyViolation = NewServiceError("referenced resource not found", 400, errors.New("foreign key violation"))
	ErrUniqueViolation     = NewServiceError("resource already exists", 409, errors.New("unique violation"))
	ErrOwnershipViolation  = NewServiceError("you don't have permission to access this resource", 403, errors.New("insufficient permissions"))
)
