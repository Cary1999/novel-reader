package command

import (
	"context"
	"net/http"

	categoryrepository "novel-reader/backend/internal/domain/category/repository"
	"novel-reader/backend/internal/domain/shared"
)

type DeleteCategoryHandler struct {
	repo categoryrepository.CategoryRepository
}

func NewDeleteCategoryHandler(repo categoryrepository.CategoryRepository) *DeleteCategoryHandler {
	return &DeleteCategoryHandler{repo: repo}
}

func (h *DeleteCategoryHandler) Handle(ctx context.Context, id int64) error {
	if err := h.repo.DeleteCategory(ctx, id); err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		return err
	}
	return nil
}
