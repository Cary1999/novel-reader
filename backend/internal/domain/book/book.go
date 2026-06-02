package book

import (
	"context"
	"time"
)

type Book struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	Author             string    `json:"author"`
	OwnerUserID        *int64    `json:"-"`
	CategoryID         *int64    `json:"categoryId,omitempty"`
	Category           string    `json:"category"`
	Description        string    `json:"description"`
	ChapterCount       int       `json:"chapterCount"`
	LatestChapterTitle string    `json:"latestChapterTitle"`
	RecommendScore     int       `json:"recommendScore"`
	CoverURL           string    `json:"coverUrl"`
	CoverPath          *string   `json:"-"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"-"`
}

type Chapter struct {
	ID      int64  `json:"id"`
	BookID  int64  `json:"bookId"`
	Index   int    `json:"index"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
}

type ChapterDraft struct {
	Index   int
	Title   string
	Content string
}

type CreateInput struct {
	Title       string
	CategoryID  int64
	Description string
	OwnerUserID int64
}

type MetadataInput struct {
	Title          string `json:"title"`
	CategoryID     int64  `json:"categoryId"`
	Description    string `json:"description"`
	RecommendScore *int   `json:"recommendScore,omitempty"`
}

type ChapterInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Repository interface {
	SearchBooks(ctx context.Context, q, category string, page, pageSize int) ([]Book, int, error)
	ListRecommendedBooks(ctx context.Context, page, pageSize int) ([]Book, int, error)
	ListBooksByOwner(ctx context.Context, ownerID int64, q string, categoryID int64, page, pageSize int) ([]Book, int, error)
	FindBook(ctx context.Context, id int64) (Book, error)
	UpdateBookCoverPath(ctx context.Context, bookID int64, coverPath *string) error
	ListChapters(ctx context.Context, bookID int64) ([]Chapter, error)
	FindChapter(ctx context.Context, bookID, chapterID int64) (Chapter, error)
	CreateBookWithChapters(ctx context.Context, input CreateInput, categoryID *int64, uploadID int64, chapters []ChapterDraft) (int64, error)
	CreateBook(ctx context.Context, input CreateInput, categoryID *int64) (Book, error)
	UpdateBookMetadata(ctx context.Context, bookID int64, input MetadataInput, categoryID *int64) (Book, error)
	DeleteBook(ctx context.Context, bookID int64) error
	AddChapter(ctx context.Context, bookID int64, input ChapterInput) (Chapter, error)
	UpdateChapter(ctx context.Context, bookID, chapterID int64, input ChapterInput) (Chapter, error)
	DeleteChapter(ctx context.Context, bookID, chapterID int64) error
}
