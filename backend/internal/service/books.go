package service

import (
	"context"
	"net/http"
	"strings"

	"novel-reader/backend/internal/domain"
	"novel-reader/backend/internal/repository"
)

type BookService struct {
	store repository.Store
}

func NewBookService(store repository.Store) *BookService {
	return &BookService{store: store}
}

func (s *BookService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.store.ListCategories(ctx)
}

func (s *BookService) SearchBooks(ctx context.Context, q, category string, page, pageSize int) ([]domain.Book, int, error) {
	q = strings.TrimSpace(q)
	category = strings.TrimSpace(category)
	if len([]rune(q)) > 80 || len([]rune(category)) > 64 {
		return nil, 0, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "query is too long")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return s.store.SearchBooks(ctx, q, category, page, pageSize)
}

func (s *BookService) ListRecommendedBooks(ctx context.Context, page, pageSize int) ([]domain.Book, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return s.store.ListRecommendedBooks(ctx, page, pageSize)
}

func (s *BookService) GetBook(ctx context.Context, id int64) (domain.Book, error) {
	book, err := s.store.FindBook(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Book{}, domain.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return domain.Book{}, err
	}
	return book, nil
}

func (s *BookService) ListChapters(ctx context.Context, bookID int64) ([]domain.Chapter, error) {
	if _, err := s.GetBook(ctx, bookID); err != nil {
		return nil, err
	}
	return s.store.ListChapters(ctx, bookID)
}

func (s *BookService) GetChapter(ctx context.Context, bookID, chapterID int64) (domain.Chapter, error) {
	chapter, err := s.store.FindChapter(ctx, bookID, chapterID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Chapter{}, domain.NewError(http.StatusNotFound, "NOT_FOUND", "chapter not found")
		}
		return domain.Chapter{}, err
	}
	return chapter, nil
}
