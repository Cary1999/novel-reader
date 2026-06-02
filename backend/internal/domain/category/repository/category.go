package repository

import (
	"context"

	categoryentity "novel-reader/backend/internal/domain/category/entity"
)

type CategoryRepository interface {
	ListCategories(ctx context.Context) ([]categoryentity.Category, error)
	FindCategoryByID(ctx context.Context, id int64) (categoryentity.Category, error)
	CreateCategory(ctx context.Context, name string) (categoryentity.Category, error)
	UpdateCategory(ctx context.Context, id int64, name string) (categoryentity.Category, error)
	DeleteCategory(ctx context.Context, id int64) error
}
