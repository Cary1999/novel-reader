package query

import (
	"context"

	"novel-reader/backend/internal/domain/shared"
	siteentity "novel-reader/backend/internal/domain/site/entity"
	siterepository "novel-reader/backend/internal/domain/site/repository"
	siteservice "novel-reader/backend/internal/domain/site/service"
)

type GetSettingsHandler struct {
	repo    siterepository.SiteRepository
	service *siteservice.SiteService
}

func NewGetSettingsHandler(repo siterepository.SiteRepository, service *siteservice.SiteService) *GetSettingsHandler {
	return &GetSettingsHandler{repo: repo, service: service}
}

func (h *GetSettingsHandler) Handle(ctx context.Context) (siteentity.Settings, error) {
	item, err := h.repo.GetSettings(ctx)
	if err != nil {
		if err == shared.ErrNotFound {
			return h.service.DefaultSettings(), nil
		}
		return siteentity.Settings{}, err
	}
	return h.service.ApplyDefaults(item), nil
}
