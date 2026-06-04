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
	FindUserAvatarByID(ctx context.Context, id int64) (string, error)
	ListUsers(ctx context.Context) ([]identityentity.FrontUserSummary, error)
	UpdateUserNickname(ctx context.Context, id int64, nickname string) (identityentity.User, error)
	UpdateUserAvatar(ctx context.Context, id int64, avatarPath string) (identityentity.User, error)
	UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error
	PromoteUserToAuthor(ctx context.Context, id int64) (identityentity.User, error)
}

type OperatorRepository interface {
	CreateOperator(ctx context.Context, username, passwordHash string, role shared.Role) (identityentity.Operator, error)
	FindOperatorByUsername(ctx context.Context, username string) (identityentity.Operator, error)
	FindOperatorByID(ctx context.Context, id int64) (identityentity.Operator, error)
	ListOperators(ctx context.Context) ([]identityentity.Operator, error)
	UpdateOperatorRole(ctx context.Context, id int64, role shared.Role) (identityentity.Operator, error)
	UpdateOperatorPassword(ctx context.Context, id int64, passwordHash string) error
}

type AuthorApplicationRepository interface {
	CreateAuthorApplication(ctx context.Context, item identityentity.AuthorApplication) (identityentity.AuthorApplication, error)
	FindLatestAuthorApplicationByUserID(ctx context.Context, userID int64) (identityentity.AuthorApplication, error)
	ListAuthorApplications(ctx context.Context) ([]identityentity.AuthorApplication, error)
	ReviewAuthorApplication(ctx context.Context, applicationID int64, status, reviewNote string, reviewedByOperatorID int64) (identityentity.AuthorApplication, error)
}
