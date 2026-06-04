package query

import (
	"context"
	"net/http"
	"strings"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	"novel-reader/backend/internal/domain/shared"
)

type SearchBooks struct {
	Keyword      string
	CategoryName string
	Page         int
	PageSize     int
}

type SearchBooksHandler struct {
	books bookrepository.BookRepository
}

func NewSearchBooksHandler(books bookrepository.BookRepository) *SearchBooksHandler {
	return &SearchBooksHandler{books: books}
}

func (h *SearchBooksHandler) Handle(ctx context.Context, input SearchBooks) ([]bookentity.Book, int, error) {
	query, err := normalizeSearchBooks(input)
	if err != nil {
		return nil, 0, err
	}
	return h.books.SearchBooks(ctx, query.Keyword, query.CategoryName, query.Page, query.PageSize)
}

func normalizeSearchBooks(input SearchBooks) (SearchBooks, error) {
	keyword := strings.TrimSpace(input.Keyword)
	categoryName := strings.TrimSpace(input.CategoryName)
	if len([]rune(keyword)) > 80 || len([]rune(categoryName)) > 64 {
		return SearchBooks{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "query is too long")
	}
	return SearchBooks{
		Keyword:      keyword,
		CategoryName: categoryName,
		Page:         normalizePage(input.Page),
		PageSize:     normalizePageSize(input.PageSize),
	}, nil
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}
