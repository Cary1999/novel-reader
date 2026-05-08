package service

import (
	"context"
	"io"
	"net/http"
	"strings"

	"novel-reader/backend/internal/domain"
	"novel-reader/backend/internal/parser"
	"novel-reader/backend/internal/repository"
	"novel-reader/backend/internal/storage"
)

type FileStore interface {
	SaveTXT(originalName string, reader io.Reader) (storage.SavedFile, error)
}

type AdminService struct {
	store repository.Store
	files FileStore
}

func NewAdminService(store repository.Store, files FileStore) *AdminService {
	return &AdminService{store: store, files: files}
}

func (s *AdminService) ListBooks(ctx context.Context, q, category string, page, pageSize int) ([]domain.Book, int, error) {
	bookSvc := NewBookService(s.store)
	return bookSvc.SearchBooks(ctx, q, category, page, pageSize)
}

func (s *AdminService) ListMyBooks(ctx context.Context, ownerID int64, q string, categoryID int64, page, pageSize int) ([]domain.Book, int, error) {
	q = strings.TrimSpace(q)
	if len([]rune(q)) > 80 {
		return nil, 0, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "query is too long")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.store.ListBooksByOwner(ctx, ownerID, q, categoryID, page, pageSize)
}

func (s *AdminService) CreateBook(ctx context.Context, ownerID int64, input domain.UploadBookInput) (domain.Book, error) {
	input.OwnerUserID = ownerID
	categoryID, err := s.normalizeCreateBookInput(ctx, &input)
	if err != nil {
		return domain.Book{}, err
	}
	book, err := s.store.CreateBook(ctx, input, categoryID)
	if err != nil {
		return domain.Book{}, err
	}
	return book, nil
}

func (s *AdminService) UploadBook(ctx context.Context, ownerID int64, input domain.UploadBookInput, originalName string, reader io.Reader) (domain.UploadResult, error) {
	input.OwnerUserID = ownerID
	categoryID, err := s.normalizeCreateBookInput(ctx, &input)
	if err != nil {
		return domain.UploadResult{}, err
	}

	saved, err := s.files.SaveTXT(originalName, reader)
	if err != nil {
		return domain.UploadResult{}, uploadError(err)
	}

	uploadID, err := s.store.CreateUpload(ctx, domain.Upload{
		AdminUserID:      ownerID,
		OriginalFilename: originalName,
		StoredPath:       saved.RelativePath,
		FileSize:         saved.Size,
		Status:           "uploaded",
	})
	if err != nil {
		return domain.UploadResult{}, err
	}

	chapters, err := parser.ParseChapters(saved.Content)
	if err != nil {
		_ = s.store.MarkUpload(ctx, uploadID, "failed", "no chapters found")
		return domain.UploadResult{}, domain.NewError(http.StatusBadRequest, "PARSE_NO_CHAPTERS", "no chapters found in txt file")
	}
	bookID, err := s.store.CreateBookWithChapters(ctx, input, categoryID, uploadID, chapters)
	if err != nil {
		_ = s.store.MarkUpload(ctx, uploadID, "failed", "book create failed")
		return domain.UploadResult{}, err
	}
	if err := s.store.MarkUpload(ctx, uploadID, "parsed", ""); err != nil {
		return domain.UploadResult{}, err
	}

	return domain.UploadResult{
		BookID:            bookID,
		ChapterCount:      len(chapters),
		UploadID:          uploadID,
		FirstChapterTitle: chapters[0].Title,
		LastChapterTitle:  chapters[len(chapters)-1].Title,
	}, nil
}

func (s *AdminService) UpdateBook(ctx context.Context, actorID int64, role domain.Role, bookID int64, input domain.BookMetadataInput) (domain.Book, error) {
	if err := s.authorizeBookWrite(ctx, actorID, role, bookID); err != nil {
		return domain.Book{}, err
	}
	categoryID, err := s.normalizeBookMetadata(ctx, &input)
	if err != nil {
		return domain.Book{}, err
	}
	book, err := s.store.UpdateBookMetadata(ctx, bookID, input, categoryID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Book{}, domain.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return domain.Book{}, err
	}
	return book, nil
}

func (s *AdminService) DeleteBook(ctx context.Context, actorID int64, role domain.Role, bookID int64) error {
	if err := s.authorizeBookWrite(ctx, actorID, role, bookID); err != nil {
		return err
	}
	if err := s.store.DeleteBook(ctx, bookID); err != nil {
		if repository.IsNotFound(err) {
			return domain.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return err
	}
	return nil
}

func (s *AdminService) AddChapter(ctx context.Context, actorID int64, role domain.Role, bookID int64, input domain.ChapterInput) (domain.Chapter, error) {
	if err := s.authorizeBookWrite(ctx, actorID, role, bookID); err != nil {
		return domain.Chapter{}, err
	}
	if err := normalizeChapterInput(&input); err != nil {
		return domain.Chapter{}, err
	}
	chapter, err := s.store.AddChapter(ctx, bookID, input)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Chapter{}, domain.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return domain.Chapter{}, err
	}
	return chapter, nil
}

func (s *AdminService) UpdateChapter(ctx context.Context, actorID int64, role domain.Role, bookID, chapterID int64, input domain.ChapterInput) (domain.Chapter, error) {
	if err := s.authorizeBookWrite(ctx, actorID, role, bookID); err != nil {
		return domain.Chapter{}, err
	}
	if err := normalizeChapterInput(&input); err != nil {
		return domain.Chapter{}, err
	}
	chapter, err := s.store.UpdateChapter(ctx, bookID, chapterID, input)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Chapter{}, domain.NewError(http.StatusNotFound, "NOT_FOUND", "chapter not found")
		}
		return domain.Chapter{}, err
	}
	return chapter, nil
}

func (s *AdminService) DeleteChapter(ctx context.Context, actorID int64, role domain.Role, bookID, chapterID int64) error {
	if err := s.authorizeBookWrite(ctx, actorID, role, bookID); err != nil {
		return err
	}
	if err := s.store.DeleteChapter(ctx, bookID, chapterID); err != nil {
		if repository.IsNotFound(err) {
			return domain.NewError(http.StatusNotFound, "NOT_FOUND", "chapter not found")
		}
		return err
	}
	return nil
}

func (s *AdminService) CreateCategory(ctx context.Context, name string) (domain.Category, error) {
	name, err := normalizeCategoryName(name)
	if err != nil {
		return domain.Category{}, err
	}
	category, err := s.store.CreateCategory(ctx, name)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return domain.Category{}, domain.NewError(http.StatusConflict, "CONFLICT", "category already exists")
		}
		return domain.Category{}, err
	}
	return category, nil
}

func (s *AdminService) UpdateCategory(ctx context.Context, id int64, name string) (domain.Category, error) {
	name, err := normalizeCategoryName(name)
	if err != nil {
		return domain.Category{}, err
	}
	category, err := s.store.UpdateCategory(ctx, id, name)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.Category{}, domain.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return domain.Category{}, domain.NewError(http.StatusConflict, "CONFLICT", "category already exists")
		}
		return domain.Category{}, err
	}
	return category, nil
}

func (s *AdminService) DeleteCategory(ctx context.Context, id int64) error {
	if err := s.store.DeleteCategory(ctx, id); err != nil {
		if repository.IsNotFound(err) {
			return domain.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		return err
	}
	return nil
}

func (s *AdminService) authorizeBookWrite(ctx context.Context, actorID int64, role domain.Role, bookID int64) error {
	book, err := s.store.FindBook(ctx, bookID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return err
	}
	if role == domain.RoleAdmin {
		return nil
	}
	if book.OwnerUserID == nil || *book.OwnerUserID != actorID {
		return domain.NewError(http.StatusForbidden, "FORBIDDEN", "only the author can manage this book")
	}
	return nil
}

func (s *AdminService) normalizeCreateBookInput(ctx context.Context, input *domain.UploadBookInput) (*int64, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return nil, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "title is required")
	}
	if len([]rune(input.Title)) > 255 {
		return nil, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "book title is too long")
	}
	if len([]rune(input.Description)) > 5000 {
		return nil, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "description is too long")
	}
	return s.requireCategory(ctx, input.CategoryID)
}

func (s *AdminService) normalizeBookMetadata(ctx context.Context, input *domain.BookMetadataInput) (*int64, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" {
		return nil, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "title is required")
	}
	if len([]rune(input.Title)) > 255 {
		return nil, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "book title is too long")
	}
	if len([]rune(input.Description)) > 5000 {
		return nil, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "description is too long")
	}
	return s.requireCategory(ctx, input.CategoryID)
}

func (s *AdminService) requireCategory(ctx context.Context, categoryID int64) (*int64, error) {
	if categoryID <= 0 {
		return nil, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "category is required")
	}
	if _, err := s.store.FindCategoryByID(ctx, categoryID); err != nil {
		if repository.IsNotFound(err) {
			return nil, domain.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		return nil, err
	}
	return &categoryID, nil
}

func normalizeCategoryName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "category name is required")
	}
	if len([]rune(name)) > 64 {
		return "", domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "category name is too long")
	}
	return name, nil
}

func normalizeChapterInput(input *domain.ChapterInput) error {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if input.Title == "" || input.Content == "" {
		return domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "chapter title and content are required")
	}
	if len([]rune(input.Title)) > 255 {
		return domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "chapter title is too long")
	}
	return nil
}

func uploadError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "invalid file type"):
		return domain.NewError(http.StatusBadRequest, "UPLOAD_INVALID_TYPE", "only .txt files are allowed")
	case strings.Contains(message, "file too large"):
		return domain.NewError(http.StatusRequestEntityTooLarge, "UPLOAD_TOO_LARGE", "txt file is too large")
	case strings.Contains(message, "empty file"):
		return domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "txt file is empty")
	default:
		return domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid upload")
	}
}
