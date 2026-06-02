package query

import (
	"context"
	"net/http"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	"novel-reader/backend/internal/domain/shared"
)

type GetChapterHandler struct {
	books bookrepository.BookRepository
}

func NewGetChapterHandler(books bookrepository.BookRepository) *GetChapterHandler {
	return &GetChapterHandler{books: books}
}

func (h *GetChapterHandler) Handle(ctx context.Context, bookID, chapterID int64) (bookentity.Chapter, error) {
	item, err := h.books.FindChapter(ctx, bookID, chapterID)
	if err != nil {
		if err == shared.ErrNotFound {
			return bookentity.Chapter{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "chapter not found")
		}
		return bookentity.Chapter{}, err
	}
	return item, nil
}
