package repository

import (
	"context"

	bookshelfentity "novel-reader/backend/internal/domain/bookshelf/entity"
)

type Repository interface {
	EnsureDefaultGroup(ctx context.Context, userID int64) (bookshelfentity.Group, error)
	ListGroups(ctx context.Context, userID int64) ([]bookshelfentity.Group, error)
	CreateGroup(ctx context.Context, userID int64, name string) (bookshelfentity.Group, error)
	RenameGroup(ctx context.Context, userID, groupID int64, name string) (bookshelfentity.Group, error)
	ReorderGroup(ctx context.Context, userID, groupID int64, sortOrder int) (bookshelfentity.Group, error)
	DeleteGroup(ctx context.Context, userID, groupID int64) error
	ListEntries(ctx context.Context, userID int64, groupID *int64, page, pageSize int) ([]bookshelfentity.Entry, int, error)
	FindEntryByBookID(ctx context.Context, userID, bookID int64) (bookshelfentity.Entry, error)
	AddBook(ctx context.Context, userID, bookID int64, groupID *int64) (bookshelfentity.Entry, error)
	UpdateBook(ctx context.Context, userID, bookID int64, groupID *int64, pinned *bool) (bookshelfentity.Entry, error)
	RemoveBook(ctx context.Context, userID, bookID int64) error
	BatchManage(ctx context.Context, userID int64, bookIDs []int64, action string, groupID *int64) error
}
