package query

import (
	"context"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
)

type ListRecommendedBooksHandler struct {
	books bookrepository.BookRepository
}

func NewListRecommendedBooksHandler(books bookrepository.BookRepository) *ListRecommendedBooksHandler {
	return &ListRecommendedBooksHandler{books: books}
}

func (h *ListRecommendedBooksHandler) Handle(ctx context.Context, page, pageSize int) ([]bookentity.Book, int, error) {
	query, err := normalizeSearchBooks(SearchBooks{Page: page, PageSize: pageSize})
	if err != nil {
		return nil, 0, err
	}
	return h.books.ListRecommendedBooks(ctx, query.Page, query.PageSize)
}
