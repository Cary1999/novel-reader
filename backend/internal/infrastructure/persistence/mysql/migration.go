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
	if err := ensureColumn(ctx, db, "site_settings", "updated_by_operator_id", `ALTER TABLE site_settings ADD COLUMN updated_by_operator_id BIGINT NULL AFTER hero_description`); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "uploads", "actor_user_id", `ALTER TABLE uploads ADD COLUMN actor_user_id BIGINT NULL AFTER id`); err != nil {
		return err
	}
	if err := ensureIndex(ctx, db, "books", "idx_books_recommend_score", `ALTER TABLE books ADD INDEX idx_books_recommend_score (recommend_score)`); err != nil {
		return err
	}
	if ok, err := hasColumn(ctx, db, "uploads", "admin_user_id"); err != nil {
		return err
	} else if ok {
		if _, err := db.ExecContext(ctx, `UPDATE uploads SET actor_user_id = admin_user_id WHERE actor_user_id IS NULL`); err != nil {
			return err
		}
	}
	if _, err := db.ExecContext(ctx, `UPDATE site_settings SET updated_by_operator_id = NULL WHERE updated_by_operator_id IS NULL`); err != nil {
		return err
	}
	if err := ensureColumnAlter(ctx, db, "site_settings", "top_line_title", `ALTER TABLE site_settings MODIFY COLUMN top_line_title VARCHAR(120) NOT NULL DEFAULT ''`); err != nil {
		return err
	}
	if err := ensureColumnAlter(ctx, db, "site_settings", "top_line_tagline", `ALTER TABLE site_settings MODIFY COLUMN top_line_tagline VARCHAR(255) NOT NULL DEFAULT ''`); err != nil {
		return err
	}
	return nil
}
