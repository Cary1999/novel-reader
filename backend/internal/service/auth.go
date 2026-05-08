package service

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"novel-reader/backend/internal/auth"
	"novel-reader/backend/internal/domain"
	"novel-reader/backend/internal/repository"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`)

type AuthService struct {
	store  repository.Store
	tokens *auth.Manager
}

func NewAuthService(store repository.Store, tokens *auth.Manager) *AuthService {
	return &AuthService{store: store, tokens: tokens}
}

func (s *AuthService) Register(ctx context.Context, username, password string) (domain.User, error) {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return domain.User{}, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "username must be 3-32 letters, numbers, or underscores")
	}
	if len(password) < 8 || len(password) > 128 {
		return domain.User{}, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "password must be 8-128 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	user, err := s.store.CreateUser(ctx, username, string(hash), domain.RoleUser)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return domain.User{}, domain.NewError(http.StatusConflict, "CONFLICT", "username already exists")
		}
		return domain.User{}, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, domain.User, error) {
	user, err := s.store.FindUserByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if repository.IsNotFound(err) {
			return "", domain.User{}, domain.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
		}
		return "", domain.User{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", domain.User{}, domain.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
	}
	token, err := s.tokens.Issue(user)
	if err != nil {
		return "", domain.User{}, err
	}
	return token, user, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, claims auth.Claims) (domain.User, error) {
	user, err := s.store.FindUserByID(ctx, claims.UserID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.User{}, domain.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return domain.User{}, err
	}
	return user, nil
}

func (s *AuthService) UpdateNickname(ctx context.Context, claims auth.Claims, nickname string) (domain.User, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return domain.User{}, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "nickname is required")
	}
	if len([]rune(nickname)) > 64 {
		return domain.User{}, domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "nickname is too long")
	}
	user, err := s.store.UpdateUserNickname(ctx, claims.UserID, nickname)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.User{}, domain.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return domain.User{}, err
	}
	return user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, claims auth.Claims, oldPassword, newPassword string) error {
	if len(newPassword) < 8 || len(newPassword) > 128 {
		return domain.NewError(http.StatusBadRequest, "BAD_REQUEST", "password must be 8-128 characters")
	}
	user, err := s.store.FindUserByID(ctx, claims.UserID)
	if err != nil {
		if repository.IsNotFound(err) {
			return domain.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)) != nil {
		return domain.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.store.UpdateUserPassword(ctx, claims.UserID, string(hash)); err != nil {
		if repository.IsNotFound(err) {
			return domain.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return err
	}
	return nil
}
