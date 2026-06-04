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
		SELECT id, username, nickname, avatar_path, password_hash, role, created_at, updated_at
		FROM users WHERE username = ?
	`, username).Scan(&record.ID, &record.Username, &record.Nickname, &record.AvatarPath, &record.PasswordHash, &record.Role, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return identityentity.User{}, mapNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r IdentityRepository) FindUserByID(ctx context.Context, id int64) (identityentity.User, error) {
	var record model.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, nickname, avatar_path, password_hash, role, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(&record.ID, &record.Username, &record.Nickname, &record.AvatarPath, &record.PasswordHash, &record.Role, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return identityentity.User{}, mapNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r IdentityRepository) ListUsers(ctx context.Context) ([]identityentity.FrontUserSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.username, u.nickname, u.role,
		       la.id, COALESCE(la.status, ''), la.created_at, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN (
			SELECT a1.id, a1.user_id, a1.status, a1.created_at
			FROM author_applications a1
			INNER JOIN (
				SELECT user_id, MAX(id) AS max_id
				FROM author_applications
				GROUP BY user_id
			) latest ON latest.max_id = a1.id
		) la ON la.user_id = u.id
		ORDER BY u.created_at DESC, u.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []identityentity.FrontUserSummary
	for rows.Next() {
		var record model.FrontUserSummary
		var latestID sql.NullInt64
		var latestStatus string
		var latestCreatedAt sql.NullTime
		if err := rows.Scan(
			&record.ID,
			&record.Username,
			&record.Nickname,
			&record.Role,
			&latestID,
			&latestStatus,
			&latestCreatedAt,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if latestID.Valid {
			record.LatestApplicationID = &latestID.Int64
		}
		if latestCreatedAt.Valid {
			record.LatestApplicationAt = &latestCreatedAt.Time
		}
		record.LatestApplication = latestStatus
		items = append(items, record.ToEntity())
	}
	return items, rows.Err()
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

func (r IdentityRepository) UpdateUserAvatar(ctx context.Context, id int64, avatarPath string) (identityentity.User, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE users SET avatar_path = ? WHERE id = ?`, avatarPath, id)
	if err != nil {
		return identityentity.User{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return identityentity.User{}, err
	}
	if affected == 0 {
		return identityentity.User{}, shared.ErrNotFound
	}
	return r.FindUserByID(ctx, id)
}

func (r IdentityRepository) FindUserAvatarByID(ctx context.Context, id int64) (string, error) {
	var avatar sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT avatar_path FROM users WHERE id = ?`, id).Scan(&avatar)
	if err != nil {
		return "", mapNotFound(err)
	}
	if !avatar.Valid {
		return "", shared.ErrNotFound
	}
	return avatar.String, nil
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

func (r IdentityRepository) PromoteUserToAuthor(ctx context.Context, id int64) (identityentity.User, error) {
	if _, err := r.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE id = ?`, shared.RoleAuthor, id); err != nil {
		return identityentity.User{}, err
	}
	return r.FindUserByID(ctx, id)
}

func (r IdentityRepository) CreateOperator(ctx context.Context, username, passwordHash string, role shared.Role) (identityentity.Operator, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO operators (username, password_hash, role) VALUES (?, ?, ?)
	`, username, passwordHash, role)
	if err != nil {
		return identityentity.Operator{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return identityentity.Operator{}, err
	}
	return r.FindOperatorByID(ctx, id)
}

func (r IdentityRepository) FindOperatorByUsername(ctx context.Context, username string) (identityentity.Operator, error) {
	var record model.Operator
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM operators WHERE username = ?
	`, username).Scan(&record.ID, &record.Username, &record.PasswordHash, &record.Role, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return identityentity.Operator{}, mapNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r IdentityRepository) FindOperatorByID(ctx context.Context, id int64) (identityentity.Operator, error) {
	var record model.Operator
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM operators WHERE id = ?
	`, id).Scan(&record.ID, &record.Username, &record.PasswordHash, &record.Role, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return identityentity.Operator{}, mapNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r IdentityRepository) ListOperators(ctx context.Context) ([]identityentity.Operator, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, password_hash, role, created_at, updated_at
		FROM operators
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []identityentity.Operator
	for rows.Next() {
		var record model.Operator
		if err := rows.Scan(&record.ID, &record.Username, &record.PasswordHash, &record.Role, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, record.ToEntity())
	}
	return items, rows.Err()
}

func (r IdentityRepository) UpdateOperatorRole(ctx context.Context, id int64, role shared.Role) (identityentity.Operator, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE operators SET role = ? WHERE id = ?`, role, id)
	if err != nil {
		return identityentity.Operator{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return identityentity.Operator{}, err
	}
	if affected == 0 {
		return identityentity.Operator{}, shared.ErrNotFound
	}
	return r.FindOperatorByID(ctx, id)
}

func (r IdentityRepository) UpdateOperatorPassword(ctx context.Context, id int64, passwordHash string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE operators SET password_hash = ? WHERE id = ?`, passwordHash, id)
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

func (r IdentityRepository) CreateAuthorApplication(ctx context.Context, item identityentity.AuthorApplication) (identityentity.AuthorApplication, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO author_applications (user_id, pen_name, reason, status, review_note)
		VALUES (?, ?, ?, ?, ?)
	`, item.UserID, item.PenName, item.Reason, item.Status, item.ReviewNote)
	if err != nil {
		return identityentity.AuthorApplication{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return identityentity.AuthorApplication{}, err
	}
	return r.findAuthorApplicationByID(ctx, id)
}

func (r IdentityRepository) FindLatestAuthorApplicationByUserID(ctx context.Context, userID int64) (identityentity.AuthorApplication, error) {
	return r.findAuthorApplication(ctx, `
		SELECT a.id, a.user_id, u.username, u.nickname, a.pen_name, a.reason, a.status, a.review_note,
		       a.reviewed_by_operator_id, a.reviewed_at, a.created_at, a.updated_at
		FROM author_applications a
		INNER JOIN users u ON u.id = a.user_id
		WHERE a.user_id = ?
		ORDER BY a.id DESC
		LIMIT 1
	`, userID)
}

func (r IdentityRepository) ListAuthorApplications(ctx context.Context) ([]identityentity.AuthorApplication, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.user_id, u.username, u.nickname, a.pen_name, a.reason, a.status, a.review_note,
		       a.reviewed_by_operator_id, a.reviewed_at, a.created_at, a.updated_at
		FROM author_applications a
		INNER JOIN users u ON u.id = a.user_id
		ORDER BY
		  CASE a.status WHEN 'pending' THEN 0 WHEN 'approved' THEN 1 ELSE 2 END,
		  a.created_at DESC,
		  a.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []identityentity.AuthorApplication
	for rows.Next() {
		item, err := scanAuthorApplication(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r IdentityRepository) ReviewAuthorApplication(ctx context.Context, applicationID int64, status, reviewNote string, reviewedByOperatorID int64) (identityentity.AuthorApplication, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE author_applications
		SET status = ?, review_note = ?, reviewed_by_operator_id = ?, reviewed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, reviewNote, reviewedByOperatorID, applicationID)
	if err != nil {
		return identityentity.AuthorApplication{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return identityentity.AuthorApplication{}, err
	}
	if affected == 0 {
		return identityentity.AuthorApplication{}, shared.ErrNotFound
	}
	return r.findAuthorApplicationByID(ctx, applicationID)
}

func (r IdentityRepository) findAuthorApplicationByID(ctx context.Context, id int64) (identityentity.AuthorApplication, error) {
	return r.findAuthorApplication(ctx, `
		SELECT a.id, a.user_id, u.username, u.nickname, a.pen_name, a.reason, a.status, a.review_note,
		       a.reviewed_by_operator_id, a.reviewed_at, a.created_at, a.updated_at
		FROM author_applications a
		INNER JOIN users u ON u.id = a.user_id
		WHERE a.id = ?
	`, id)
}

func (r IdentityRepository) findAuthorApplication(ctx context.Context, query string, args ...any) (identityentity.AuthorApplication, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	return scanAuthorApplication(row)
}

type applicationScanner interface {
	Scan(dest ...any) error
}

func scanAuthorApplication(scanner applicationScanner) (identityentity.AuthorApplication, error) {
	var record model.AuthorApplication
	var reviewedBy sql.NullInt64
	var reviewedAt sql.NullTime
	err := scanner.Scan(
		&record.ID,
		&record.UserID,
		&record.Username,
		&record.Nickname,
		&record.PenName,
		&record.Reason,
		&record.Status,
		&record.ReviewNote,
		&reviewedBy,
		&reviewedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return identityentity.AuthorApplication{}, mapNotFound(err)
	}
	if reviewedBy.Valid {
		record.ReviewedByOperatorID = &reviewedBy.Int64
	}
	if reviewedAt.Valid {
		at := reviewedAt.Time.UTC().Local()
		record.ReviewedAt = &at
	}
	return record.ToEntity(), nil
}
