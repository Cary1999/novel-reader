package repo

import (
	"context"
	"database/sql"

	uploadentity "novel-reader/backend/internal/domain/upload/entity"
	"novel-reader/backend/internal/infrastructure/data/mysql/model"
)

type UploadRepository struct {
	db *sql.DB
}

func NewUploadRepository(db *sql.DB) UploadRepository {
	return UploadRepository{db: db}
}

func (r UploadRepository) CreateUpload(ctx context.Context, item uploadentity.Upload) (int64, error) {
	var record model.Upload
	record.FromEntity(&item)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO uploads (admin_user_id, original_filename, stored_path, file_size, status)
		VALUES (?, ?, ?, ?, ?)
	`, record.ActorUserID, record.OriginalFilename, record.StoredPath, record.FileSize, record.Status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r UploadRepository) MarkUpload(ctx context.Context, id int64, status, message string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE uploads SET status = ?, error_message = ? WHERE id = ?`, status, message, id)
	return err
}
