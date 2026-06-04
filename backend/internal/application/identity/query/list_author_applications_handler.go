package query

import (
	"context"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	identityrepository "novel-reader/backend/internal/domain/identity/repository"
)

type ListAuthorApplicationsHandler struct {
	applications identityrepository.AuthorApplicationRepository
}

func NewListAuthorApplicationsHandler(applications identityrepository.AuthorApplicationRepository) *ListAuthorApplicationsHandler {
	return &ListAuthorApplicationsHandler{applications: applications}
}

func (h *ListAuthorApplicationsHandler) Handle(ctx context.Context) ([]identityentity.AuthorApplication, error) {
	return h.applications.ListAuthorApplications(ctx)
}
