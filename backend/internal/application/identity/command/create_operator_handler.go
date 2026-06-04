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

type CreateOperator struct {
	Username string      `json:"username"`
	Password string      `json:"password"`
	Role     shared.Role `json:"role"`
}

type CreateOperatorHandler struct {
	operators identityrepository.OperatorRepository
	service   *identityservice.IdentityService
}

func NewCreateOperatorHandler(operators identityrepository.OperatorRepository, service *identityservice.IdentityService) *CreateOperatorHandler {
	return &CreateOperatorHandler{operators: operators, service: service}
}

func (h *CreateOperatorHandler) Handle(ctx context.Context, actor shared.Actor, input CreateOperator) (identityentity.Operator, error) {
	if !actor.IsSuperAdmin() {
		return identityentity.Operator{}, shared.NewError(http.StatusForbidden, "FORBIDDEN", "super admin required")
	}
	username, err := h.service.NormalizeUsername(input.Username)
	if err != nil {
		return identityentity.Operator{}, err
	}
	if err := h.service.ValidatePasswordLength(input.Password); err != nil {
		return identityentity.Operator{}, err
	}
	role, err := h.service.NormalizeOperatorRole(input.Role)
	if err != nil {
		return identityentity.Operator{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return identityentity.Operator{}, err
	}
	item, err := h.operators.CreateOperator(ctx, username, string(hash), role)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return identityentity.Operator{}, shared.NewError(http.StatusConflict, "CONFLICT", "username already exists")
		}
		return identityentity.Operator{}, err
	}
	return item, nil
}
