package command

import (
	"context"

	"novel-reader/backend/internal/domain/shared"
	siteentity "novel-reader/backend/internal/domain/site/entity"
	siterepository "novel-reader/backend/internal/domain/site/repository"
	siteservice "novel-reader/backend/internal/domain/site/service"
)

type UpdateSettings struct {
	BrandName       string `json:"brandName"`
	BrandSubtitle   string `json:"brandSubtitle"`
	HeroEyebrow     string `json:"heroEyebrow"`
	HeroTitle       string `json:"heroTitle"`
	HeroDescription string `json:"heroDescription"`
}

type UpdateSettingsHandler struct {
	repo    siterepository.SiteRepository
	service *siteservice.SiteService
}

func NewUpdateSettingsHandler(repo siterepository.SiteRepository, service *siteservice.SiteService) *UpdateSettingsHandler {
	return &UpdateSettingsHandler{repo: repo, service: service}
}

func (h *UpdateSettingsHandler) Handle(ctx context.Context, actor shared.Actor, input UpdateSettings) (siteentity.Settings, error) {
	current, err := h.repo.GetSettings(ctx)
	if err != nil && err != shared.ErrNotFound {
		return siteentity.Settings{}, err
	}
	if err == shared.ErrNotFound {
		current = h.service.DefaultSettings()
	} else {
		current = h.service.ApplyDefaults(current)
	}

	item, err := h.service.BuildSettings(siteservice.SettingsInput{
		BrandName:       input.BrandName,
		BrandSubtitle:   input.BrandSubtitle,
		HeroEyebrow:     input.HeroEyebrow,
		HeroTitle:       input.HeroTitle,
		HeroDescription: input.HeroDescription,
	}, current.BrandIconPath, actor.ActorID)
	if err != nil {
		return siteentity.Settings{}, err
	}
	return h.repo.UpsertSettings(ctx, item)
}
