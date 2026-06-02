package command

import (
	"context"
	"net/http"

	bookrepository "novel-reader/backend/internal/domain/book/repository"
	bookservice "novel-reader/backend/internal/domain/book/service"
	"novel-reader/backend/internal/domain/shared"
)

type DeleteChapterHandler struct {
	booksRepo   bookrepository.BookRepository
	bookService *bookservice.BookService
}

func NewDeleteChapterHandler(books bookrepository.BookRepository, service *bookservice.BookService) *DeleteChapterHandler {
	return &DeleteChapterHandler{booksRepo: books, bookService: service}
}

func (h *DeleteChapterHandler) Handle(ctx context.Context, actor shared.Actor, bookID, chapterID int64) error {
	if err := h.bookService.CanManageBook(ctx, bookID, actor); err != nil {
		return err
	}
	if err := h.booksRepo.DeleteChapter(ctx, bookID, chapterID); err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusNotFound, "NOT_FOUND", "chapter not found")
		}
		return err
	}
	return nil
}
