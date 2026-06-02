package query

import (
	"context"
	"net/http"
	"strings"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	"novel-reader/backend/internal/domain/shared"
)

type ListOwnedBooks struct {
	OwnerUserID int64
	Keyword     string
	CategoryID  int64
	Page        int
	PageSize    int
}

type ListOwnedBooksHandler struct {
	books bookrepository.BookRepository
}

func NewListOwnedBooksHandler(books bookrepository.BookRepository) *ListOwnedBooksHandler {
	return &ListOwnedBooksHandler{books: books}
}

func (h *ListOwnedBooksHandler) Handle(ctx context.Context, input ListOwnedBooks) ([]bookentity.Book, int, error) {
	query, err := normalizeListOwnedBooks(input)
	if err != nil {
		return nil, 0, err
	}
	return h.books.ListBooksByOwner(ctx, query.OwnerUserID, query.Keyword, query.CategoryID, query.Page, query.PageSize)
}

func normalizeListOwnedBooks(input ListOwnedBooks) (ListOwnedBooks, error) {
	keyword := strings.TrimSpace(input.Keyword)
	if len([]rune(keyword)) > 80 {
		return ListOwnedBooks{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "query is too long")
	}
	return ListOwnedBooks{
		OwnerUserID: input.OwnerUserID,
		Keyword:     keyword,
		CategoryID:  input.CategoryID,
		Page:        normalizePage(input.Page),
		PageSize:    normalizePageSize(input.PageSize),
	}, nil
}
