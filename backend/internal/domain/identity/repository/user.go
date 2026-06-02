package repository

import (
	"context"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	"novel-reader/backend/internal/domain/shared"
)

type UserRepository interface {
	CreateUser(ctx context.Context, username, passwordHash string, role shared.Role) (identityentity.User, error)
	FindUserByUsername(ctx context.Context, username string) (identityentity.User, error)
	FindUserByID(ctx context.Context, id int64) (identityentity.User, error)
	UpdateUserNickname(ctx context.Context, id int64, nickname string) (identityentity.User, error)
	UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error
}
