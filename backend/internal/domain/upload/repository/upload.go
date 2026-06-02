package repository

import (
	"context"

	uploadentity "novel-reader/backend/internal/domain/upload/entity"
)

type UploadRepository interface {
	CreateUpload(ctx context.Context, item uploadentity.Upload) (int64, error)
	MarkUpload(ctx context.Context, id int64, status, message string) error
}
