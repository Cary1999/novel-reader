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
	if err := ensureColumn(ctx, db, "users", "avatar_path", `ALTER TABLE users ADD COLUMN avatar_path VARCHAR(255) NULL AFTER nickname`); err != nil {
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
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bookshelf_groups (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT NOT NULL,
			name VARCHAR(64) NOT NULL,
			sort_order INT NOT NULL DEFAULT 0,
			is_default TINYINT(1) NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_bookshelf_groups_user_name (user_id, name),
			INDEX idx_bookshelf_groups_user_sort (user_id, sort_order, id),
			CONSTRAINT fk_bookshelf_groups_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bookshelf_items (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT NOT NULL,
			book_id BIGINT NOT NULL,
			group_id BIGINT NOT NULL,
			is_pinned TINYINT(1) NOT NULL DEFAULT 0,
			pinned_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_bookshelf_items_user_book (user_id, book_id),
			INDEX idx_bookshelf_items_group_pinned (group_id, is_pinned, pinned_at, created_at),
			INDEX idx_bookshelf_items_user_created (user_id, created_at),
			CONSTRAINT fk_bookshelf_items_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			CONSTRAINT fk_bookshelf_items_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
			CONSTRAINT fk_bookshelf_items_group FOREIGN KEY (group_id) REFERENCES bookshelf_groups(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`); err != nil {
		return err
	}
	return nil
}
