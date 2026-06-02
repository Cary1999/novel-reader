package bookapp

import (
	"context"
	"net/http"
	"strings"

	"novel-reader/backend/internal/domain/book"
	"novel-reader/backend/internal/domain/shared"
)

type QueryService struct {
	books book.Repository
}

func NewQueryService(books book.Repository) *QueryService {
	return &QueryService{books: books}
}

func (s *QueryService) SearchBooks(ctx context.Context, q, category string, page, pageSize int) ([]book.Book, int, error) {
	q = strings.TrimSpace(q)
	category = strings.TrimSpace(category)
	if len([]rune(q)) > 80 || len([]rune(category)) > 64 {
		return nil, 0, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "query is too long")
	}
	page, pageSize = normalizePage(page, pageSize)
	return s.books.SearchBooks(ctx, q, category, page, pageSize)
}

func (s *QueryService) ListRecommendedBooks(ctx context.Context, page, pageSize int) ([]book.Book, int, error) {
	page, pageSize = normalizePage(page, pageSize)
	return s.books.ListRecommendedBooks(ctx, page, pageSize)
}

func (s *QueryService) GetBook(ctx context.Context, id int64) (book.Book, error) {
	item, err := s.books.FindBook(ctx, id)
	if err != nil {
		if err == shared.ErrNotFound {
			return book.Book{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return book.Book{}, err
	}
	return item, nil
}

func (s *QueryService) ListChapters(ctx context.Context, bookID int64) ([]book.Chapter, error) {
	if _, err := s.GetBook(ctx, bookID); err != nil {
		return nil, err
	}
	return s.books.ListChapters(ctx, bookID)
}

func (s *QueryService) GetChapter(ctx context.Context, bookID, chapterID int64) (book.Chapter, error) {
	item, err := s.books.FindChapter(ctx, bookID, chapterID)
	if err != nil {
		if err == shared.ErrNotFound {
			return book.Chapter{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "chapter not found")
		}
		return book.Chapter{}, err
	}
	return item, nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}
