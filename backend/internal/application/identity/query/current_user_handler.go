package query

import (
	"context"
	"net/http"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	"novel-reader/backend/internal/domain/shared"
)

type CurrentUserHandler struct {
	users identityrepository.UserRepository
}

func NewCurrentUserHandler(users identityrepository.UserRepository) *CurrentUserHandler {
	return &CurrentUserHandler{users: users}
}

func (h *CurrentUserHandler) Handle(ctx context.Context, actor shared.Actor) (identityentity.User, error) {
	user, err := h.users.FindUserByID(ctx, actor.ActorID)
	if err != nil {
		if err == shared.ErrNotFound {
			return identityentity.User{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return identityentity.User{}, err
	}
	return user, nil
}
