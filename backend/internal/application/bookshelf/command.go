package bookshelfapp

import (
	"context"
	"net/http"

	bookshelfentity "novel-reader/backend/internal/domain/bookshelf/entity"
	bookshelfrepository "novel-reader/backend/internal/domain/bookshelf/repository"
	bookshelfservice "novel-reader/backend/internal/domain/bookshelf/service"
	"novel-reader/backend/internal/domain/shared"
)

type CreateGroup struct {
	Name string `json:"name"`
}

type RenameGroup struct {
	Name string `json:"name"`
}

type ReorderGroup struct {
	SortOrder int `json:"sortOrder"`
}

type AddBook struct {
	GroupID *int64 `json:"groupId,omitempty"`
}

type UpdateBook struct {
	GroupID *int64 `json:"groupId,omitempty"`
	Pinned  *bool  `json:"pinned,omitempty"`
}

type BatchManage struct {
	Action  string  `json:"action"`
	GroupID *int64  `json:"groupId,omitempty"`
	BookIDs []int64 `json:"bookIds"`
}

type CreateGroupHandler struct {
	repo    bookshelfrepository.Repository
	service *bookshelfservice.Service
}

func NewCreateGroupHandler(repo bookshelfrepository.Repository, service *bookshelfservice.Service) *CreateGroupHandler {
	return &CreateGroupHandler{repo: repo, service: service}
}

func (h *CreateGroupHandler) Handle(ctx context.Context, actor shared.Actor, input CreateGroup) (bookshelfentity.Group, error) {
	name, err := h.service.NormalizeGroupName(input.Name)
	if err != nil {
		return bookshelfentity.Group{}, err
	}
	return h.repo.CreateGroup(ctx, actor.ActorID, name)
}

type RenameGroupHandler struct {
	repo    bookshelfrepository.Repository
	service *bookshelfservice.Service
}

func NewRenameGroupHandler(repo bookshelfrepository.Repository, service *bookshelfservice.Service) *RenameGroupHandler {
	return &RenameGroupHandler{repo: repo, service: service}
}

func (h *RenameGroupHandler) Handle(ctx context.Context, actor shared.Actor, groupID int64, input RenameGroup) (bookshelfentity.Group, error) {
	name, err := h.service.NormalizeGroupName(input.Name)
	if err != nil {
		return bookshelfentity.Group{}, err
	}
	return h.repo.RenameGroup(ctx, actor.ActorID, groupID, name)
}

type ReorderGroupHandler struct {
	repo bookshelfrepository.Repository
}

func NewReorderGroupHandler(repo bookshelfrepository.Repository, service *bookshelfservice.Service) *ReorderGroupHandler {
	_ = service
	return &ReorderGroupHandler{repo: repo}
}

func (h *ReorderGroupHandler) Handle(ctx context.Context, actor shared.Actor, groupID int64, input ReorderGroup) (bookshelfentity.Group, error) {
	return h.repo.ReorderGroup(ctx, actor.ActorID, groupID, input.SortOrder)
}

type DeleteGroupHandler struct {
	repo bookshelfrepository.Repository
}

func NewDeleteGroupHandler(repo bookshelfrepository.Repository) *DeleteGroupHandler {
	return &DeleteGroupHandler{repo: repo}
}

func (h *DeleteGroupHandler) Handle(ctx context.Context, actor shared.Actor, groupID int64) error {
	return h.repo.DeleteGroup(ctx, actor.ActorID, groupID)
}

type AddBookHandler struct {
	repo bookshelfrepository.Repository
}

func NewAddBookHandler(repo bookshelfrepository.Repository) *AddBookHandler {
	return &AddBookHandler{repo: repo}
}

func (h *AddBookHandler) Handle(ctx context.Context, actor shared.Actor, bookID int64, input AddBook) (bookshelfentity.Entry, error) {
	return h.repo.AddBook(ctx, actor.ActorID, bookID, input.GroupID)
}

type UpdateBookHandler struct {
	repo bookshelfrepository.Repository
}

func NewUpdateBookHandler(repo bookshelfrepository.Repository) *UpdateBookHandler {
	return &UpdateBookHandler{repo: repo}
}

func (h *UpdateBookHandler) Handle(ctx context.Context, actor shared.Actor, bookID int64, input UpdateBook) (bookshelfentity.Entry, error) {
	return h.repo.UpdateBook(ctx, actor.ActorID, bookID, input.GroupID, input.Pinned)
}

type RemoveBookHandler struct {
	repo bookshelfrepository.Repository
}

func NewRemoveBookHandler(repo bookshelfrepository.Repository) *RemoveBookHandler {
	return &RemoveBookHandler{repo: repo}
}

func (h *RemoveBookHandler) Handle(ctx context.Context, actor shared.Actor, bookID int64) error {
	return h.repo.RemoveBook(ctx, actor.ActorID, bookID)
}

type BatchManageHandler struct {
	repo    bookshelfrepository.Repository
	service *bookshelfservice.Service
}

func NewBatchManageHandler(repo bookshelfrepository.Repository, service *bookshelfservice.Service) *BatchManageHandler {
	return &BatchManageHandler{repo: repo, service: service}
}

func (h *BatchManageHandler) Handle(ctx context.Context, actor shared.Actor, input BatchManage) error {
	action, err := h.service.NormalizeAction(input.Action)
	if err != nil {
		return err
	}
	if len(input.BookIDs) == 0 {
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "book ids are required")
	}
	return h.repo.BatchManage(ctx, actor.ActorID, input.BookIDs, action, input.GroupID)
}
