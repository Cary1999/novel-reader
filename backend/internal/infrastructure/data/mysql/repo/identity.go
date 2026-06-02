package repo

import (
	"context"
	"database/sql"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	"novel-reader/backend/internal/domain/shared"
	"novel-reader/backend/internal/infrastructure/data/mysql/model"
)

type IdentityRepository struct {
	db *sql.DB
}

func NewIdentityRepository(db *sql.DB) IdentityRepository {
	return IdentityRepository{db: db}
}

func (r IdentityRepository) CreateUser(ctx context.Context, username, passwordHash string, role shared.Role) (identityentity.User, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (username, nickname, password_hash, role) VALUES (?, ?, ?, ?)
	`, username, username, passwordHash, role)
	if err != nil {
		return identityentity.User{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return identityentity.User{}, err
	}
	return r.FindUserByID(ctx, id)
}

func (r IdentityRepository) FindUserByUsername(ctx context.Context, username string) (identityentity.User, error) {
	var record model.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, nickname, password_hash, role, created_at, updated_at
		FROM users WHERE username = ?
	`, username).Scan(&record.ID, &record.Username, &record.Nickname, &record.PasswordHash, &record.Role, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return identityentity.User{}, mapNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r IdentityRepository) FindUserByID(ctx context.Context, id int64) (identityentity.User, error) {
	var record model.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, nickname, password_hash, role, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(&record.ID, &record.Username, &record.Nickname, &record.PasswordHash, &record.Role, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return identityentity.User{}, mapNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r IdentityRepository) UpdateUserNickname(ctx context.Context, id int64, nickname string) (identityentity.User, error) {
	if _, err := r.FindUserByID(ctx, id); err != nil {
		return identityentity.User{}, err
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE users SET nickname = ? WHERE id = ?`, nickname, id); err != nil {
		return identityentity.User{}, err
	}
	return r.FindUserByID(ctx, id)
}

func (r IdentityRepository) UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, id)
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
