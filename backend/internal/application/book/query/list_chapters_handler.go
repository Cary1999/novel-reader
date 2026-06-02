package query

import (
	"context"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
)

type ListChaptersHandler struct {
	getBook *GetBookHandler
	books   bookrepository.BookRepository
}

func NewListChaptersHandler(books bookrepository.BookRepository) *ListChaptersHandler {
	return &ListChaptersHandler{
		getBook: NewGetBookHandler(books),
		books:   books,
	}
}

func (h *ListChaptersHandler) Handle(ctx context.Context, bookID int64) ([]bookentity.Chapter, error) {
	if _, err := h.getBook.Handle(ctx, bookID); err != nil {
		return nil, err
	}
	return h.books.ListChapters(ctx, bookID)
}
