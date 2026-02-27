package errors

import "net/http"

// AppError is a structured application error carrying an HTTP status code.
type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

// Sentinel errors used across the application.
var (
	ErrNotFound      = &AppError{Code: http.StatusNotFound, Message: "not found"}
	ErrUnauthorized  = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrBadRequest    = &AppError{Code: http.StatusBadRequest, Message: "bad request"}
	ErrInvalidPIN    = &AppError{Code: http.StatusUnauthorized, Message: "invalid PIN"}
	ErrPINAlreadySet = &AppError{Code: http.StatusConflict, Message: "PIN already configured"}
)
