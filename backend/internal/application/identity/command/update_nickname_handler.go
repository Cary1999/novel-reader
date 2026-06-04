package command

import (
	"context"
	"net/http"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	identityservice "novel-reader/backend/internal/domain/identity/service"
	"novel-reader/backend/internal/domain/shared"
)

type UpdateNickname struct {
	Nickname string `json:"nickname"`
}

type UpdateNicknameHandler struct {
	users   identityrepository.UserRepository
	service *identityservice.IdentityService
}

func NewUpdateNicknameHandler(users identityrepository.UserRepository, service *identityservice.IdentityService) *UpdateNicknameHandler {
	return &UpdateNicknameHandler{users: users, service: service}
}

func (h *UpdateNicknameHandler) Handle(ctx context.Context, actor shared.Actor, input UpdateNickname) (identityentity.User, error) {
	nickname, err := h.service.NormalizeNickname(input.Nickname)
	if err != nil {
		return identityentity.User{}, err
	}
	user, err := h.users.UpdateUserNickname(ctx, actor.ActorID, nickname)
	if err != nil {
		if err == shared.ErrNotFound {
			return identityentity.User{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return identityentity.User{}, err
	}
	return user, nil
}
