package bookshelfapp

import (
	"context"

	bookshelfentity "novel-reader/backend/internal/domain/bookshelf/entity"
	bookshelfrepository "novel-reader/backend/internal/domain/bookshelf/repository"
	bookshelfservice "novel-reader/backend/internal/domain/bookshelf/service"
	"novel-reader/backend/internal/domain/shared"
)

type ListEntries struct {
	GroupID  *int64
	Page     int
	PageSize int
}

type ListEntriesHandler struct {
	repo    bookshelfrepository.Repository
	service *bookshelfservice.Service
}

func NewListEntriesHandler(repo bookshelfrepository.Repository, service *bookshelfservice.Service) *ListEntriesHandler {
	return &ListEntriesHandler{repo: repo, service: service}
}

func (h *ListEntriesHandler) Handle(ctx context.Context, actor shared.Actor, input ListEntries) ([]bookshelfentity.Entry, int, error) {
	return h.repo.ListEntries(ctx, actor.ActorID, input.GroupID, input.Page, h.service.NormalizePageSize(input.PageSize))
}

type ListGroupsHandler struct {
	repo bookshelfrepository.Repository
}

func NewListGroupsHandler(repo bookshelfrepository.Repository) *ListGroupsHandler {
	return &ListGroupsHandler{repo: repo}
}

func (h *ListGroupsHandler) Handle(ctx context.Context, actor shared.Actor) ([]bookshelfentity.Group, error) {
	if _, err := h.repo.EnsureDefaultGroup(ctx, actor.ActorID); err != nil {
		return nil, err
	}
	return h.repo.ListGroups(ctx, actor.ActorID)
}

type FindEntryHandler struct {
	repo bookshelfrepository.Repository
}

func NewFindEntryHandler(repo bookshelfrepository.Repository) *FindEntryHandler {
	return &FindEntryHandler{repo: repo}
}

func (h *FindEntryHandler) Handle(ctx context.Context, actor shared.Actor, bookID int64) (bookshelfentity.Entry, error) {
	return h.repo.FindEntryByBookID(ctx, actor.ActorID, bookID)
}
