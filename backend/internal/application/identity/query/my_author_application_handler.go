package query

import (
	"context"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
	"novel-reader/backend/internal/domain/shared"
)

type MyAuthorApplicationHandler struct {
	applications identityrepository.AuthorApplicationRepository
}

func NewMyAuthorApplicationHandler(applications identityrepository.AuthorApplicationRepository) *MyAuthorApplicationHandler {
	return &MyAuthorApplicationHandler{applications: applications}
}

func (h *MyAuthorApplicationHandler) Handle(ctx context.Context, actor shared.Actor) (identityentity.AuthorApplication, error) {
	return h.applications.FindLatestAuthorApplicationByUserID(ctx, actor.ActorID)
}
