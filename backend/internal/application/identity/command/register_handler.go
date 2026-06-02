package command

import (
	"context"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	identityservice "novel-reader/backend/internal/domain/identity/service"
	"novel-reader/backend/internal/domain/shared"
)

type Register struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterHandler struct {
	users   identityrepository.UserRepository
	service *identityservice.IdentityService
}

func NewRegisterHandler(users identityrepository.UserRepository, service *identityservice.IdentityService) *RegisterHandler {
	return &RegisterHandler{users: users, service: service}
}

func (h *RegisterHandler) Handle(ctx context.Context, input Register) (identityentity.User, error) {
	username, err := h.service.NormalizeUsername(input.Username)
	if err != nil {
		return identityentity.User{}, err
	}
	if err := h.service.ValidatePasswordLength(input.Password); err != nil {
		return identityentity.User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return identityentity.User{}, err
	}
	user := h.service.NewRegisteredUser(username, string(hash), shared.RoleUser)
	item, err := h.users.CreateUser(ctx, user.Username, user.PasswordHash, user.Role)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return identityentity.User{}, shared.NewError(http.StatusConflict, "CONFLICT", "username already exists")
		}
		return identityentity.User{}, err
	}
	return item, nil
}
