package service

import (
	"context"
	"net/http"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	"novel-reader/backend/internal/domain/shared"
)

type BookService struct {
	booksRepo bookrepository.BookRepository
}

func NewBookService(booksRepo bookrepository.BookRepository) *BookService {
	return &BookService{booksRepo: booksRepo}
}

func (s *BookService) NewBook(title, description string, ownerUserID int64) bookentity.Book {
	return bookentity.Book{
		Title:       title,
		Description: description,
		OwnerUserID: &ownerUserID,
	}
}

func (s *BookService) BuildMetadata(title, description string, recommendScore int) bookentity.Book {
	return bookentity.Book{
		Title:          title,
		Description:    description,
		RecommendScore: recommendScore,
	}
}

func (s *BookService) NewChapter(title, content string) bookentity.Chapter {
	return bookentity.Chapter{
		Title:   title,
		Content: content,
	}
}

func (s *BookService) CanManageBook(ctx context.Context, bookID int64, actor shared.Actor) error {
	item, err := s.booksRepo.FindBook(ctx, bookID)
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
