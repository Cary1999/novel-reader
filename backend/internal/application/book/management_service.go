package bookapp

import (
	"context"
	"net/http"
	"strings"

	"novel-reader/backend/internal/domain/book"
	"novel-reader/backend/internal/domain/category"
	"novel-reader/backend/internal/domain/shared"
)

type ManagementService struct {
	books      book.Repository
	categories category.Repository
}

func NewManagementService(books book.Repository, categories category.Repository) *ManagementService {
	return &ManagementService{books: books, categories: categories}
}

func (s *ManagementService) ListBooks(ctx context.Context, q, categoryName string, page, pageSize int) ([]book.Book, int, error) {
	q = strings.TrimSpace(q)
	categoryName = strings.TrimSpace(categoryName)
	if len([]rune(q)) > 80 || len([]rune(categoryName)) > 64 {
		return nil, 0, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "query is too long")
	}
	page, pageSize = normalizePage(page, pageSize)
	return s.books.SearchBooks(ctx, q, categoryName, page, pageSize)
}

func (s *ManagementService) ListMyBooks(ctx context.Context, ownerID int64, q string, categoryID int64, page, pageSize int) ([]book.Book, int, error) {
	q = strings.TrimSpace(q)
	if len([]rune(q)) > 80 {
		return nil, 0, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "query is too long")
	}
	page, pageSize = normalizePage(page, pageSize)
	return s.books.ListBooksByOwner(ctx, ownerID, q, categoryID, page, pageSize)
}

func (s *ManagementService) CreateBook(ctx context.Context, actor shared.Actor, input book.CreateInput) (book.Book, error) {
	input.OwnerUserID = actor.UserID
	categoryID, err := s.normalizeCreateInput(ctx, &input)
	if err != nil {
		return book.Book{}, err
	}
	return s.books.CreateBook(ctx, input, categoryID)
}

func (s *ManagementService) UpdateBook(ctx context.Context, actor shared.Actor, bookID int64, input book.MetadataInput) (book.Book, error) {
	if err := s.authorizeBookWrite(ctx, actor, bookID); err != nil {
		return book.Book{}, err
	}
	categoryID, err := s.normalizeMetadataInput(ctx, &input)
	if err != nil {
		return book.Book{}, err
	}
	item, err := s.books.UpdateBookMetadata(ctx, bookID, input, categoryID)
	if err != nil {
		if err == shared.ErrNotFound {
			return book.Book{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return book.Book{}, err
	}
	return item, nil
}

func (s *ManagementService) DeleteBook(ctx context.Context, actor shared.Actor, bookID int64) error {
	if err := s.authorizeBookWrite(ctx, actor, bookID); err != nil {
		return err
	}
	err := s.books.DeleteBook(ctx, bookID)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return err
	}
	return nil
}

func (s *ManagementService) AddChapter(ctx context.Context, actor shared.Actor, bookID int64, input book.ChapterInput) (book.Chapter, error) {
	if err := s.authorizeBookWrite(ctx, actor, bookID); err != nil {
		return book.Chapter{}, err
	}
	if err := normalizeChapterInput(&input); err != nil {
		return book.Chapter{}, err
	}
	item, err := s.books.AddChapter(ctx, bookID, input)
	if err != nil {
		if err == shared.ErrNotFound {
			return book.Chapter{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return book.Chapter{}, err
	}
	return item, nil
}

func (s *ManagementService) UpdateChapter(ctx context.Context, actor shared.Actor, bookID, chapterID int64, input book.ChapterInput) (book.Chapter, error) {
	if err := s.authorizeBookWrite(ctx, actor, bookID); err != nil {
		return book.Chapter{}, err
	}
	if err := normalizeChapterInput(&input); err != nil {
		return book.Chapter{}, err
	}
	item, err := s.books.UpdateChapter(ctx, bookID, chapterID, input)
	if err != nil {
		if err == shared.ErrNotFound {
			return book.Chapter{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "chapter not found")
		}
		return book.Chapter{}, err
	}
	return item, nil
}

func (s *ManagementService) DeleteChapter(ctx context.Context, actor shared.Actor, bookID, chapterID int64) error {
	if err := s.authorizeBookWrite(ctx, actor, bookID); err != nil {
		return err
	}
	err := s.books.DeleteChapter(ctx, bookID, chapterID)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusNotFound, "NOT_FOUND", "chapter not found")
		}
		return err
	}
	return nil
}

func (s *ManagementService) authorizeBookWrite(ctx context.Context, actor shared.Actor, bookID int64) error {
	item, err := s.books.FindBook(ctx, bookID)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return err
	}
	if actor.Role == shared.RoleAdmin {
		return nil
	}
	if item.OwnerUserID == nil || *item.OwnerUserID != actor.UserID {
		return shared.NewError(http.StatusForbidden, "FORBIDDEN", "only the author can manage this book")
	}
	return nil
}

func (s *ManagementService) normalizeCreateInput(ctx context.Context, input *book.CreateInput) (*int64, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return nil, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "title is required")
	}
	if len([]rune(input.Title)) > 255 {
		return nil, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "book title is too long")
	}
	if len([]rune(input.Description)) > 5000 {
		return nil, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "description is too long")
	}
	return s.requireCategory(ctx, input.CategoryID)
}

func (s *ManagementService) normalizeMetadataInput(ctx context.Context, input *book.MetadataInput) (*int64, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return nil, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "title is required")
	}
	if len([]rune(input.Title)) > 255 {
		return nil, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "book title is too long")
	}
	if len([]rune(input.Description)) > 5000 {
		return nil, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "description is too long")
	}
	return s.requireCategory(ctx, input.CategoryID)
}

func (s *ManagementService) requireCategory(ctx context.Context, categoryID int64) (*int64, error) {
	if categoryID <= 0 {
		return nil, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "category is required")
	}
	if _, err := s.categories.FindCategoryByID(ctx, categoryID); err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		return nil, err
	}
	return &categoryID, nil
}

func normalizeChapterInput(input *book.ChapterInput) error {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if input.Title == "" || input.Content == "" {
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "chapter title and content are required")
	}
	if len([]rune(input.Title)) > 255 {
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "chapter title is too long")
	}
	return nil
}
