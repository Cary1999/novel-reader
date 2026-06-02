package service

import (
	"net/http"
	"strings"

	categoryentity "novel-reader/backend/internal/domain/category/entity"
	"novel-reader/backend/internal/domain/shared"
)

type CategoryService struct{}

func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

func (s *CategoryService) NewCategory(name string) categoryentity.Category {
	return categoryentity.Category{Name: name}
}

func (s *CategoryService) NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "category name is required")
	}
	if len([]rune(name)) > 64 {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "category name is too long")
	}
	return name, nil
}
