package command

import (
	"context"
	"net/http"
	"strings"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	bookservice "novel-reader/backend/internal/domain/book/service"
	"novel-reader/backend/internal/domain/shared"
)

type SaveChapter struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type AddChapterHandler struct {
	booksRepo   bookrepository.BookRepository
	bookService *bookservice.BookService
}

func NewAddChapterHandler(books bookrepository.BookRepository, service *bookservice.BookService) *AddChapterHandler {
	return &AddChapterHandler{booksRepo: books, bookService: service}
}

func (h *AddChapterHandler) Handle(ctx context.Context, actor shared.Actor, bookID int64, input SaveChapter) (bookentity.Chapter, error) {
	if err := h.bookService.CanManageBook(ctx, bookID, actor); err != nil {
		return bookentity.Chapter{}, err
	}
	normalizedInput, err := normalizeSaveChapter(input)
	if err != nil {
		return bookentity.Chapter{}, err
	}
	item, err := h.booksRepo.AddChapter(ctx, bookID, h.bookService.NewChapter(normalizedInput.Title, normalizedInput.Content))
	if err != nil {
		if err == shared.ErrNotFound {
			return bookentity.Chapter{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return bookentity.Chapter{}, err
	}
	return item, nil
}

func normalizeSaveChapter(input SaveChapter) (SaveChapter, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" || content == "" {
		return SaveChapter{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "chapter title and content are required")
	}
	if len([]rune(title)) > 255 {
		return SaveChapter{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "chapter title is too long")
	}
	return SaveChapter{Title: title, Content: content}, nil
}
