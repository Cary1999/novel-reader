package repo

import (
	"context"
	"database/sql"

	siteentity "novel-reader/backend/internal/domain/site/entity"
	"novel-reader/backend/internal/infrastructure/data/mysql/model"
)

type SiteRepository struct {
	db *sql.DB
}

func NewSiteRepository(db *sql.DB) SiteRepository {
	return SiteRepository{db: db}
}

func (r SiteRepository) GetSettings(ctx context.Context) (siteentity.Settings, error) {
	var record model.SiteSettings
	err := r.db.QueryRowContext(ctx, `
		SELECT id, brand_name, brand_subtitle, brand_icon_path,
		       hero_eyebrow, hero_title, hero_description, updated_by_user_id, created_at, updated_at
		FROM site_settings WHERE id = 1
	`).Scan(
		&record.ID,
		&record.BrandName,
		&record.BrandSubtitle,
		&record.BrandIconPath,
		&record.HeroEyebrow,
		&record.HeroTitle,
		&record.HeroDescription,
		&record.UpdatedByUserID,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return siteentity.Settings{}, mapNotFound(err)
	}
	return record.ToEntity(), nil
}

func (r SiteRepository) UpsertSettings(ctx context.Context, item siteentity.Settings) (siteentity.Settings, error) {
	var record model.SiteSettings
	record.FromEntity(&item)
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO site_settings (
			id, brand_name, brand_subtitle, brand_icon_path,
			hero_eyebrow, hero_title, hero_description, updated_by_user_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			brand_name = VALUES(brand_name),
			brand_subtitle = VALUES(brand_subtitle),
			brand_icon_path = VALUES(brand_icon_path),
			hero_eyebrow = VALUES(hero_eyebrow),
			hero_title = VALUES(hero_title),
			hero_description = VALUES(hero_description),
			updated_by_user_id = VALUES(updated_by_user_id)
	`, 1, record.BrandName, record.BrandSubtitle, nullableString(record.BrandIconPath), record.HeroEyebrow, record.HeroTitle, record.HeroDescription, nullableInt64(record.UpdatedByUserID)); err != nil {
		return siteentity.Settings{}, err
	}
	return r.GetSettings(ctx)
}

func nullableString(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func nullableInt64(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}
