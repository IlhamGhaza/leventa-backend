package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a domain error that carries an HTTP status code.
type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	// Cause is the underlying error of an internal failure. It is logged but
	// never sent to the client.
	Cause error `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Unwrap exposes the underlying cause to errors.Is/As.
func (e *AppError) Unwrap() error { return e.Cause }

// Internal creates a 500 error that keeps cause for the server log while the
// client only sees "Internal server error".
func Internal(cause error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: "Internal server error", Cause: cause}
}

// New creates a new AppError.
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Newf creates a new AppError with a formatted message.
func Newf(code int, format string, args ...interface{}) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Common domain errors.
var (
	ErrNotFound       = New(http.StatusNotFound, "Resource not found")
	ErrUnauthorized   = New(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden      = New(http.StatusForbidden, "Forbidden")
	ErrConflict       = New(http.StatusConflict, "Resource already exists")
	ErrValidation     = New(http.StatusUnprocessableEntity, "Validation failed")
	ErrBadRequest     = New(http.StatusBadRequest, "Bad request")
	ErrInternalServer = New(http.StatusInternalServerError, "Internal server error")
)

// NotFound creates a not-found error with a custom message.
func NotFound(resource string) *AppError {
	return Newf(http.StatusNotFound, "%s not found", resource)
}

// Conflict creates a conflict error with a custom message.
func Conflict(message string) *AppError {
	return New(http.StatusConflict, message)
}

// Forbidden creates a forbidden error with a custom message.
func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, message)
}

// Validation creates a validation error with a custom message.
func Validation(message string) *AppError {
	return New(http.StatusUnprocessableEntity, message)
}

// BadRequest creates a bad-request error with a custom message.
func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message)
}

// IsAppError checks if an error is an AppError and returns it.
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
