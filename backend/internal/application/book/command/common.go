package command

import (
	"context"
	"net/http"

	categoryrepository "novel-reader/backend/internal/domain/category/repository"
	"novel-reader/backend/internal/domain/shared"
)

func requireCategory(ctx context.Context, categories categoryrepository.CategoryRepository, categoryID int64) (*int64, error) {
	if categoryID <= 0 {
		return nil, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "category is required")
	}
	if _, err := categories.FindCategoryByID(ctx, categoryID); err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		return nil, err
	}
	return &categoryID, nil
}
