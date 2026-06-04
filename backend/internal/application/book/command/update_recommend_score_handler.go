package command

import (
	"context"
	"net/http"

	bookentity "novel-reader/backend/internal/domain/book/entity"
	bookrepository "novel-reader/backend/internal/domain/book/repository"
	"novel-reader/backend/internal/domain/shared"
)

type UpdateRecommendScore struct {
	RecommendScore int `json:"recommendScore"`
}

type UpdateRecommendScoreHandler struct {
	booksRepo bookrepository.BookRepository
}

func NewUpdateRecommendScoreHandler(books bookrepository.BookRepository) *UpdateRecommendScoreHandler {
	return &UpdateRecommendScoreHandler{booksRepo: books}
}

func (h *UpdateRecommendScoreHandler) Handle(ctx context.Context, actor shared.Actor, bookID int64, input UpdateRecommendScore) (bookentity.Book, error) {
	if !actor.IsReviewer() {
		return bookentity.Book{}, shared.NewError(http.StatusForbidden, "FORBIDDEN", "reviewer required")
	}
	if input.RecommendScore < 0 {
		return bookentity.Book{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "recommend score must be greater than or equal to 0")
	}
	item, err := h.booksRepo.UpdateRecommendScore(ctx, bookID, input.RecommendScore)
	if err != nil {
		if err == shared.ErrNotFound {
			return bookentity.Book{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "book not found")
		}
		return bookentity.Book{}, err
	}
	return item, nil
}
