package command

import (
	"context"
	"net/http"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	identityservice "novel-reader/backend/internal/domain/identity/service"
	"novel-reader/backend/internal/domain/shared"
)

type SubmitAuthorApplication struct {
	PenName string `json:"penName"`
	Reason  string `json:"reason"`
}

type SubmitAuthorApplicationHandler struct {
	users        identityrepository.UserRepository
	applications identityrepository.AuthorApplicationRepository
	service      *identityservice.IdentityService
}

func NewSubmitAuthorApplicationHandler(users identityrepository.UserRepository, applications identityrepository.AuthorApplicationRepository, service *identityservice.IdentityService) *SubmitAuthorApplicationHandler {
	return &SubmitAuthorApplicationHandler{users: users, applications: applications, service: service}
}

func (h *SubmitAuthorApplicationHandler) Handle(ctx context.Context, actor shared.Actor, input SubmitAuthorApplication) (identityentity.AuthorApplication, error) {
	if !actor.IsFront() {
		return identityentity.AuthorApplication{}, shared.NewError(http.StatusForbidden, "FORBIDDEN", "front account required")
	}
	user, err := h.users.FindUserByID(ctx, actor.ActorID)
	if err != nil {
		if err == shared.ErrNotFound {
			return identityentity.AuthorApplication{}, shared.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "login required")
		}
		return identityentity.AuthorApplication{}, err
	}
	if user.Role != shared.RoleReader {
		return identityentity.AuthorApplication{}, shared.NewError(http.StatusForbidden, "FORBIDDEN", "only readers can submit author applications")
	}
	if existing, err := h.applications.FindLatestAuthorApplicationByUserID(ctx, actor.ActorID); err == nil && existing.Status == "pending" {
		return identityentity.AuthorApplication{}, shared.NewError(http.StatusConflict, "CONFLICT", "an author application is already pending")
	} else if err != nil && err != shared.ErrNotFound {
		return identityentity.AuthorApplication{}, err
	}
	penName, err := h.service.NormalizePenName(input.PenName)
	if err != nil {
		return identityentity.AuthorApplication{}, err
	}
	reason, err := h.service.NormalizeApplicationReason(input.Reason)
	if err != nil {
		return identityentity.AuthorApplication{}, err
	}
	return h.applications.CreateAuthorApplication(ctx, identityentity.AuthorApplication{
		UserID:  actor.ActorID,
		PenName: penName,
		Reason:  reason,
		Status:  "pending",
	})
}
