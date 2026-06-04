package mysql

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"

	bookentity "novel-reader/backend/internal/domain/book/entity"
)

func seed(ctx context.Context, db *sql.DB, opts SeedOptions) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(opts.SuperAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET role = 'reader' WHERE role NOT IN ('reader', 'author')`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO operators (username, password_hash, role)
		VALUES (?, ?, 'super_admin')
		ON DUPLICATE KEY UPDATE username = username
	`, opts.SuperAdminUsername, string(hash)); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET nickname = username WHERE nickname = ''`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE books b
		LEFT JOIN users u ON u.username = b.author
		SET b.owner_user_id = u.id
		WHERE b.owner_user_id IS NULL AND u.id IS NOT NULL
	`); err != nil {
		return err
	}

	categories := []string{"玄幻", "都市", "科幻", "历史", "游戏"}
	for _, name := range categories {
		if _, err := db.ExecContext(ctx, `INSERT IGNORE INTO categories (name) VALUES (?)`, name); err != nil {
			return err
		}
	}

	var existing int
	var authorID int64
	if err := db.QueryRowContext(ctx, `
		SELECT id FROM users WHERE role = 'author' ORDER BY id ASC LIMIT 1
	`).Scan(&authorID); err != nil {
		if err == sql.ErrNoRows {
			readerHash, hashErr := bcrypt.GenerateFromPassword([]byte("author1234"), bcrypt.DefaultCost)
			if hashErr != nil {
				return hashErr
			}
			result, insertErr := db.ExecContext(ctx, `
				INSERT INTO users (username, nickname, password_hash, role)
				VALUES ('demo_author', 'demo_author', ?, 'author')
			`, string(readerHash))
			if insertErr != nil {
				return insertErr
			}
			authorID, err = result.LastInsertId()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books WHERE title = ? AND owner_user_id = ?`, "星河书页", authorID).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	var categoryID int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM categories WHERE name = ?`, "科幻").Scan(&categoryID); err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	chapters := []bookentity.ChapterDraft{
		{Index: 1, Title: "第一章 本地书架", Content: "清晨的终端亮起，新的书架在本地服务中安静展开。"},
		{Index: 2, Title: "Chapter 2 The Reader", Content: "A reader signs in and opens the next page without touching any remote content."},
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO books (title, author, owner_user_id, category_id, description, chapter_count, latest_chapter_title)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "星河书页", "demo_author", authorID, &categoryID, "用于本地 smoke check 的示例小说。", len(chapters), chapters[len(chapters)-1].Title)
	if err != nil {
		return err
	}
	bookID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for _, chapterItem := range chapters {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO chapters (book_id, chapter_index, title, content)
			VALUES (?, ?, ?, ?)
		`, bookID, chapterItem.Index, chapterItem.Title, chapterItem.Content); err != nil {
			return err
		}
	}
	return tx.Commit()
}
