package query

import (
	"context"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
)

type ListOperatorsHandler struct {
	operators identityrepository.OperatorRepository
}

func NewListOperatorsHandler(operators identityrepository.OperatorRepository) *ListOperatorsHandler {
	return &ListOperatorsHandler{operators: operators}
}

func (h *ListOperatorsHandler) Handle(ctx context.Context) ([]identityentity.Operator, error) {
	return h.operators.ListOperators(ctx)
}
