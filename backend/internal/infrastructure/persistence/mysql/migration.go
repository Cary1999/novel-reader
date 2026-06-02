package mysql

import (
	"context"
	"database/sql"
)

func migrate(ctx context.Context, db *sql.DB) error {
	for _, stmt := range splitSQL(schemaSQL) {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	if err := ensureColumn(ctx, db, "users", "nickname", `ALTER TABLE users ADD COLUMN nickname VARCHAR(64) NOT NULL DEFAULT '' AFTER username`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "books", "owner_user_id", `ALTER TABLE books ADD COLUMN owner_user_id BIGINT NULL AFTER author`); err != nil {
		return err
	}
	if err := ensureIndex(ctx, db, "books", "idx_books_owner_user_id", `ALTER TABLE books ADD INDEX idx_books_owner_user_id (owner_user_id)`); err != nil {
		return err
	}
	if err := ensureForeignKey(ctx, db, "books", "fk_books_owner_user", `ALTER TABLE books ADD CONSTRAINT fk_books_owner_user FOREIGN KEY (owner_user_id) REFERENCES users(id)`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "books", "recommend_score", `ALTER TABLE books ADD COLUMN recommend_score INT NOT NULL DEFAULT 0 AFTER latest_chapter_title`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "books", "cover_path", `ALTER TABLE books ADD COLUMN cover_path VARCHAR(255) NULL AFTER recommend_score`); err != nil {
		return err
	}
	if err := ensureIndex(ctx, db, "books", "idx_books_recommend_score", `ALTER TABLE books ADD INDEX idx_books_recommend_score (recommend_score)`); err != nil {
		return err
	}
	return nil
}
