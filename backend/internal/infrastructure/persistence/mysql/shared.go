package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func legacyAuthor(ownerUserID int64) string {
	if ownerUserID <= 0 {
		return "系统"
	}
	return ""
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

func ensureColumn(ctx context.Context, db *sql.DB, tableName, columnName, alterSQL string) error {
	count, err := columnCount(ctx, db, tableName, columnName)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if _, err := db.ExecContext(ctx, alterSQL); err != nil {
		return fmt.Errorf("alter %s add %s: %w", tableName, columnName, err)
	}
	return nil
}

func ensureColumnAlter(ctx context.Context, db *sql.DB, tableName, columnName, alterSQL string) error {
	count, err := columnCount(ctx, db, tableName, columnName)
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	if _, err := db.ExecContext(ctx, alterSQL); err != nil {
		return fmt.Errorf("alter %s modify %s: %w", tableName, columnName, err)
	}
	return nil
}

func hasColumn(ctx context.Context, db *sql.DB, tableName, columnName string) (bool, error) {
	count, err := columnCount(ctx, db, tableName, columnName)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func columnCount(ctx context.Context, db *sql.DB, tableName, columnName string) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, tableName, columnName).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func ensureIndex(ctx context.Context, db *sql.DB, tableName, indexName, alterSQL string) error {
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?
	`, tableName, indexName).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if _, err := db.ExecContext(ctx, alterSQL); err != nil {
		return fmt.Errorf("alter %s add index %s: %w", tableName, indexName, err)
	}
	return nil
}

func ensureForeignKey(ctx context.Context, db *sql.DB, tableName, constraintName, alterSQL string) error {
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.TABLE_CONSTRAINTS
		WHERE CONSTRAINT_SCHEMA = DATABASE() AND TABLE_NAME = ? AND CONSTRAINT_NAME = ?
	`, tableName, constraintName).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if _, err := db.ExecContext(ctx, alterSQL); err != nil {
		return fmt.Errorf("alter %s add constraint %s: %w", tableName, constraintName, err)
	}
	return nil
}
