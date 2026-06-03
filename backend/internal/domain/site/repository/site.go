package repository

import (
	"context"

	siteentity "novel-reader/backend/internal/domain/site/entity"
)

type SiteRepository interface {
	GetSettings(ctx context.Context) (siteentity.Settings, error)
	UpsertSettings(ctx context.Context, item siteentity.Settings) (siteentity.Settings, error)
}
