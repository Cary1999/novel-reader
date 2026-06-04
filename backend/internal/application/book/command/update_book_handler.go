package command

import (
	"context"
	"net/http"
	"strings"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	bookservice "novel-reader/backend/internal/domain/book/service"
	categoryrepository "novel-reader/backend/internal/domain/category/repository"
	"novel-reader/backend/internal/domain/shared"
)

type UpdateBookMetadata struct {
	Title          string `json:"title"`
	CategoryID     int64  `json:"categoryId"`
	Description    string `json:"description"`
	RecommendScore *int   `json:"recommendScore,omitempty"`
}

type UpdateBookHandler struct {
	booksRepo      bookrepository.BookRepository
	bookService    *bookservice.BookService
	categoriesRepo categoryrepository.CategoryRepository
}

func NewUpdateBookHandler(books bookrepository.BookRepository, service *bookservice.BookService, categories categoryrepository.CategoryRepository) *UpdateBookHandler {
	return &UpdateBookHandler{booksRepo: books, bookService: service, categoriesRepo: categories}
}

func (h *UpdateBookHandler) Handle(ctx context.Context, actor shared.Actor, bookID int64, input UpdateBookMetadata) (bookentity.Book, error) {
	if err := h.bookService.CanManageBook(ctx, bookID, actor); err != nil {
		return bookentity.Book{}, err
	}
	normalizedInput, err := normalizeUpdateBook(input)
	if err != nil {
		return bookentity.Book{}, err
	}
	categoryID, err := requireCategory(ctx, h.categoriesRepo, normalizedInput.CategoryID)
	if err != nil {
		return bookentity.Book{}, err
	}
	current, err := h.booksRepo.FindBook(ctx, bookID)
	if err != nil {
		if err == shared.ErrNotFound {
			return bookentity.Book{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return bookentity.Book{}, err
	}
	if normalizedInput.RecommendScore != nil {
		return bookentity.Book{}, shared.NewError(http.StatusForbidden, "FORBIDDEN", "recommend score is read only")
	}
	recommendScore := current.RecommendScore
	item, err := h.booksRepo.UpdateBookMetadata(ctx, bookID, h.bookService.BuildMetadata(normalizedInput.Title, normalizedInput.Description, recommendScore), categoryID)
	if err != nil {
		if err == shared.ErrNotFound {
			return bookentity.Book{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return bookentity.Book{}, err
	}
	return item, nil
}

func normalizeUpdateBook(input UpdateBookMetadata) (UpdateBookMetadata, error) {
	title := strings.TrimSpace(input.Title)
	description := strings.TrimSpace(input.Description)
	if title == "" {
		return UpdateBookMetadata{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "title is required")
	}
	if len([]rune(title)) > 255 {
		return UpdateBookMetadata{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "book title is too long")
	}
	if len([]rune(description)) > 5000 {
		return UpdateBookMetadata{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "description is too long")
	}
	return UpdateBookMetadata{
		Title:          title,
		CategoryID:     input.CategoryID,
		Description:    description,
		RecommendScore: input.RecommendScore,
	}, nil
}
