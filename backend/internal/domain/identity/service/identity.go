package service

import (
	"net/http"
	"regexp"
	"strings"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	"novel-reader/backend/internal/domain/shared"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`)

type IdentityService struct{}

func NewIdentityService() *IdentityService {
	return &IdentityService{}
}

func (s *IdentityService) NewRegisteredUser(username, passwordHash string, role shared.Role) identityentity.User {
	return identityentity.User{
		Username:     username,
		Nickname:     username,
		PasswordHash: passwordHash,
		Role:         role,
	}
}

func (s *IdentityService) NewOperator(username, passwordHash string, role shared.Role) identityentity.Operator {
	return identityentity.Operator{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
	}
}

func (s *IdentityService) NormalizeUsername(username string) (string, error) {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "username must be 3-32 letters, numbers, or underscores")
	}
	return username, nil
}

func (s *IdentityService) ValidatePasswordLength(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "password must be 8-128 characters")
	}
	return nil
}

func (s *IdentityService) NormalizeNickname(nickname string) (string, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "nickname is required")
	}
	if len([]rune(nickname)) > 64 {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "nickname is too long")
	}
	return nickname, nil
}

func (s *IdentityService) NormalizePenName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "pen name is required")
	}
	if len([]rune(value)) > 64 {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "pen name is too long")
	}
	return value, nil
}

func (s *IdentityService) NormalizeApplicationReason(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "reason is required")
	}
	if len([]rune(value)) > 500 {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "reason is too long")
	}
	return value, nil
}

func (s *IdentityService) NormalizeOperatorRole(role shared.Role) (shared.Role, error) {
	switch role {
	case shared.RoleReviewer, shared.RoleSuperAdmin:
		return role, nil
	default:
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid operator role")
	}
}

func (s *IdentityService) NormalizeApplicationDecision(decision string) (string, error) {
	decision = strings.TrimSpace(decision)
	switch decision {
	case "approved", "rejected":
		return decision, nil
	default:
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "decision must be approved or rejected")
	}
}
