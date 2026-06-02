package identity

import (
	"context"
	"time"

	"novel-reader/backend/internal/domain/shared"
)

type User struct {
	ID           int64       `json:"id"`
	Username     string      `json:"username"`
	Nickname     string      `json:"nickname"`
	PasswordHash string      `json:"-"`
	Role         shared.Role `json:"role"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type UserRepository interface {
	CreateUser(ctx context.Context, username, passwordHash string, role shared.Role) (User, error)
	FindUserByUsername(ctx context.Context, username string) (User, error)
	FindUserByID(ctx context.Context, id int64) (User, error)
	UpdateUserNickname(ctx context.Context, id int64, nickname string) (User, error)
	UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error
}
