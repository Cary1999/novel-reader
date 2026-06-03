package command

import (
	"context"
	"io"
	"net/http"
	"strings"

	"novel-reader/backend/internal/domain/shared"
	siteentity "novel-reader/backend/internal/domain/site/entity"
	siterepository "novel-reader/backend/internal/domain/site/repository"
	siteservice "novel-reader/backend/internal/domain/site/service"
)

type IconStore interface {
	SaveSiteIcon(originalName string, reader io.Reader) (SavedIcon, error)
}

type SavedIcon struct {
	RelativePath string
	Size         int64
}

type UpdateIconHandler struct {
	repo      siterepository.SiteRepository
	service   *siteservice.SiteService
	iconStore IconStore
}

func NewUpdateIconHandler(repo siterepository.SiteRepository, service *siteservice.SiteService, iconStore IconStore) *UpdateIconHandler {
	return &UpdateIconHandler{repo: repo, service: service, iconStore: iconStore}
}

func (h *UpdateIconHandler) Handle(ctx context.Context, actor shared.Actor, originalName string, reader io.Reader) (siteentity.Settings, error) {
	saved, err := h.iconStore.SaveSiteIcon(originalName, reader)
	if err != nil {
		return siteentity.Settings{}, uploadSiteIconError(err)
	}
	current, err := h.repo.GetSettings(ctx)
	if err != nil && err != shared.ErrNotFound {
		return siteentity.Settings{}, err
	}
	if err == shared.ErrNotFound {
		current = h.service.DefaultSettings()
	}
	item := h.service.AttachIcon(current, saved.RelativePath, actor.UserID)
	return h.repo.UpsertSettings(ctx, item)
}

func uploadSiteIconError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "invalid file type"):
		return shared.NewError(http.StatusBadRequest, "UPLOAD_INVALID_TYPE", "only jpg/jpeg/png/webp/svg files are allowed")
	case strings.Contains(message, "file too large"):
		return shared.NewError(http.StatusRequestEntityTooLarge, "UPLOAD_TOO_LARGE", "site icon is too large")
	case strings.Contains(message, "empty file"):
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "site icon is empty")
	default:
		return shared.NewError(http.StatusBadRequest, "BAD_REQUEST", "invalid upload")
	}
}
