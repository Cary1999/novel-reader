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

type OperatorTokenIssuer interface {
	IssueOperator(operator identityentity.Operator) (string, error)
}

type AdminLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminLoginResult struct {
	Token    string
	Operator identityentity.Operator
}

type AdminLoginHandler struct {
	operators identityrepository.OperatorRepository
	tokens    OperatorTokenIssuer
}

func NewAdminLoginHandler(operators identityrepository.OperatorRepository, tokens OperatorTokenIssuer) *AdminLoginHandler {
	return &AdminLoginHandler{operators: operators, tokens: tokens}
}

func (h *AdminLoginHandler) Handle(ctx context.Context, input AdminLogin) (AdminLoginResult, error) {
	operator, err := h.operators.FindOperatorByUsername(ctx, strings.TrimSpace(input.Username))
	if err != nil {
		if err == shared.ErrNotFound {
			return AdminLoginResult{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
		}
		return AdminLoginResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(operator.PasswordHash), []byte(input.Password)) != nil {
		return AdminLoginResult{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
	}
	token, err := h.tokens.IssueOperator(operator)
	if err != nil {
		return AdminLoginResult{}, err
	}
	return AdminLoginResult{Token: token, Operator: operator}, nil
}
