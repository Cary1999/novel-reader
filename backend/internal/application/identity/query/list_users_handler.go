package query

import (
	"context"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
)

type ListUsersHandler struct {
	users identityrepository.UserRepository
}

func NewListUsersHandler(users identityrepository.UserRepository) *ListUsersHandler {
	return &ListUsersHandler{users: users}
}

func (h *ListUsersHandler) Handle(ctx context.Context) ([]identityentity.FrontUserSummary, error) {
	return h.users.ListUsers(ctx)
}
