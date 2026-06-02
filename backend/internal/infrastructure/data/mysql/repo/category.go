package repo

import (
	"context"
	"database/sql"

	categoryentity "novel-reader/backend/internal/domain/category/entity"
	"novel-reader/backend/internal/domain/shared"
	"novel-reader/backend/internal/infrastructure/data/mysql/model"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return CategoryRepository{db: db}
}

func (r CategoryRepository) ListCategories(ctx context.Context) ([]categoryentity.Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, created_at FROM categories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []categoryentity.Category
	for rows.Next() {
		var record model.Category
		if err := rows.Scan(&record.ID, &record.Name, &record.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, record.ToEntity())
	}
	return items, rows.Err()
}

func (r CategoryRepository) FindCategoryByID(ctx context.Context, id int64) (categoryentity.Category, error) {
	var record model.Category
	err := r.db.QueryRowContext(ctx, `SELECT id, name, created_at FROM categories WHERE id = ?`, id).Scan(&record.ID, &record.Name, &record.CreatedAt)
	if err != nil {
		return categoryentity.Category{}, mapNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r CategoryRepository) CreateCategory(ctx context.Context, name string) (categoryentity.Category, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO categories (name) VALUES (?)`, name)
	if err != nil {
		return categoryentity.Category{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return categoryentity.Category{}, err
	}
	return r.FindCategoryByID(ctx, id)
}

func (r CategoryRepository) UpdateCategory(ctx context.Context, id int64, name string) (categoryentity.Category, error) {
	if _, err := r.FindCategoryByID(ctx, id); err != nil {
		return categoryentity.Category{}, err
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE categories SET name = ? WHERE id = ?`, name, id); err != nil {
		return categoryentity.Category{}, err
	}
	return r.FindCategoryByID(ctx, id)
}

func (r CategoryRepository) DeleteCategory(ctx context.Context, id int64) error {
	var used int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books WHERE category_id = ?`, id).Scan(&used); err != nil {
		return err
	}
	if used > 0 {
		return shared.NewError(409, "CONFLICT", "category is used by books")
	}
	result, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return shared.ErrNotFound
	}
	return nil
}
