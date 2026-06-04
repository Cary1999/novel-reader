package mysql

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

func seed(ctx context.Context, db *sql.DB, opts SeedOptions) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(opts.SuperAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO operators (username, password_hash, role)
		VALUES (?, ?, 'super_admin')
		ON DUPLICATE KEY UPDATE username = username
	`, opts.SuperAdminUsername, string(hash)); err != nil {
		return err
	}

	categories := []string{"玄幻", "都市", "科幻", "历史", "游戏"}
	for _, name := range categories {
		if _, err := db.ExecContext(ctx, `INSERT IGNORE INTO categories (name) VALUES (?)`, name); err != nil {
			return err
		}
	}
	return nil
}
