package model

import (
	"database/sql"
	"time"

	siteentity "novel-reader/backend/internal/domain/site/entity"
)

type SiteSettings struct {
	ID                  int64          `gorm:"column:id;primaryKey"`
	BrandName           string         `gorm:"column:brand_name"`
	BrandSubtitle       string         `gorm:"column:brand_subtitle"`
	BrandIconPath       sql.NullString `gorm:"column:brand_icon_path"`
	HeroEyebrow         string         `gorm:"column:hero_eyebrow"`
	HeroTitle           string         `gorm:"column:hero_title"`
	HeroDescription     string         `gorm:"column:hero_description"`
	UpdatedByOperatorID sql.NullInt64  `gorm:"column:updated_by_operator_id"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
}

func (SiteSettings) TableName() string {
	return "site_settings"
}

func (m *SiteSettings) FromEntity(e *siteentity.Settings) {
	m.ID = e.ID
	m.BrandName = e.BrandName
	m.BrandSubtitle = e.BrandSubtitle
	if e.BrandIconPath != nil {
		m.BrandIconPath = sql.NullString{String: *e.BrandIconPath, Valid: true}
	}
	m.HeroEyebrow = e.HeroEyebrow
	m.HeroTitle = e.HeroTitle
	m.HeroDescription = e.HeroDescription
	if e.UpdatedByOperatorID != nil {
		m.UpdatedByOperatorID = sql.NullInt64{Int64: *e.UpdatedByOperatorID, Valid: true}
	}
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
}

func (m SiteSettings) ToEntity() siteentity.Settings {
	item := siteentity.Settings{
		ID:              m.ID,
		BrandName:       m.BrandName,
		BrandSubtitle:   m.BrandSubtitle,
		HeroEyebrow:     m.HeroEyebrow,
		HeroTitle:       m.HeroTitle,
		HeroDescription: m.HeroDescription,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
	if m.BrandIconPath.Valid {
		item.BrandIconPath = &m.BrandIconPath.String
	}
	if m.UpdatedByOperatorID.Valid {
		item.UpdatedByOperatorID = &m.UpdatedByOperatorID.Int64
	}
	return item
}
