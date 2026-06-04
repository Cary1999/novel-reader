package query

import (
	"context"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
)

type GetUserHandler struct {
	users identityrepository.UserRepository
}

func NewGetUserHandler(users identityrepository.UserRepository) *GetUserHandler {
	return &GetUserHandler{users: users}
}

func (h *GetUserHandler) Handle(ctx context.Context, id int64) (identityentity.User, error) {
	user, err := h.users.FindUserByID(ctx, id)
	if err != nil {
		return identityentity.User{}, err
	}
	return user, nil
}
