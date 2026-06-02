package command

import (
	"context"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	"novel-reader/backend/internal/domain/shared"
)

type TokenIssuer interface {
	Issue(user identityentity.User) (string, error)
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResult struct {
	Token string
	User  identityentity.User
}

type LoginHandler struct {
	users  identityrepository.UserRepository
	tokens TokenIssuer
}

func NewLoginHandler(users identityrepository.UserRepository, tokens TokenIssuer) *LoginHandler {
	return &LoginHandler{users: users, tokens: tokens}
}

func (h *LoginHandler) Handle(ctx context.Context, input Login) (LoginResult, error) {
	user, err := h.users.FindUserByUsername(ctx, strings.TrimSpace(input.Username))
	if err != nil {
		if err == shared.ErrNotFound {
			return LoginResult{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
		}
		return LoginResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return LoginResult{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
	}
	token, err := h.tokens.Issue(user)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Token: token, User: user}, nil
}
