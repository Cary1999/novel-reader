package query

import (
	"context"

	categoryentity "novel-reader/backend/internal/domain/category/entity"
	categoryrepository "novel-reader/backend/internal/domain/category/repository"
)

type ListCategoriesHandler struct {
	repo categoryrepository.CategoryRepository
}

func NewListCategoriesHandler(repo categoryrepository.CategoryRepository) *ListCategoriesHandler {
	return &ListCategoriesHandler{repo: repo}
}

func (h *ListCategoriesHandler) Handle(ctx context.Context) ([]categoryentity.Category, error) {
	return h.repo.ListCategories(ctx)
}
