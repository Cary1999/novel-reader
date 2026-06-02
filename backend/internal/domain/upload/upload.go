package upload

import (
	"context"
	"time"
)

type Upload struct {
	ID               int64
	ActorUserID      int64
	OriginalFilename string
	StoredPath       string
	FileSize         int64
	Status           string
	ErrorMessage     string
	CreatedAt        time.Time
}

type ImportResult struct {
	BookID            int64  `json:"bookId"`
	ChapterCount      int    `json:"chapterCount"`
	UploadID          int64  `json:"uploadId"`
	FirstChapterTitle string `json:"firstChapterTitle"`
	LastChapterTitle  string `json:"lastChapterTitle"`
}

type Repository interface {
	CreateUpload(ctx context.Context, item Upload) (int64, error)
	MarkUpload(ctx context.Context, id int64, status, message string) error
}
