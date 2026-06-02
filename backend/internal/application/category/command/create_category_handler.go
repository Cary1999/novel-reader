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

type CreateCategory struct {
	Name string `json:"name"`
}

type CreateCategoryHandler struct {
	repo    categoryrepository.CategoryRepository
	service *categoryservice.CategoryService
}

func NewCreateCategoryHandler(repo categoryrepository.CategoryRepository, service *categoryservice.CategoryService) *CreateCategoryHandler {
	return &CreateCategoryHandler{repo: repo, service: service}
}

func (h *CreateCategoryHandler) Handle(ctx context.Context, input CreateCategory) (categoryentity.Category, error) {
	name, err := h.service.NormalizeName(input.Name)
	if err != nil {
		return categoryentity.Category{}, err
	}
	item, err := h.repo.CreateCategory(ctx, h.service.NewCategory(name).Name)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return categoryentity.Category{}, shared.NewError(http.StatusConflict, "CONFLICT", "category already exists")
		}
		return categoryentity.Category{}, err
	}
	return item, nil
}
