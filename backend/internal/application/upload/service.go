package uploadapp

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"novel-reader/backend/internal/domain/book"
	"novel-reader/backend/internal/domain/category"
	"novel-reader/backend/internal/domain/shared"
	"novel-reader/backend/internal/domain/upload"
)

type FileStore interface {
	SaveTXT(originalName string, reader io.Reader) (SavedFile, error)
}

type CoverStore interface {
	SaveCover(originalName string, reader io.Reader) (SavedFile, error)
}

type ChapterParser interface {
	ParseChapters(text string) ([]book.ChapterDraft, error)
}

type SavedFile struct {
	RelativePath string
	Size         int64
	Content      string
}

type Service struct {
	books      book.Repository
	categories category.Repository
	uploads    upload.Repository
	files      FileStore
	covers     CoverStore
	parser     ChapterParser
}

func NewService(books book.Repository, categories category.Repository, uploads upload.Repository, files FileStore, covers CoverStore, parser ChapterParser) *Service {
	return &Service{
		books:      books,
		categories: categories,
		uploads:    uploads,
		files:      files,
		covers:     covers,
		parser:     parser,
	}
}

func (s *Service) UploadBook(ctx context.Context, actor shared.Actor, input book.CreateInput, originalName string, reader io.Reader) (upload.ImportResult, error) {
	input.OwnerUserID = actor.UserID
	categoryID, err := s.normalizeCreateInput(ctx, &input)
	if err != nil {
		return upload.ImportResult{}, err
	}

	saved, err := s.files.SaveTXT(originalName, reader)
	if err != nil {
		return upload.ImportResult{}, uploadError(err)
	}

	uploadID, err := s.uploads.CreateUpload(ctx, upload.Upload{
		ActorUserID:      actor.UserID,
		OriginalFilename: originalName,
		StoredPath:       saved.RelativePath,
		FileSize:         saved.Size,
		Status:           "uploaded",
	})
	if err != nil {
		return upload.ImportResult{}, err
	}

	chapters, err := s.parser.ParseChapters(saved.Content)
	if err != nil {
		_ = s.uploads.MarkUpload(ctx, uploadID, "failed", "no chapters found")
		return upload.ImportResult{}, shared.NewError(http.StatusBadRequest, "PARSE_NO_CHAPTERS", "no chapters found in txt file")
	}
	bookID, err := s.books.CreateBookWithChapters(ctx, input, categoryID, uploadID, chapters)
	if err != nil {
		_ = s.uploads.MarkUpload(ctx, uploadID, "failed", "book create failed")
		return upload.ImportResult{}, err
	}
	if err := s.uploads.MarkUpload(ctx, uploadID, "parsed", ""); err != nil {
		return upload.ImportResult{}, err
	}

	return upload.ImportResult{
		BookID:            bookID,
		ChapterCount:      len(chapters),
		UploadID:          uploadID,
		FirstChapterTitle: chapters[0].Title,
		LastChapterTitle:  chapters[len(chapters)-1].Title,
	}, nil
}

func (s *Service) UploadCover(ctx context.Context, actor shared.Actor, bookID int64, originalName string, reader io.Reader) (string, error) {
	if err := s.authorizeBookWrite(ctx, actor, bookID); err != nil {
		return "", err
	}
	saved, err := s.covers.SaveCover(originalName, reader)
	if err != nil {
		return "", uploadCoverError(err)
	}
	if err := s.books.UpdateBookCoverPath(ctx, bookID, &saved.RelativePath); err != nil {
		if err == shared.ErrNotFound {
			return "", shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return "", err
	}
	return "/api/books/" + strconv.FormatInt(bookID, 10) + "/cover", nil
}

func (s *Service) authorizeBookWrite(ctx context.Context, actor shared.Actor, bookID int64) error {
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

func (s *Service) normalizeCreateInput(ctx context.Context, input *book.CreateInput) (*int64, error) {
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
	if input.CategoryID <= 0 {
		return nil, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "category is required")
	}
	if _, err := s.categories.FindCategoryByID(ctx, input.CategoryID); err != nil {
		if err == shared.ErrNotFound {
			return nil, shared.NewError(http.StatusNotFound, "NOT_FOUND", "category not found")
		}
		return nil, err
	}
	return &input.CategoryID, nil
}

func uploadError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "invalid file type"):
		return shared.NewError(http.StatusBadRequest, "UPLOAD_INVALID_TYPE", "only .txt files are allowed")
	case strings.Contains(message, "file too large"):
		return shared.NewError(http.StatusRequestEntityTooLarge, "UPLOAD_TOO_LARGE", "txt file is too large")
	case strings.Contains(message, "empty file"):
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "txt file is empty")
	default:
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid upload")
	}
}

func uploadCoverError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "invalid file type"):
		return shared.NewError(http.StatusBadRequest, "UPLOAD_INVALID_TYPE", "only jpg/png/webp files are allowed")
	case strings.Contains(message, "file too large"):
		return shared.NewError(http.StatusRequestEntityTooLarge, "UPLOAD_TOO_LARGE", "cover image is too large")
	case strings.Contains(message, "empty file"):
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "cover image is empty")
	default:
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid upload")
	}
}
