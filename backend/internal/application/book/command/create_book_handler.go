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

type CreateBook struct {
	Title       string
	CategoryID  int64
	Description string
	OwnerUserID int64
}

type CreateBookHandler struct {
	booksRepo      bookrepository.BookRepository
	bookService    *bookservice.BookService
	categoriesRepo categoryrepository.CategoryRepository
}

func NewCreateBookHandler(books bookrepository.BookRepository, service *bookservice.BookService, categories categoryrepository.CategoryRepository) *CreateBookHandler {
	return &CreateBookHandler{booksRepo: books, bookService: service, categoriesRepo: categories}
}

func (h *CreateBookHandler) Handle(ctx context.Context, actor shared.Actor, input CreateBook) (bookentity.Book, error) {
	normalizedInput, err := normalizeCreateBook(input, actor.UserID)
	if err != nil {
		return bookentity.Book{}, err
	}
	categoryID, err := requireCategory(ctx, h.categoriesRepo, normalizedInput.CategoryID)
	if err != nil {
		return bookentity.Book{}, err
	}
	item := h.bookService.NewBook(normalizedInput.Title, normalizedInput.Description, normalizedInput.OwnerUserID)
	return h.booksRepo.CreateBook(ctx, item, categoryID)
}

func normalizeCreateBook(input CreateBook, ownerUserID int64) (CreateBook, error) {
	title := strings.TrimSpace(input.Title)
	description := strings.TrimSpace(input.Description)
	if title == "" {
		return CreateBook{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "title is required")
	}
	if len([]rune(title)) > 255 {
		return CreateBook{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "book title is too long")
	}
	if len([]rune(description)) > 5000 {
		return CreateBook{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "description is too long")
	}
	return CreateBook{
		Title:       title,
		CategoryID:  input.CategoryID,
		Description: description,
		OwnerUserID: ownerUserID,
	}, nil
}

func NormalizeCreateBookForActor(input CreateBook, ownerUserID int64) (CreateBook, error) {
	return normalizeCreateBook(input, ownerUserID)
}
