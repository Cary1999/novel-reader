package repo

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	bookentity "novel-reader/backend/internal/domain/book/entity"
	"novel-reader/backend/internal/domain/shared"
	"novel-reader/backend/internal/infrastructure/data/mysql/model"
)

type BookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) BookRepository {
	return BookRepository{db: db}
}

func (r BookRepository) SearchBooks(ctx context.Context, q, categoryName string, page, pageSize int) ([]bookentity.Book, int, error) {
	query := r.bookListQuery(ctx)
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("(b.title LIKE ? OR "+authorExprSQL()+" LIKE ? OR b.author LIKE ? OR b.description LIKE ?)", like, like, like, like)
	}
	if categoryName != "" {
		query = query.Where("c.name = ?", categoryName)
	}

	total, err := countBooksQuery(query)
	if err != nil {
		return nil, 0, err
	}

	var records []model.Book
	if err := query.
		Select(bookSelectColumns()).
		Order("b.created_at DESC, b.id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&records).Error; err != nil {
		return nil, 0, err
	}
	return toBookEntities(records), total, nil
}

func (r BookRepository) ListRecommendedBooks(ctx context.Context, page, pageSize int) ([]bookentity.Book, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	query := r.bookListQuery(ctx)
	total, err := countBooksQuery(query)
	if err != nil {
		return nil, 0, err
	}

	var records []model.Book
	if err := query.
		Select(bookSelectColumns()).
		Order("b.recommend_score DESC, b.created_at DESC, b.id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&records).Error; err != nil {
		return nil, 0, err
	}
	return toBookEntities(records), total, nil
}

func (r BookRepository) ListBooksByOwner(ctx context.Context, ownerID int64, q string, categoryID int64, page, pageSize int) ([]bookentity.Book, int, error) {
	query := r.bookListQuery(ctx).Where("b.owner_user_id = ?", ownerID)
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("(b.title LIKE ? OR b.description LIKE ?)", like, like)
	}
	if categoryID > 0 {
		query = query.Where("b.category_id = ?", categoryID)
	}

	total, err := countBooksQuery(query)
	if err != nil {
		return nil, 0, err
	}

	var records []model.Book
	if err := query.
		Select(bookSelectColumns()).
		Order("b.created_at DESC, b.id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&records).Error; err != nil {
		return nil, 0, err
	}
	return toBookEntities(records), total, nil
}

func (r BookRepository) FindBook(ctx context.Context, id int64) (bookentity.Book, error) {
	var record model.Book
	err := r.bookListQuery(ctx).
		Select(bookSelectColumns()).
		Where("b.id = ?", id).
		Take(&record).Error
	if err != nil {
		return bookentity.Book{}, mapGormNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r BookRepository) UpdateBookCoverPath(ctx context.Context, bookID int64, coverPath *string) error {
	result := r.db.WithContext(ctx).Model(&model.Book{}).Where("id = ?", bookID).Update("cover_path", coverPath)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r BookRepository) ListChapters(ctx context.Context, bookID int64) ([]bookentity.Chapter, error) {
	var records []model.Chapter
	if err := r.db.WithContext(ctx).
		Model(&model.Chapter{}).
		Select("id", "book_id", "chapter_index", "title").
		Where("book_id = ?", bookID).
		Order("chapter_index ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]bookentity.Chapter, 0, len(records))
	for _, record := range records {
		items = append(items, record.ToEntity())
	}
	return items, nil
}

func (r BookRepository) FindChapter(ctx context.Context, bookID, chapterID int64) (bookentity.Chapter, error) {
	var record model.Chapter
	if err := r.db.WithContext(ctx).
		Model(&model.Chapter{}).
		Where("book_id = ? AND id = ?", bookID, chapterID).
		Take(&record).Error; err != nil {
		return bookentity.Chapter{}, mapGormNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r BookRepository) CreateBookWithChapters(ctx context.Context, item bookentity.Book, categoryID *int64, uploadID int64, chapters []bookentity.ChapterDraft) (int64, error) {
	var createdID int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		bookModel := model.Book{}
		bookModel.FromEntity(&item)
		bookModel.Author = legacyAuthor(ownerIDValue(item))
		if categoryID != nil {
			bookModel.CategoryID = sql.NullInt64{Int64: *categoryID, Valid: true}
		}
		if uploadID > 0 {
			bookModel.SourceUploadID = sql.NullInt64{Int64: uploadID, Valid: true}
		}
		bookModel.ChapterCount = len(chapters)
		if len(chapters) > 0 {
			bookModel.LatestChapterTitle = chapters[len(chapters)-1].Title
		}

		if err := tx.WithContext(ctx).Create(&bookModel).Error; err != nil {
			return err
		}
		createdID = bookModel.ID

		chapterModels := make([]model.Chapter, 0, len(chapters))
		for i, chapterItem := range chapters {
			index := chapterItem.Index
			if index == 0 {
				index = i + 1
			}
			chapterModels = append(chapterModels, model.Chapter{
				BookID:  createdID,
				Index:   index,
				Title:   chapterItem.Title,
				Content: chapterItem.Content,
			})
		}
		if len(chapterModels) > 0 {
			if err := tx.Create(&chapterModels).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return createdID, nil
}

func (r BookRepository) CreateBook(ctx context.Context, item bookentity.Book, categoryID *int64) (bookentity.Book, error) {
	bookModel := model.Book{}
	bookModel.FromEntity(&item)
	bookModel.Author = legacyAuthor(ownerIDValue(item))
	if categoryID != nil {
		bookModel.CategoryID = sql.NullInt64{Int64: *categoryID, Valid: true}
	}
	bookModel.ChapterCount = 0
	bookModel.LatestChapterTitle = ""

	if err := r.db.WithContext(ctx).Create(&bookModel).Error; err != nil {
		return bookentity.Book{}, err
	}

	return r.FindBook(ctx, bookModel.ID)
}

func (r BookRepository) UpdateBookMetadata(ctx context.Context, bookID int64, item bookentity.Book, categoryID *int64) (bookentity.Book, error) {
	updates := map[string]any{
		"title":       item.Title,
		"category_id": categoryID,
		"description": item.Description,
	}
	if item.RecommendScore != 0 {
		updates["recommend_score"] = item.RecommendScore
	}

	result := r.db.WithContext(ctx).Model(&model.Book{}).Where("id = ?", bookID).Updates(updates)
	if result.Error != nil {
		return bookentity.Book{}, result.Error
	}
	if result.RowsAffected == 0 {
		return bookentity.Book{}, shared.ErrNotFound
	}
	return r.FindBook(ctx, bookID)
}

func (r BookRepository) UpdateRecommendScore(ctx context.Context, bookID int64, recommendScore int) (bookentity.Book, error) {
	result := r.db.WithContext(ctx).Model(&model.Book{}).Where("id = ?", bookID).Update("recommend_score", recommendScore)
	if result.Error != nil {
		return bookentity.Book{}, result.Error
	}
	if result.RowsAffected == 0 {
		return bookentity.Book{}, shared.ErrNotFound
	}
	return r.FindBook(ctx, bookID)
}

func (r BookRepository) DeleteBook(ctx context.Context, bookID int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", bookID).Delete(&model.Book{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r BookRepository) AddChapter(ctx context.Context, bookID int64, item bookentity.Chapter) (bookentity.Chapter, error) {
	var chapterID int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureBookExistsGorm(tx, bookID); err != nil {
			return err
		}
		nextIndex, err := nextChapterIndex(tx, bookID)
		if err != nil {
			return err
		}

		record := model.Chapter{
			BookID:  bookID,
			Index:   nextIndex,
			Title:   item.Title,
			Content: item.Content,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		chapterID = record.ID
		return refreshBookChapterStatsGorm(tx, bookID)
	})
	if err != nil {
		return bookentity.Chapter{}, err
	}
	return r.FindChapter(ctx, bookID, chapterID)
}

func (r BookRepository) UpdateChapter(ctx context.Context, bookID, chapterID int64, item bookentity.Chapter) (bookentity.Chapter, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureChapterExistsGorm(tx, bookID, chapterID); err != nil {
			return err
		}
		if err := tx.Model(&model.Chapter{}).
			Where("book_id = ? AND id = ?", bookID, chapterID).
			Updates(map[string]any{"title": item.Title, "content": item.Content}).Error; err != nil {
			return err
		}
		return refreshBookChapterStatsGorm(tx, bookID)
	})
	if err != nil {
		return bookentity.Chapter{}, err
	}
	return r.FindChapter(ctx, bookID, chapterID)
}

func (r BookRepository) DeleteChapter(ctx context.Context, bookID, chapterID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record model.Chapter
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("book_id = ? AND id = ?", bookID, chapterID).
			Take(&record).Error; err != nil {
			return mapGormNotFound(err)
		}
		if err := tx.Where("book_id = ? AND id = ?", bookID, chapterID).Delete(&model.Chapter{}).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			UPDATE chapters SET chapter_index = chapter_index - 1
			WHERE book_id = ? AND chapter_index > ?
		`, bookID, record.Index).Error; err != nil {
			return err
		}
		return refreshBookChapterStatsGorm(tx, bookID)
	})
}

func (r BookRepository) bookListQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("books AS b").
		Joins("LEFT JOIN users u ON u.id = b.owner_user_id").
		Joins("LEFT JOIN categories c ON c.id = b.category_id")
}

func bookSelectColumns() string {
	return "b.id, b.title, " + authorExprSQL() + " AS author, b.owner_user_id, b.category_id, COALESCE(c.name, '') AS category_name, b.description, b.chapter_count, b.latest_chapter_title, b.recommend_score, b.cover_path, b.created_at, b.updated_at"
}

func countBooksQuery(query *gorm.DB) (int, error) {
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return int(total), nil
}

func toBookEntities(records []model.Book) []bookentity.Book {
	items := make([]bookentity.Book, 0, len(records))
	for _, record := range records {
		items = append(items, record.ToEntity())
	}
	return items
}

func mapGormNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.ErrNotFound
	}
	return err
}

func ensureBookExistsGorm(tx *gorm.DB, bookID int64) error {
	var record model.Book
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("id = ?", bookID).Take(&record).Error; err != nil {
		return mapGormNotFound(err)
	}
	return nil
}

func ensureChapterExistsGorm(tx *gorm.DB, bookID, chapterID int64) error {
	var record model.Chapter
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("book_id = ? AND id = ?", bookID, chapterID).Take(&record).Error; err != nil {
		return mapGormNotFound(err)
	}
	return nil
}

func nextChapterIndex(tx *gorm.DB, bookID int64) (int, error) {
	var record model.Chapter
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("chapter_index").
		Where("book_id = ?", bookID).
		Order("chapter_index DESC").
		Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 1, nil
		}
		return 0, err
	}
	return record.Index + 1, nil
}

func refreshBookChapterStatsGorm(tx *gorm.DB, bookID int64) error {
	var total int64
	if err := tx.Model(&model.Chapter{}).Where("book_id = ?", bookID).Count(&total).Error; err != nil {
		return err
	}
	var latest model.Chapter
	err := tx.Select("title").Where("book_id = ?", bookID).Order("chapter_index DESC").Take(&latest).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	latestTitle := ""
	if err == nil {
		latestTitle = latest.Title
	}
	return tx.Model(&model.Book{}).Where("id = ?", bookID).Updates(map[string]any{
		"chapter_count":        total,
		"latest_chapter_title": latestTitle,
	}).Error
}

func ownerIDValue(item bookentity.Book) int64 {
	if item.OwnerUserID == nil {
		return 0
	}
	return *item.OwnerUserID
}
