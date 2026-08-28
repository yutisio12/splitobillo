package apperr

import "net/http"

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *AppError) Error() string {
	return e.Code + ": " + e.Message
}

func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

func BadRequest(message string) *AppError {
	return New("INVALID_REQUEST", message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return New("UNAUTHORIZED", message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New("FORBIDDEN", message, http.StatusForbidden)
}

func NotFound(message string) *AppError {
	return New("NOT_FOUND", message, http.StatusNotFound)
}

func Conflict(message string) *AppError {
	return New("CONFLICT", message, http.StatusConflict)
}

func TooManyRequests(message string) *AppError {
	return New("RATE_LIMITED", message, http.StatusTooManyRequests)
}

func Unprocessable(message string) *AppError {
	return New("INVALID_RECEIPT", message, http.StatusUnprocessableEntity)
}

func ServiceUnavailable(message string) *AppError {
	return New("SERVICE_UNAVAILABLE", message, http.StatusServiceUnavailable)
}

func Internal(message string) *AppError {
	return New("INTERNAL_SERVER_ERROR", message, http.StatusInternalServerError)
}
