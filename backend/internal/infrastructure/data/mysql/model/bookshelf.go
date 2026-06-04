package model

import "time"

type BookshelfGroup struct {
	ID        int64      `gorm:"column:id;primaryKey"`
	UserID    int64      `gorm:"column:user_id"`
	Name      string     `gorm:"column:name"`
	SortOrder int        `gorm:"column:sort_order"`
	IsPinned  bool       `gorm:"column:is_pinned"`
	PinnedAt  *time.Time `gorm:"column:pinned_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (BookshelfGroup) TableName() string {
	return "bookshelf_groups"
}

type BookshelfItem struct {
	ID        int64      `gorm:"column:id;primaryKey"`
	UserID    int64      `gorm:"column:user_id"`
	BookID    int64      `gorm:"column:book_id"`
	GroupID   *int64     `gorm:"column:group_id"`
	IsPinned  bool       `gorm:"column:is_pinned"`
	PinnedAt  *time.Time `gorm:"column:pinned_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (BookshelfItem) TableName() string {
	return "bookshelf_items"
}
