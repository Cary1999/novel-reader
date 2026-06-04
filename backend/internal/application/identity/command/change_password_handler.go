package command

import (
	"context"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	identityservice "novel-reader/backend/internal/domain/identity/service"
	"novel-reader/backend/internal/domain/shared"
)

type ChangePassword struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type ChangePasswordHandler struct {
	users   identityrepository.UserRepository
	service *identityservice.IdentityService
}

func NewChangePasswordHandler(users identityrepository.UserRepository, service *identityservice.IdentityService) *ChangePasswordHandler {
	return &ChangePasswordHandler{users: users, service: service}
}

func (h *ChangePasswordHandler) Handle(ctx context.Context, actor shared.Actor, input ChangePassword) error {
	if err := h.service.ValidatePasswordLength(input.NewPassword); err != nil {
		return err
	}
	user, err := h.users.FindUserByID(ctx, actor.ActorID)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.OldPassword)) != nil {
		return shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := h.users.UpdateUserPassword(ctx, actor.ActorID, string(hash)); err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return err
	}
	return nil
}
