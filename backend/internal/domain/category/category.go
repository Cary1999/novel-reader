package category

import (
	"context"
	"time"
)

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type Repository interface {
	ListCategories(ctx context.Context) ([]Category, error)
	FindCategoryByID(ctx context.Context, id int64) (Category, error)
	CreateCategory(ctx context.Context, name string) (Category, error)
	UpdateCategory(ctx context.Context, id int64, name string) (Category, error)
	DeleteCategory(ctx context.Context, id int64) error
}
