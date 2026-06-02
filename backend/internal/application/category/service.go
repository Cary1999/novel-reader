package categoryapp

import (
	"context"
	"net/http"
	"strings"

	"novel-reader/backend/internal/domain/category"
	"novel-reader/backend/internal/domain/shared"
)

type Service struct {
	repo category.Repository
}

func NewService(repo category.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListCategories(ctx context.Context) ([]category.Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, name string) (category.Category, error) {
	name, err := normalizeCategoryName(name)
	if err != nil {
		return category.Category{}, err
	}
	item, err := s.repo.CreateCategory(ctx, name)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return category.Category{}, shared.NewError(http.StatusConflict, "CONFLICT", "category already exists")
		}
		return category.Category{}, err
	}
	return item, nil
}

func (s *Service) UpdateCategory(ctx context.Context, id int64, name string) (category.Category, error) {
	name, err := normalizeCategoryName(name)
	if err != nil {
		return category.Category{}, err
	}
	item, err := s.repo.UpdateCategory(ctx, id, name)
	if err != nil {
		if err == shared.ErrNotFound {
			return category.Category{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return category.Category{}, shared.NewError(http.StatusConflict, "CONFLICT", "category already exists")
		}
		return category.Category{}, err
	}
	return item, nil
}

func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		return err
	}
	return nil
}

func normalizeCategoryName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "category name is required")
	}
	if len([]rune(name)) > 64 {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "category name is too long")
	}
	return name, nil
}
