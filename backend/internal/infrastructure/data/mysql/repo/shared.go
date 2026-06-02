package repo

import (
	"context"
	"database/sql"
	"errors"

	"novel-reader/backend/internal/domain/shared"
)

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

func mapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return shared.ErrNotFound
	}
	return err
}
