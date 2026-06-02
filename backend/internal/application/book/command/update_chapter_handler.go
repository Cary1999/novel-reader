package command

import (
	"context"
	"net/http"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	bookservice "novel-reader/backend/internal/domain/book/service"
	"novel-reader/backend/internal/domain/shared"
)

type UpdateChapterHandler struct {
	booksRepo   bookrepository.BookRepository
	bookService *bookservice.BookService
}

func NewUpdateChapterHandler(books bookrepository.BookRepository, service *bookservice.BookService) *UpdateChapterHandler {
	return &UpdateChapterHandler{booksRepo: books, bookService: service}
}

func (h *UpdateChapterHandler) Handle(ctx context.Context, actor shared.Actor, bookID, chapterID int64, input SaveChapter) (bookentity.Chapter, error) {
	if err := h.bookService.CanManageBook(ctx, bookID, actor); err != nil {
		return bookentity.Chapter{}, err
	}
	normalizedInput, err := normalizeSaveChapter(input)
	if err != nil {
		return bookentity.Chapter{}, err
	}
	item, err := h.booksRepo.UpdateChapter(ctx, bookID, chapterID, h.bookService.NewChapter(normalizedInput.Title, normalizedInput.Content))
	if err != nil {
		if err == shared.ErrNotFound {
			return bookentity.Chapter{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "chapter not found")
		}
		return bookentity.Chapter{}, err
	}
	return item, nil
}
