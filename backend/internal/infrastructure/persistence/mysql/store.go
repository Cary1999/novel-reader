package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"novel-reader/backend/internal/domain/book"
	"novel-reader/backend/internal/domain/category"
	"novel-reader/backend/internal/domain/identity"
	"novel-reader/backend/internal/domain/shared"
	"novel-reader/backend/internal/domain/upload"
)

type SeedOptions struct {
	AdminUsername string
	AdminPassword string
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Migrate(ctx context.Context) error {
	for _, stmt := range splitSQL(schemaSQL) {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec schema statement: %w", err)
		}
	}
	if err := s.ensureColumn(ctx, "users", "nickname", `ALTER TABLE users ADD COLUMN nickname VARCHAR(64) NOT NULL DEFAULT '' AFTER username`); err != nil {
		return err
	}
	if err := s.ensureColumn(ctx, "books", "owner_user_id", `ALTER TABLE books ADD COLUMN owner_user_id BIGINT NULL AFTER author`); err != nil {
		return err
	}
	if err := s.ensureIndex(ctx, "books", "idx_books_owner_user_id", `ALTER TABLE books ADD INDEX idx_books_owner_user_id (owner_user_id)`); err != nil {
		return err
	}
	if err := s.ensureForeignKey(ctx, "books", "fk_books_owner_user", `ALTER TABLE books ADD CONSTRAINT fk_books_owner_user FOREIGN KEY (owner_user_id) REFERENCES users(id)`); err != nil {
		return err
	}
	if err := s.ensureColumn(ctx, "books", "recommend_score", `ALTER TABLE books ADD COLUMN recommend_score INT NOT NULL DEFAULT 0 AFTER latest_chapter_title`); err != nil {
		return err
	}
	if err := s.ensureColumn(ctx, "books", "cover_path", `ALTER TABLE books ADD COLUMN cover_path VARCHAR(255) NULL AFTER recommend_score`); err != nil {
		return err
	}
	if err := s.ensureIndex(ctx, "books", "idx_books_recommend_score", `ALTER TABLE books ADD INDEX idx_books_recommend_score (recommend_score)`); err != nil {
		return err
	}
	return nil
}

func (s *Store) Seed(ctx context.Context, opts SeedOptions) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(opts.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO users (username, nickname, password_hash, role)
		VALUES (?, ?, ?, 'admin')
		ON DUPLICATE KEY UPDATE username = username
	`, opts.AdminUsername, opts.AdminUsername, string(hash)); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE users SET nickname = username WHERE nickname = ''`); err != nil {
		return err
	}
	var adminID int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = ?`, opts.AdminUsername).Scan(&adminID); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE books b
		LEFT JOIN users u ON u.username = b.author
		SET b.owner_user_id = COALESCE(u.id, ?)
		WHERE b.owner_user_id IS NULL
	`, adminID); err != nil {
		return err
	}

	categories := []string{"玄幻", "都市", "科幻", "历史", "游戏"}
	for _, name := range categories {
		if _, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO categories (name) VALUES (?)`, name); err != nil {
			return err
		}
	}

	var existing int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books WHERE title = ? AND owner_user_id = ?`, "星河书页", adminID).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	var categoryID int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM categories WHERE name = ?`, "科幻").Scan(&categoryID); err != nil {
		return err
	}
	chapters := []book.ChapterDraft{
		{Index: 1, Title: "第一章 本地书架", Content: "清晨的终端亮起，新的书架在本地服务中安静展开。"},
		{Index: 2, Title: "Chapter 2 The Reader", Content: "A reader signs in and opens the next page without touching any remote content."},
	}
	_, err = s.CreateBookWithChapters(ctx, book.CreateInput{
		Title:       "星河书页",
		OwnerUserID: adminID,
		Description: "用于本地 smoke check 的示例小说。",
	}, &categoryID, 0, chapters)
	return err
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string, role shared.Role) (identity.User, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO users (username, nickname, password_hash, role) VALUES (?, ?, ?, ?)
	`, username, username, passwordHash, role)
	if err != nil {
		return identity.User{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return identity.User{}, err
	}
	return s.FindUserByID(ctx, id)
}

func (s *Store) FindUserByUsername(ctx context.Context, username string) (identity.User, error) {
	return scanUser(s.db.QueryRowContext(ctx, `
		SELECT id, username, nickname, password_hash, role, created_at, updated_at
		FROM users WHERE username = ?
	`, username))
}

func (s *Store) FindUserByID(ctx context.Context, id int64) (identity.User, error) {
	return scanUser(s.db.QueryRowContext(ctx, `
		SELECT id, username, nickname, password_hash, role, created_at, updated_at
		FROM users WHERE id = ?
	`, id))
}

func (s *Store) UpdateUserNickname(ctx context.Context, id int64, nickname string) (identity.User, error) {
	if _, err := s.FindUserByID(ctx, id); err != nil {
		return identity.User{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE users SET nickname = ? WHERE id = ?`, nickname, id); err != nil {
		return identity.User{}, err
	}
	return s.FindUserByID(ctx, id)
}

func (s *Store) UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) ListCategories(ctx context.Context) ([]category.Category, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at FROM categories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []category.Category
	for rows.Next() {
		var item category.Category
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) FindCategoryByID(ctx context.Context, id int64) (category.Category, error) {
	var item category.Category
	err := s.db.QueryRowContext(ctx, `SELECT id, name, created_at FROM categories WHERE id = ?`, id).Scan(&item.ID, &item.Name, &item.CreatedAt)
	if err != nil {
		return category.Category{}, mapNotFound(err)
	}
	return item, nil
}

func (s *Store) CreateCategory(ctx context.Context, name string) (category.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return category.Category{}, shared.ErrNotFound
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO categories (name) VALUES (?)`, name)
	if err != nil {
		return category.Category{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return category.Category{}, err
	}
	return s.FindCategoryByID(ctx, id)
}

func (s *Store) UpdateCategory(ctx context.Context, id int64, name string) (category.Category, error) {
	if _, err := s.FindCategoryByID(ctx, id); err != nil {
		return category.Category{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE categories SET name = ? WHERE id = ?`, strings.TrimSpace(name), id); err != nil {
		return category.Category{}, err
	}
	return s.FindCategoryByID(ctx, id)
}

func (s *Store) DeleteCategory(ctx context.Context, id int64) error {
	var used int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books WHERE category_id = ?`, id).Scan(&used); err != nil {
		return err
	}
	if used > 0 {
		return shared.NewError(409, "CONFLICT", "category is used by books")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) SearchBooks(ctx context.Context, q, categoryName string, page, pageSize int) ([]book.Book, int, error) {
	args := []any{}
	where := "WHERE 1=1"
	if q != "" {
		like := "%" + q + "%"
		where += " AND (b.title LIKE ? OR " + authorExprSQL() + " LIKE ? OR b.author LIKE ? OR b.description LIKE ?)"
		args = append(args, like, like, like, like)
	}
	if categoryName != "" {
		where += " AND c.name = ?"
		args = append(args, categoryName)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, booksBaseCountSQL()+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitArgs := append([]any{}, args...)
	limitArgs = append(limitArgs, pageSize, (page-1)*pageSize)
	rows, err := s.db.QueryContext(ctx, `
		`+booksSelectSQL()+where+`
		ORDER BY b.created_at DESC, b.id DESC
		LIMIT ? OFFSET ?
	`, limitArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanBooks(rows)
	return items, total, err
}

func (s *Store) ListRecommendedBooks(ctx context.Context, page, pageSize int) ([]book.Book, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx, `
		`+booksSelectSQL()+`WHERE 1=1
		ORDER BY b.recommend_score DESC, b.created_at DESC, b.id DESC
		LIMIT ? OFFSET ?
	`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanBooks(rows)
	return items, total, err
}

func (s *Store) ListBooksByOwner(ctx context.Context, ownerID int64, q string, categoryID int64, page, pageSize int) ([]book.Book, int, error) {
	args := []any{ownerID}
	where := "WHERE b.owner_user_id = ?"
	if q != "" {
		like := "%" + q + "%"
		where += " AND (b.title LIKE ? OR b.description LIKE ?)"
		args = append(args, like, like)
	}
	if categoryID > 0 {
		where += " AND b.category_id = ?"
		args = append(args, categoryID)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, booksBaseCountSQL()+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limitArgs := append([]any{}, args...)
	limitArgs = append(limitArgs, pageSize, (page-1)*pageSize)
	rows, err := s.db.QueryContext(ctx, `
		`+booksSelectSQL()+where+`
		ORDER BY b.created_at DESC, b.id DESC
		LIMIT ? OFFSET ?
	`, limitArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanBooks(rows)
	return items, total, err
}

func (s *Store) FindBook(ctx context.Context, id int64) (book.Book, error) {
	row := s.db.QueryRowContext(ctx, `
		`+booksSelectSQL()+`WHERE b.id = ?
	`, id)
	var item book.Book
	var categoryID sql.NullInt64
	var ownerID sql.NullInt64
	var coverPath sql.NullString
	err := row.Scan(&item.ID, &item.Title, &item.Author, &ownerID, &categoryID, &item.Category, &item.Description, &item.ChapterCount, &item.LatestChapterTitle, &item.RecommendScore, &coverPath, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return book.Book{}, mapNotFound(err)
	}
	if ownerID.Valid {
		item.OwnerUserID = &ownerID.Int64
	}
	if categoryID.Valid {
		item.CategoryID = &categoryID.Int64
	}
	if coverPath.Valid {
		item.CoverPath = &coverPath.String
	}
	return item, nil
}

func (s *Store) UpdateBookCoverPath(ctx context.Context, bookID int64, coverPath *string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE books SET cover_path = ? WHERE id = ?`, coverPath, bookID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) ListChapters(ctx context.Context, bookID int64) ([]book.Chapter, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, book_id, chapter_index, title
		FROM chapters WHERE book_id = ? ORDER BY chapter_index
	`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []book.Chapter
	for rows.Next() {
		var item book.Chapter
		if err := rows.Scan(&item.ID, &item.BookID, &item.Index, &item.Title); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) FindChapter(ctx context.Context, bookID, chapterID int64) (book.Chapter, error) {
	var item book.Chapter
	err := s.db.QueryRowContext(ctx, `
		SELECT id, book_id, chapter_index, title, content
		FROM chapters WHERE book_id = ? AND id = ?
	`, bookID, chapterID).Scan(&item.ID, &item.BookID, &item.Index, &item.Title, &item.Content)
	if err != nil {
		return book.Chapter{}, mapNotFound(err)
	}
	return item, nil
}

func (s *Store) CreateUpload(ctx context.Context, item upload.Upload) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO uploads (admin_user_id, original_filename, stored_path, file_size, status)
		VALUES (?, ?, ?, ?, ?)
	`, item.ActorUserID, item.OriginalFilename, item.StoredPath, item.FileSize, item.Status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) MarkUpload(ctx context.Context, id int64, status, message string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE uploads SET status = ?, error_message = ? WHERE id = ?`, status, message, id)
	return err
}

func (s *Store) CreateBookWithChapters(ctx context.Context, input book.CreateInput, categoryID *int64, uploadID int64, chapters []book.ChapterDraft) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	latest := ""
	if len(chapters) > 0 {
		latest = chapters[len(chapters)-1].Title
	}
	var uploadValue any
	if uploadID > 0 {
		uploadValue = uploadID
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO books (title, author, owner_user_id, category_id, description, chapter_count, latest_chapter_title, source_upload_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, input.Title, legacyAuthor(input.OwnerUserID), input.OwnerUserID, categoryID, input.Description, len(chapters), latest, uploadValue)
	if err != nil {
		return 0, err
	}
	bookID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	for i, chapterItem := range chapters {
		index := chapterItem.Index
		if index == 0 {
			index = i + 1
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO chapters (book_id, chapter_index, title, content)
			VALUES (?, ?, ?, ?)
		`, bookID, index, chapterItem.Title, chapterItem.Content); err != nil {
			return 0, err
		}
	}
	return bookID, tx.Commit()
}

func (s *Store) CreateBook(ctx context.Context, input book.CreateInput, categoryID *int64) (book.Book, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO books (title, author, owner_user_id, category_id, description, chapter_count, latest_chapter_title)
		VALUES (?, ?, ?, ?, ?, 0, '')
	`, input.Title, legacyAuthor(input.OwnerUserID), input.OwnerUserID, categoryID, input.Description)
	if err != nil {
		return book.Book{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return book.Book{}, err
	}
	return s.FindBook(ctx, id)
}

func (s *Store) UpdateBookMetadata(ctx context.Context, bookID int64, input book.MetadataInput, categoryID *int64) (book.Book, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return book.Book{}, err
	}
	defer tx.Rollback()

	if err := ensureBookExists(ctx, tx, bookID); err != nil {
		return book.Book{}, err
	}
	if input.RecommendScore != nil {
		if _, err := tx.ExecContext(ctx, `
			UPDATE books
			SET title = ?, category_id = ?, description = ?, recommend_score = ?
			WHERE id = ?
		`, input.Title, categoryID, input.Description, *input.RecommendScore, bookID); err != nil {
			return book.Book{}, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE books
			SET title = ?, category_id = ?, description = ?
			WHERE id = ?
		`, input.Title, categoryID, input.Description, bookID); err != nil {
			return book.Book{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return book.Book{}, err
	}
	return s.FindBook(ctx, bookID)
}

func (s *Store) DeleteBook(ctx context.Context, bookID int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM books WHERE id = ?`, bookID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (s *Store) AddChapter(ctx context.Context, bookID int64, input book.ChapterInput) (book.Chapter, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return book.Chapter{}, err
	}
	defer tx.Rollback()

	if err := ensureBookExists(ctx, tx, bookID); err != nil {
		return book.Chapter{}, err
	}
	nextIndex := int64(1)
	var maxIndex int64
	err = tx.QueryRowContext(ctx, `
		SELECT chapter_index FROM chapters
		WHERE book_id = ?
		ORDER BY chapter_index DESC
		LIMIT 1
		FOR UPDATE
	`, bookID).Scan(&maxIndex)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return book.Chapter{}, err
	}
	if err == nil {
		nextIndex = maxIndex + 1
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO chapters (book_id, chapter_index, title, content)
		VALUES (?, ?, ?, ?)
	`, bookID, nextIndex, input.Title, input.Content)
	if err != nil {
		return book.Chapter{}, err
	}
	chapterID, err := result.LastInsertId()
	if err != nil {
		return book.Chapter{}, err
	}
	if err := refreshBookChapterStats(ctx, tx, bookID); err != nil {
		return book.Chapter{}, err
	}
	if err := tx.Commit(); err != nil {
		return book.Chapter{}, err
	}
	return s.FindChapter(ctx, bookID, chapterID)
}

func (s *Store) UpdateChapter(ctx context.Context, bookID, chapterID int64, input book.ChapterInput) (book.Chapter, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return book.Chapter{}, err
	}
	defer tx.Rollback()

	var id int64
	if err := tx.QueryRowContext(ctx, `
		SELECT id FROM chapters WHERE book_id = ? AND id = ? FOR UPDATE
	`, bookID, chapterID).Scan(&id); err != nil {
		return book.Chapter{}, mapNotFound(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE chapters SET title = ?, content = ?
		WHERE book_id = ? AND id = ?
	`, input.Title, input.Content, bookID, chapterID); err != nil {
		return book.Chapter{}, err
	}
	if err := refreshBookChapterStats(ctx, tx, bookID); err != nil {
		return book.Chapter{}, err
	}
	if err := tx.Commit(); err != nil {
		return book.Chapter{}, err
	}
	return s.FindChapter(ctx, bookID, chapterID)
}

func (s *Store) DeleteChapter(ctx context.Context, bookID, chapterID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var deletedIndex int
	if err := tx.QueryRowContext(ctx, `
		SELECT chapter_index FROM chapters WHERE book_id = ? AND id = ? FOR UPDATE
	`, bookID, chapterID).Scan(&deletedIndex); err != nil {
		return mapNotFound(err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM chapters WHERE book_id = ? AND id = ?`, bookID, chapterID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE chapters SET chapter_index = chapter_index - 1
		WHERE book_id = ? AND chapter_index > ?
		ORDER BY chapter_index ASC
	`, bookID, deletedIndex); err != nil {
		return err
	}
	if err := refreshBookChapterStats(ctx, tx, bookID); err != nil {
		return err
	}
	return tx.Commit()
}

func scanUser(row *sql.Row) (identity.User, error) {
	var user identity.User
	err := row.Scan(&user.ID, &user.Username, &user.Nickname, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return identity.User{}, mapNotFound(err)
	}
	return user, nil
}

type txQueryer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func ensureBookExists(ctx context.Context, tx txQueryer, bookID int64) error {
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM books WHERE id = ? FOR UPDATE`, bookID).Scan(&id)
	return mapNotFound(err)
}

func refreshBookChapterStats(ctx context.Context, tx txQueryer, bookID int64) error {
	var count int
	var latest string
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE((
			SELECT title FROM chapters
			WHERE book_id = ?
			ORDER BY chapter_index DESC
			LIMIT 1
		), '')
		FROM chapters WHERE book_id = ?
	`, bookID, bookID).Scan(&count, &latest); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE books SET chapter_count = ?, latest_chapter_title = ?
		WHERE id = ?
	`, count, latest, bookID); err != nil {
		return err
	}
	return nil
}

func scanBooks(rows *sql.Rows) ([]book.Book, error) {
	items := []book.Book{}
	for rows.Next() {
		var item book.Book
		var ownerID sql.NullInt64
		var categoryID sql.NullInt64
		var coverPath sql.NullString
		if err := rows.Scan(&item.ID, &item.Title, &item.Author, &ownerID, &categoryID, &item.Category, &item.Description, &item.ChapterCount, &item.LatestChapterTitle, &item.RecommendScore, &coverPath, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if ownerID.Valid {
			item.OwnerUserID = &ownerID.Int64
		}
		if categoryID.Valid {
			item.CategoryID = &categoryID.Int64
		}
		if coverPath.Valid {
			item.CoverPath = &coverPath.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func authorExprSQL() string {
	return "CASE WHEN u.role = 'admin' THEN '系统' ELSE COALESCE(u.username, b.author) END"
}

func booksSelectSQL() string {
	return `
		SELECT b.id, b.title, ` + authorExprSQL() + `, b.owner_user_id, b.category_id, COALESCE(c.name, ''), b.description,
		       b.chapter_count, b.latest_chapter_title, b.recommend_score, b.cover_path, b.created_at, b.updated_at
		FROM books b
		LEFT JOIN users u ON u.id = b.owner_user_id
		LEFT JOIN categories c ON c.id = b.category_id
	`
}

func booksBaseCountSQL() string {
	return `
		SELECT COUNT(*)
		FROM books b
		LEFT JOIN users u ON u.id = b.owner_user_id
		LEFT JOIN categories c ON c.id = b.category_id
	`
}

func legacyAuthor(ownerUserID int64) string {
	if ownerUserID <= 0 {
		return "系统"
	}
	return ""
}

func (s *Store) ensureColumn(ctx context.Context, tableName, columnName, alterSQL string) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, tableName, columnName).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, alterSQL); err != nil {
		return fmt.Errorf("alter %s add %s: %w", tableName, columnName, err)
	}
	return nil
}

func (s *Store) ensureIndex(ctx context.Context, tableName, indexName, alterSQL string) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?
	`, tableName, indexName).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, alterSQL); err != nil {
		return fmt.Errorf("alter %s add index %s: %w", tableName, indexName, err)
	}
	return nil
}

func (s *Store) ensureForeignKey(ctx context.Context, tableName, constraintName, alterSQL string) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.TABLE_CONSTRAINTS
		WHERE CONSTRAINT_SCHEMA = DATABASE() AND TABLE_NAME = ? AND CONSTRAINT_NAME = ?
	`, tableName, constraintName).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, alterSQL); err != nil {
		return fmt.Errorf("alter %s add constraint %s: %w", tableName, constraintName, err)
	}
	return nil
}

func splitSQL(sqlText string) []string {
	raw := strings.Split(sqlText, ";")
	statements := make([]string, 0, len(raw))
	for _, stmt := range raw {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}
	return statements
}

func mapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return shared.ErrNotFound
	}
	return err
}
