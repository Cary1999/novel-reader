package query

import (
	"context"
	"net/http"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	"novel-reader/backend/internal/domain/shared"
)

type GetBookHandler struct {
	books bookrepository.BookRepository
}

func NewGetBookHandler(books bookrepository.BookRepository) *GetBookHandler {
	return &GetBookHandler{books: books}
}

func (h *GetBookHandler) Handle(ctx context.Context, id int64) (bookentity.Book, error) {
	item, err := h.books.FindBook(ctx, id)
	if err != nil {
		if err == shared.ErrNotFound {
			return bookentity.Book{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return bookentity.Book{}, err
	}
	return item, nil
}
