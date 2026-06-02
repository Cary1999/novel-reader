package command

import (
	"context"
	"net/http"
	"strings"

	categoryentity "novel-reader/backend/internal/domain/category/entity"
	categoryrepository "novel-reader/backend/internal/domain/category/repository"
	categoryservice "novel-reader/backend/internal/domain/category/service"
	"novel-reader/backend/internal/domain/shared"
)

type UpdateCategory struct {
	Name string `json:"name"`
}

type UpdateCategoryHandler struct {
	repo    categoryrepository.CategoryRepository
	service *categoryservice.CategoryService
}

func NewUpdateCategoryHandler(repo categoryrepository.CategoryRepository, service *categoryservice.CategoryService) *UpdateCategoryHandler {
	return &UpdateCategoryHandler{repo: repo, service: service}
}

func (h *UpdateCategoryHandler) Handle(ctx context.Context, id int64, input UpdateCategory) (categoryentity.Category, error) {
	name, err := h.service.NormalizeName(input.Name)
	if err != nil {
		return categoryentity.Category{}, err
	}
	item, err := h.repo.UpdateCategory(ctx, id, name)
	if err != nil {
		if err == shared.ErrNotFound {
			return categoryentity.Category{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return categoryentity.Category{}, shared.NewError(http.StatusConflict, "CONFLICT", "category already exists")
		}
		return categoryentity.Category{}, err
	}
	return item, nil
}
