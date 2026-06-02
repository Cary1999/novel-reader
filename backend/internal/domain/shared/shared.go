package shared

import (
	"errors"
	"fmt"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

var ErrNotFound = errors.New("not found")

type Actor struct {
	UserID   int64
	Username string
	Role     Role
}

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}
