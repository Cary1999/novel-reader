package entity

import (
	"time"

	bookentity "novel-reader/backend/internal/domain/book/entity"
)

type Group struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"userId"`
	Name      string     `json:"name"`
	SortOrder int        `json:"sortOrder"`
	ItemCount int        `json:"itemCount"`
	IsPinned  bool       `json:"isPinned"`
	PinnedAt  *time.Time `json:"pinnedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type Entry struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"userId"`
	BookID    int64           `json:"bookId"`
	GroupID   *int64          `json:"groupId,omitempty"`
	GroupName string          `json:"groupName,omitempty"`
	IsPinned  bool            `json:"isPinned"`
	PinnedAt  *time.Time      `json:"pinnedAt,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Book      bookentity.Book `json:"book"`
}
