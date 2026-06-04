package uploadapp

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	bookcommand "novel-reader/backend/internal/application/book/command"
	uploadquery "novel-reader/backend/internal/application/upload/query"
	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	bookservice "novel-reader/backend/internal/domain/book/service"
	categoryrepository "novel-reader/backend/internal/domain/category/repository"
	"novel-reader/backend/internal/domain/shared"
	uploadentity "novel-reader/backend/internal/domain/upload/entity"
	uploadrepository "novel-reader/backend/internal/domain/upload/repository"
)

type FileStore interface {
	SaveTXT(originalName string, reader io.Reader) (SavedFile, error)
}

type CoverStore interface {
	SaveCover(originalName string, reader io.Reader) (SavedFile, error)
}

type ChapterParser interface {
	ParseChapters(text string) ([]bookentity.ChapterDraft, error)
}

type SavedFile struct {
	RelativePath string
	Size         int64
	Content      string
}

type Service struct {
	books       bookrepository.BookRepository
	bookService *bookservice.BookService
	categories  categoryrepository.CategoryRepository
	uploads     uploadrepository.UploadRepository
	files       FileStore
	covers      CoverStore
	parser      ChapterParser
}

func NewService(books bookrepository.BookRepository, categories categoryrepository.CategoryRepository, uploads uploadrepository.UploadRepository, files FileStore, covers CoverStore, parser ChapterParser) *Service {
	return &Service{
		books:       books,
		bookService: bookservice.NewBookService(books),
		categories:  categories,
		uploads:     uploads,
		files:       files,
		covers:      covers,
		parser:      parser,
	}
}

func (s *Service) UploadBook(ctx context.Context, actor shared.Actor, input bookcommand.CreateBook, originalName string, reader io.Reader) (uploadquery.ImportResult, error) {
	if !actor.IsAuthor() {
		return uploadquery.ImportResult{}, shared.NewError(http.StatusForbidden, "FORBIDDEN", "author required")
	}
	normalizedInput, err := bookcommand.NormalizeCreateBookForActor(input, actor.ActorID)
	if err != nil {
		return uploadquery.ImportResult{}, err
	}
	categoryID, err := s.requireCategory(ctx, normalizedInput.CategoryID)
	if err != nil {
		return uploadquery.ImportResult{}, err
	}
	bookItem := s.bookService.NewBook(normalizedInput.Title, normalizedInput.Description, normalizedInput.OwnerUserID)

	saved, err := s.files.SaveTXT(originalName, reader)
	if err != nil {
		return uploadquery.ImportResult{}, uploadError(err)
	}

	uploadID, err := s.uploads.CreateUpload(ctx, uploadentity.Upload{
		ActorUserID:      actor.ActorID,
		OriginalFilename: originalName,
		StoredPath:       saved.RelativePath,
		FileSize:         saved.Size,
		Status:           "uploaded",
	})
	if err != nil {
		return uploadquery.ImportResult{}, err
	}

	chapters, err := s.parser.ParseChapters(saved.Content)
	if err != nil {
		_ = s.uploads.MarkUpload(ctx, uploadID, "failed", "no chapters found")
		return uploadquery.ImportResult{}, shared.NewError(http.StatusBadRequest, "PARSE_NO_CHAPTERS", "no chapters found in txt file")
	}
	bookID, err := s.books.CreateBookWithChapters(ctx, bookItem, categoryID, uploadID, chapters)
	if err != nil {
		_ = s.uploads.MarkUpload(ctx, uploadID, "failed", "book create failed")
		return uploadquery.ImportResult{}, err
	}
	if err := s.uploads.MarkUpload(ctx, uploadID, "parsed", ""); err != nil {
		return uploadquery.ImportResult{}, err
	}

	return uploadquery.ImportResult{
		BookID:            bookID,
		ChapterCount:      len(chapters),
		UploadID:          uploadID,
		FirstChapterTitle: chapters[0].Title,
		LastChapterTitle:  chapters[len(chapters)-1].Title,
	}, nil
}

func (s *Service) UploadCover(ctx context.Context, actor shared.Actor, bookID int64, originalName string, reader io.Reader) (string, error) {
	if err := s.bookService.CanManageBook(ctx, bookID, actor); err != nil {
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

func (s *Service) requireCategory(ctx context.Context, categoryID int64) (*int64, error) {
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
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", message)
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
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", message)
	}
}
