package service

import (
	"net/http"
	"strings"

	"novel-reader/backend/internal/domain/shared"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) NormalizeGroupName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "group name is required")
	}
	if len([]rune(value)) > 64 {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "group name is too long")
	}
	return value, nil
}

func (s *Service) NormalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

func (s *Service) NormalizeAction(action string) (string, error) {
	action = strings.TrimSpace(strings.ToLower(action))
	switch action {
	case "move", "pin", "unpin", "remove":
		return action, nil
	default:
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid bookshelf action")
	}
}
