package command

import (
	"context"
	"net/http"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	"novel-reader/backend/internal/domain/shared"
)

type PromoteUserToAuthorHandler struct {
	users identityrepository.UserRepository
}

func NewPromoteUserToAuthorHandler(users identityrepository.UserRepository) *PromoteUserToAuthorHandler {
	return &PromoteUserToAuthorHandler{users: users}
}

func (h *PromoteUserToAuthorHandler) Handle(ctx context.Context, actor shared.Actor, userID int64) (identityentity.User, error) {
	if !actor.IsReviewer() {
		return identityentity.User{}, shared.NewError(http.StatusForbidden, "FORBIDDEN", "reviewer required")
	}
	user, err := h.users.PromoteUserToAuthor(ctx, userID)
	if err != nil {
		if err == shared.ErrNotFound {
			return identityentity.User{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "user not found")
		}
		return identityentity.User{}, err
	}
	if user.Role != shared.RoleAuthor {
		return identityentity.User{}, shared.NewError(http.StatusInternalServerError, "INTERNAL", "promote user failed")
	}
	return user, nil
}
