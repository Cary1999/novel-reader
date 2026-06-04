package query

import (
	"context"
	"net/http"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	"novel-reader/backend/internal/domain/shared"
)

type CurrentOperatorHandler struct {
	operators identityrepository.OperatorRepository
}

func NewCurrentOperatorHandler(operators identityrepository.OperatorRepository) *CurrentOperatorHandler {
	return &CurrentOperatorHandler{operators: operators}
}

func (h *CurrentOperatorHandler) Handle(ctx context.Context, actor shared.Actor) (identityentity.Operator, error) {
	operator, err := h.operators.FindOperatorByID(ctx, actor.ActorID)
	if err != nil {
		if err == shared.ErrNotFound {
			return identityentity.Operator{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return identityentity.Operator{}, err
	}
	return operator, nil
}
