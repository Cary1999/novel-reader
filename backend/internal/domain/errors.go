package domain

import "fmt"

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
