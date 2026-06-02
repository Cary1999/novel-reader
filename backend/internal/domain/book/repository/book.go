package repository

import (
	"context"

	bookentity "novel-reader/backend/internal/domain/book/entity"
)

type BookRepository interface {
	SearchBooks(ctx context.Context, q, category string, page, pageSize int) ([]bookentity.Book, int, error)
	ListRecommendedBooks(ctx context.Context, page, pageSize int) ([]bookentity.Book, int, error)
	ListBooksByOwner(ctx context.Context, ownerID int64, q string, categoryID int64, page, pageSize int) ([]bookentity.Book, int, error)
	FindBook(ctx context.Context, id int64) (bookentity.Book, error)
	UpdateBookCoverPath(ctx context.Context, bookID int64, coverPath *string) error
	ListChapters(ctx context.Context, bookID int64) ([]bookentity.Chapter, error)
	FindChapter(ctx context.Context, bookID, chapterID int64) (bookentity.Chapter, error)
	CreateBookWithChapters(ctx context.Context, item bookentity.Book, categoryID *int64, uploadID int64, chapters []bookentity.ChapterDraft) (int64, error)
	CreateBook(ctx context.Context, item bookentity.Book, categoryID *int64) (bookentity.Book, error)
	UpdateBookMetadata(ctx context.Context, bookID int64, item bookentity.Book, categoryID *int64) (bookentity.Book, error)
	DeleteBook(ctx context.Context, bookID int64) error
	AddChapter(ctx context.Context, bookID int64, item bookentity.Chapter) (bookentity.Chapter, error)
	UpdateChapter(ctx context.Context, bookID, chapterID int64, item bookentity.Chapter) (bookentity.Chapter, error)
	DeleteChapter(ctx context.Context, bookID, chapterID int64) error
}
