package command

import (
	"context"
	"net/http"
	"strings"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	identityservice "novel-reader/backend/internal/domain/identity/service"
	"novel-reader/backend/internal/domain/shared"
)

type ReviewAuthorApplication struct {
	Decision   string `json:"decision"`
	ReviewNote string `json:"reviewNote"`
}

type ReviewAuthorApplicationHandler struct {
	users        identityrepository.UserRepository
	applications identityrepository.AuthorApplicationRepository
	service      *identityservice.IdentityService
}

func NewReviewAuthorApplicationHandler(users identityrepository.UserRepository, applications identityrepository.AuthorApplicationRepository, service *identityservice.IdentityService) *ReviewAuthorApplicationHandler {
	return &ReviewAuthorApplicationHandler{users: users, applications: applications, service: service}
}

func (h *ReviewAuthorApplicationHandler) Handle(ctx context.Context, actor shared.Actor, applicationID int64, input ReviewAuthorApplication) (identityentity.AuthorApplication, error) {
	if !actor.IsReviewer() {
		return identityentity.AuthorApplication{}, shared.NewError(http.StatusForbidden, "FORBIDDEN", "reviewer required")
	}
	decision, err := h.service.NormalizeApplicationDecision(input.Decision)
	if err != nil {
		return identityentity.AuthorApplication{}, err
	}
	item, err := h.applications.ReviewAuthorApplication(ctx, applicationID, decision, strings.TrimSpace(input.ReviewNote), actor.ActorID)
	if err != nil {
		if err == shared.ErrNotFound {
			return identityentity.AuthorApplication{}, shared.NewError(http.StatusNotFound, "NOT_FOUND", "author application not found")
		}
		return identityentity.AuthorApplication{}, err
	}
	if decision == "approved" {
		if _, err := h.users.PromoteUserToAuthor(ctx, item.UserID); err != nil {
			return identityentity.AuthorApplication{}, err
		}
	}
	return item, nil
}
