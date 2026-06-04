package command

import (
	"context"
	"net/http"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	identityservice "novel-reader/backend/internal/domain/identity/service"
	"novel-reader/backend/internal/domain/shared"
)

type UpdateOperator struct {
	Role shared.Role `json:"role"`
}

type UpdateOperatorHandler struct {
	operators identityrepository.OperatorRepository
	service   *identityservice.IdentityService
}

func NewUpdateOperatorHandler(operators identityrepository.OperatorRepository, service *identityservice.IdentityService) *UpdateOperatorHandler {
	return &UpdateOperatorHandler{operators: operators, service: service}
}

func (h *UpdateOperatorHandler) Handle(ctx context.Context, actor shared.Actor, operatorID int64, input UpdateOperator) (identityentity.Operator, error) {
	if !actor.IsSuperAdmin() {
		return identityentity.Operator{}, shared.NewError(http.StatusForbidden, "FORBIDDEN", "super admin required")
	}
	role, err := h.service.NormalizeOperatorRole(input.Role)
	if err != nil {
		return identityentity.Operator{}, err
	}
	item, err := h.operators.UpdateOperatorRole(ctx, operatorID, role)
	if err != nil {
		if err == shared.ErrNotFound {
			return identityentity.Operator{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "operator not found")
		}
		return identityentity.Operator{}, err
	}
	return item, nil
}
