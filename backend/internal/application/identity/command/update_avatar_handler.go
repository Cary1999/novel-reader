package command

import (
	"context"
	"errors"
	"io"
	"net/http"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	"novel-reader/backend/internal/domain/shared"
)

type SavedAvatar struct {
	RelativePath string
	Size         int64
}

type AvatarStore interface {
	SaveAvatar(originalName string, reader io.Reader) (SavedAvatar, error)
}

type UpdateAvatarHandler struct {
	users     identityrepository.UserRepository
	avatarStore AvatarStore
}

func NewUpdateAvatarHandler(users identityrepository.UserRepository, avatarStore AvatarStore) *UpdateAvatarHandler {
	return &UpdateAvatarHandler{users: users, avatarStore: avatarStore}
}

func (h *UpdateAvatarHandler) Handle(ctx context.Context, actor shared.Actor, originalName string, reader io.Reader) (identityentity.User, error) {
	saved, err := h.avatarStore.SaveAvatar(originalName, reader)
	if err != nil {
		switch {
		case errors.Is(err, io.EOF):
			return identityentity.User{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "empty file")
		case err.Error() == "invalid file type" || err.Error() == "file too large" || err.Error() == "empty file":
			return identityentity.User{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", err.Error())
		default:
			return identityentity.User{}, err
		}
	}
	user, err := h.users.UpdateUserAvatar(ctx, actor.ActorID, saved.RelativePath)
	if err != nil {
		return identityentity.User{}, err
	}
	return user, nil
}
