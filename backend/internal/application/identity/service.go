package identityapp

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"novel-reader/backend/internal/domain/identity"
	"novel-reader/backend/internal/domain/shared"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`)

type TokenIssuer interface {
	Issue(user identity.User) (string, error)
}

type Service struct {
	users  identity.UserRepository
	tokens TokenIssuer
}

func NewService(users identity.UserRepository, tokens TokenIssuer) *Service {
	return &Service{users: users, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, username, password string) (identity.User, error) {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return identity.User{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "username must be 3-32 letters, numbers, or underscores")
	}
	if len(password) < 8 || len(password) > 128 {
		return identity.User{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "password must be 8-128 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return identity.User{}, err
	}
	user, err := s.users.CreateUser(ctx, username, string(hash), shared.RoleUser)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return identity.User{}, shared.NewError(http.StatusConflict, "CONFLICT", "username already exists")
		}
		return identity.User{}, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (string, identity.User, error) {
	user, err := s.users.FindUserByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if err == shared.ErrNotFound {
			return "", identity.User{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
		}
		return "", identity.User{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", identity.User{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
	}
	token, err := s.tokens.Issue(user)
	if err != nil {
		return "", identity.User{}, err
	}
	return token, user, nil
}

func (s *Service) CurrentUser(ctx context.Context, actor shared.Actor) (identity.User, error) {
	user, err := s.users.FindUserByID(ctx, actor.UserID)
	if err != nil {
		if err == shared.ErrNotFound {
			return identity.User{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return identity.User{}, err
	}
	return user, nil
}

func (s *Service) UpdateNickname(ctx context.Context, actor shared.Actor, nickname string) (identity.User, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return identity.User{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "nickname is required")
	}
	if len([]rune(nickname)) > 64 {
		return identity.User{}, shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "nickname is too long")
	}

	user, err := s.users.UpdateUserNickname(ctx, actor.UserID, nickname)
	if err != nil {
		if err == shared.ErrNotFound {
			return identity.User{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return identity.User{}, err
	}
	return user, nil
}

func (s *Service) ChangePassword(ctx context.Context, actor shared.Actor, oldPassword, newPassword string) error {
	if len(newPassword) < 8 || len(newPassword) > 128 {
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "password must be 8-128 characters")
	}
	user, err := s.users.FindUserByID(ctx, actor.UserID)
	if err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)) != nil {
		return shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.users.UpdateUserPassword(ctx, actor.UserID, string(hash)); err != nil {
		if err == shared.ErrNotFound {
			return shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return err
	}
	return nil
}
