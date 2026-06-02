package command

import (
	"context"
	"net/http"

	bookrepository "novel-reader/backend/internal/domain/book/repository"
	bookservice "novel-reader/backend/internal/domain/book/service"
	"novel-reader/backend/internal/domain/shared"
)

type DeleteBookHandler struct {
	booksRepo   bookrepository.BookRepository
	bookService *bookservice.BookService
}

func NewDeleteBookHandler(books bookrepository.BookRepository, service *bookservice.BookService) *DeleteBookHandler {
	return &DeleteBookHandler{booksRepo: books, bookService: service}
}

func (h *DeleteBookHandler) Handle(ctx context.Context, actor shared.Actor, bookID int64) error {
	if err := h.bookService.CanManageBook(ctx, bookID, actor); err != nil {
		return err
	}
	if err := h.booksRepo.DeleteBook(ctx, bookID); err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return err
	}
	return nil
}
