package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	bookshelfentity "novel-reader/backend/internal/domain/bookshelf/entity"
	bookshelfrepository "novel-reader/backend/internal/domain/bookshelf/repository"
	"novel-reader/backend/internal/domain/shared"
	"novel-reader/backend/internal/infrastructure/data/mysql/model"
)

const defaultBookshelfGroupName = "默认书架"

type BookshelfRepository struct {
	db *gorm.DB
}

var _ bookshelfrepository.Repository = BookshelfRepository{}

func NewBookshelfRepository(db *gorm.DB) BookshelfRepository {
	return BookshelfRepository{db: db}
}

func (r BookshelfRepository) EnsureDefaultGroup(ctx context.Context, userID int64) (bookshelfentity.Group, error) {
	var record model.BookshelfGroup
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_default = 1", userID).
		Order("id ASC").
		Take(&record).Error
	if err == nil {
		return toBookshelfGroup(record), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return bookshelfentity.Group{}, mapGormNotFound(err)
	}
	record = model.BookshelfGroup{
		UserID:    userID,
		Name:      defaultBookshelfGroupName,
		SortOrder: 0,
		IsDefault: true,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return bookshelfentity.Group{}, err
	}
	return toBookshelfGroup(record), nil
}

func (r BookshelfRepository) ListGroups(ctx context.Context, userID int64) ([]bookshelfentity.Group, error) {
	type groupRow struct {
		model.BookshelfGroup
		ItemCount int `gorm:"column:item_count"`
	}
	var rows []groupRow
	if err := r.db.WithContext(ctx).
		Table("bookshelf_groups g").
		Select("g.id, g.user_id, g.name, g.sort_order, g.is_default, g.created_at, g.updated_at, COUNT(i.id) AS item_count").
		Joins("LEFT JOIN bookshelf_items i ON i.group_id = g.id").
		Where("g.user_id = ?", userID).
		Group("g.id, g.user_id, g.name, g.sort_order, g.is_default, g.created_at, g.updated_at").
		Order("g.sort_order ASC, g.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]bookshelfentity.Group, 0, len(rows))
	for _, row := range rows {
		item := toBookshelfGroup(row.BookshelfGroup)
		item.ItemCount = row.ItemCount
		items = append(items, item)
	}
	return items, nil
}

func (r BookshelfRepository) CreateGroup(ctx context.Context, userID int64, name string) (bookshelfentity.Group, error) {
	var created model.BookshelfGroup
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxSort sql.NullInt64
		if err := tx.Table("bookshelf_groups").Select("COALESCE(MAX(sort_order), 0)").Where("user_id = ?", userID).Scan(&maxSort).Error; err != nil {
			return err
		}
		created = model.BookshelfGroup{
			UserID:    userID,
			Name:      name,
			SortOrder: int(maxSort.Int64) + 1,
			IsDefault: false,
		}
		return tx.Create(&created).Error
	}); err != nil {
		return bookshelfentity.Group{}, err
	}
	return toBookshelfGroup(created), nil
}

func (r BookshelfRepository) RenameGroup(ctx context.Context, userID, groupID int64, name string) (bookshelfentity.Group, error) {
	if err := r.db.WithContext(ctx).Model(&model.BookshelfGroup{}).
		Where("id = ? AND user_id = ?", groupID, userID).
		Update("name", name).Error; err != nil {
		return bookshelfentity.Group{}, err
	}
	return r.findGroup(ctx, userID, groupID)
}

func (r BookshelfRepository) ReorderGroup(ctx context.Context, userID, groupID int64, sortOrder int) (bookshelfentity.Group, error) {
	if err := r.db.WithContext(ctx).Model(&model.BookshelfGroup{}).
		Where("id = ? AND user_id = ?", groupID, userID).
		Update("sort_order", sortOrder).Error; err != nil {
		return bookshelfentity.Group{}, err
	}
	return r.findGroup(ctx, userID, groupID)
}

func (r BookshelfRepository) DeleteGroup(ctx context.Context, userID, groupID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var group model.BookshelfGroup
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", groupID, userID).
			Take(&group).Error; err != nil {
			return mapGormNotFound(err)
		}
		if group.IsDefault {
			return shared.NewError(httpStatusBadRequest(), "BAD_REQUEST", "default group cannot be deleted")
		}

		defaultGroup, err := ensureDefaultGroupTx(ctx, tx, userID)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.BookshelfItem{}).
			Where("user_id = ? AND group_id = ?", userID, groupID).
			Update("group_id", defaultGroup.ID).Error; err != nil {
			return err
		}
		return tx.Delete(&model.BookshelfGroup{}, groupID).Error
	})
}

func (r BookshelfRepository) ListEntries(ctx context.Context, userID int64, groupID *int64, page, pageSize int) ([]bookshelfentity.Entry, int, error) {
	page, pageSize = normalizePaging(page, pageSize)
	query := r.entryQuery(ctx, userID, groupID)
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []bookshelfEntryRecord
	if err := query.
		Select(entrySelectColumns()).
		Order(entryOrderSQL()).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]bookshelfentity.Entry, 0, len(rows))
	for _, row := range rows {
		items = append(items, toBookshelfEntry(row))
	}
	return items, int(total), nil
}

func (r BookshelfRepository) FindEntryByBookID(ctx context.Context, userID, bookID int64) (bookshelfentity.Entry, error) {
	var row bookshelfEntryRecord
	err := r.entryQuery(ctx, userID, nil).
		Where("bi.book_id = ?", bookID).
		Select(entrySelectColumns()).
		Take(&row).Error
	if err != nil {
		return bookshelfentity.Entry{}, mapGormNotFound(err)
	}
	return toBookshelfEntry(row), nil
}

func (r BookshelfRepository) AddBook(ctx context.Context, userID, bookID int64, groupID *int64) (bookshelfentity.Entry, error) {
	var result bookshelfentity.Entry
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		targetGroup, err := r.ensureGroupTx(ctx, tx, userID, groupID)
		if err != nil {
			return err
		}
		if err := ensureBookExistsGorm(tx, bookID); err != nil {
			return err
		}
		var item model.BookshelfItem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND book_id = ?", userID, bookID).
			Take(&item).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return mapGormNotFound(err)
			}
			item = model.BookshelfItem{
				UserID:    userID,
				BookID:    bookID,
				GroupID:   targetGroup.ID,
				IsPinned:  false,
				PinnedAt:  nil,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		} else {
			item.GroupID = targetGroup.ID
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
		}
		entry, err := r.findEntryTx(ctx, tx, userID, bookID)
		if err != nil {
			return err
		}
		result = entry
		return nil
	})
	return result, err
}

func (r BookshelfRepository) UpdateBook(ctx context.Context, userID, bookID int64, groupID *int64, pinned *bool) (bookshelfentity.Entry, error) {
	var result bookshelfentity.Entry
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item model.BookshelfItem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND book_id = ?", userID, bookID).
			Take(&item).Error; err != nil {
			return mapGormNotFound(err)
		}
		if groupID != nil {
			targetGroup, err := r.ensureGroupTx(ctx, tx, userID, groupID)
			if err != nil {
				return err
			}
			item.GroupID = targetGroup.ID
		}
		if pinned != nil {
			item.IsPinned = *pinned
			if *pinned {
				now := time.Now().UTC()
				item.PinnedAt = &now
			} else {
				item.PinnedAt = nil
			}
		}
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		entry, err := r.findEntryTx(ctx, tx, userID, bookID)
		if err != nil {
			return err
		}
		result = entry
		return nil
	})
	return result, err
}

func (r BookshelfRepository) RemoveBook(ctx context.Context, userID, bookID int64) error {
	result := r.db.WithContext(ctx).Where("user_id = ? AND book_id = ?", userID, bookID).Delete(&model.BookshelfItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r BookshelfRepository) BatchManage(ctx context.Context, userID int64, bookIDs []int64, action string, groupID *int64) error {
	if len(bookIDs) == 0 {
		return shared.NewError(httpStatusBadRequest(), "BAD_REQUEST", "book ids are required")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var targetGroupID int64
		if action == "move" {
			targetGroup, err := r.ensureGroupTx(ctx, tx, userID, groupID)
			if err != nil {
				return err
			}
			targetGroupID = targetGroup.ID
		}

		for _, bookID := range bookIDs {
			switch action {
			case "move":
				if err := tx.Model(&model.BookshelfItem{}).
					Where("user_id = ? AND book_id = ?", userID, bookID).
					Updates(map[string]any{"group_id": targetGroupID}).Error; err != nil {
					return mapGormNotFound(err)
				}
			case "pin":
				now := time.Now().UTC()
				if err := tx.Model(&model.BookshelfItem{}).
					Where("user_id = ? AND book_id = ?", userID, bookID).
					Updates(map[string]any{"is_pinned": true, "pinned_at": now}).Error; err != nil {
					return mapGormNotFound(err)
				}
			case "unpin":
				if err := tx.Model(&model.BookshelfItem{}).
					Where("user_id = ? AND book_id = ?", userID, bookID).
					Updates(map[string]any{"is_pinned": false, "pinned_at": nil}).Error; err != nil {
					return mapGormNotFound(err)
				}
			case "remove":
				if err := tx.Where("user_id = ? AND book_id = ?", userID, bookID).Delete(&model.BookshelfItem{}).Error; err != nil {
					return err
				}
			default:
				return shared.NewError(httpStatusBadRequest(), "BAD_REQUEST", "invalid bookshelf action")
			}
		}
		return nil
	})
}

func (r BookshelfRepository) findGroup(ctx context.Context, userID, groupID int64) (bookshelfentity.Group, error) {
	var record model.BookshelfGroup
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", groupID, userID).Take(&record).Error; err != nil {
		return bookshelfentity.Group{}, mapGormNotFound(err)
	}
	return toBookshelfGroup(record), nil
}

func (r BookshelfRepository) ensureGroupTx(ctx context.Context, tx *gorm.DB, userID int64, groupID *int64) (model.BookshelfGroup, error) {
	if groupID == nil || *groupID <= 0 {
		return ensureDefaultGroupTx(ctx, tx, userID)
	}
	var record model.BookshelfGroup
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", *groupID, userID).
		Take(&record).Error; err != nil {
		return model.BookshelfGroup{}, mapGormNotFound(err)
	}
	return record, nil
}

func ensureDefaultGroupTx(ctx context.Context, tx *gorm.DB, userID int64) (model.BookshelfGroup, error) {
	var record model.BookshelfGroup
	if err := tx.Where("user_id = ? AND is_default = 1", userID).Order("id ASC").Take(&record).Error; err == nil {
		return record, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.BookshelfGroup{}, mapGormNotFound(err)
	}
	record = model.BookshelfGroup{
		UserID:    userID,
		Name:      defaultBookshelfGroupName,
		SortOrder: 0,
		IsDefault: true,
	}
	if err := tx.Create(&record).Error; err != nil {
		return model.BookshelfGroup{}, err
	}
	return record, nil
}

func (r BookshelfRepository) entryQuery(ctx context.Context, userID int64, groupID *int64) *gorm.DB {
	query := r.db.WithContext(ctx).
		Table("bookshelf_items bi").
		Joins("JOIN bookshelf_groups g ON g.id = bi.group_id").
		Joins("LEFT JOIN users u ON u.id = bi.user_id").
		Joins("LEFT JOIN books b ON b.id = bi.book_id").
		Joins("LEFT JOIN categories c ON c.id = b.category_id").
		Where("bi.user_id = ?", userID)
	if groupID != nil && *groupID > 0 {
		query = query.Where("bi.group_id = ?", *groupID)
	}
	return query
}

type bookshelfEntryRecord struct {
	ItemID    int64          `gorm:"column:item_id"`
	UserID    int64          `gorm:"column:user_id"`
	BookID    int64          `gorm:"column:book_id"`
	GroupID   int64          `gorm:"column:group_id"`
	GroupName string         `gorm:"column:group_name"`
	IsDefault bool           `gorm:"column:is_default"`
	IsPinned  bool           `gorm:"column:is_pinned"`
	PinnedAt  *time.Time     `gorm:"column:pinned_at"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	BookCreatedAt time.Time  `gorm:"column:book_created_at"`
	BookUpdatedAt time.Time  `gorm:"column:book_updated_at"`
	BookModel model.Book     `gorm:"embedded"`
}

func entrySelectColumns() string {
	return `bi.id AS item_id, bi.user_id, bi.book_id, bi.group_id, g.name AS group_name, g.is_default,
		bi.is_pinned, bi.pinned_at, bi.created_at, bi.updated_at,
		b.id AS id, b.title AS title, ` + authorExprSQL() + ` AS author, b.owner_user_id, b.category_id,
		COALESCE(c.name, '') AS category_name, b.description, b.chapter_count, b.latest_chapter_title,
		b.recommend_score, b.cover_path, b.created_at AS book_created_at, b.updated_at AS book_updated_at`
}

func entryOrderSQL() string {
	return "g.sort_order ASC, g.id ASC, bi.is_pinned DESC, COALESCE(bi.pinned_at, bi.created_at) DESC, bi.created_at DESC, bi.id DESC"
}

func toBookshelfGroup(record model.BookshelfGroup) bookshelfentity.Group {
	return bookshelfentity.Group{
		ID:        record.ID,
		UserID:    record.UserID,
		Name:      record.Name,
		SortOrder: record.SortOrder,
		IsDefault: record.IsDefault,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func toBookshelfEntry(record bookshelfEntryRecord) bookshelfentity.Entry {
	item := bookshelfentity.Entry{
		ID:        record.ItemID,
		UserID:    record.UserID,
		BookID:    record.BookID,
		GroupID:   record.GroupID,
		GroupName: record.GroupName,
		IsDefault: record.IsDefault,
		IsPinned:  record.IsPinned,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
	bookItem := record.BookModel.ToEntity()
	bookItem.ID = record.BookID
	bookItem.CreatedAt = record.BookCreatedAt
	bookItem.UpdatedAt = record.BookUpdatedAt
	item.Book = bookItem
	if record.PinnedAt != nil {
		item.PinnedAt = record.PinnedAt
	}
	return item
}

func (r BookshelfRepository) findEntryTx(ctx context.Context, tx *gorm.DB, userID, bookID int64) (bookshelfentity.Entry, error) {
	var row bookshelfEntryRecord
	if err := tx.WithContext(ctx).
		Table("bookshelf_items bi").
		Joins("JOIN bookshelf_groups g ON g.id = bi.group_id").
		Joins("LEFT JOIN users u ON u.id = bi.user_id").
		Joins("LEFT JOIN books b ON b.id = bi.book_id").
		Joins("LEFT JOIN categories c ON c.id = b.category_id").
		Where("bi.user_id = ? AND bi.book_id = ?", userID, bookID).
		Select(entrySelectColumns()).
		Take(&row).Error; err != nil {
		return bookshelfentity.Entry{}, mapGormNotFound(err)
	}
	return toBookshelfEntry(row), nil
}

func normalizePaging(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func httpStatusBadRequest() int {
	return 400
}
