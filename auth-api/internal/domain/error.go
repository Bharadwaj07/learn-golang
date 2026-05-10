package domain

import "errors"

// sentinel errors — used with errors.Is
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("already exists")
	ErrUnauthorized = errors.New("unauthorized")
)

// AppError carries an HTTP code + safe client message + internal detail
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// constructors
func NewNotFound(msg string, err error) *AppError {
	return &AppError{Code: 404, Message: msg, Err: err}
}

func NewUnauthorized(msg string) *AppError {
	return &AppError{Code: 401, Message: msg}
}

func NewConflict(msg string) *AppError {
	return &AppError{Code: 409, Message: msg}
}

func NewBadRequest(msg string) *AppError {
	return &AppError{Code: 400, Message: msg}
}

func NewInternal(err error) *AppError {
	return &AppError{Code: 500, Message: "internal server error", Err: err}
}
