package service

import (
	"net/http"
	"strings"

	"novel-reader/backend/internal/domain/shared"
	siteentity "novel-reader/backend/internal/domain/site/entity"
)

const (
	defaultBrandName       = "阅卷书屋"
	defaultBrandSubtitle   = "Local Reading Archive"
	defaultHeroEyebrow     = "发现好故事"
	defaultHeroTitle       = "一站式书屋"
	defaultHeroDescription = "搜索、筛选、查看目录、登录后按章节阅读，也可以上传和维护你自己的 txt 小说。"
)

type SettingsInput struct {
	BrandName       string
	BrandSubtitle   string
	HeroEyebrow     string
	HeroTitle       string
	HeroDescription string
}

type SiteService struct{}

func NewSiteService() *SiteService {
	return &SiteService{}
}

func (s *SiteService) DefaultSettings() siteentity.Settings {
	return siteentity.Settings{
		ID:              1,
		BrandName:       defaultBrandName,
		BrandSubtitle:   defaultBrandSubtitle,
		HeroEyebrow:     defaultHeroEyebrow,
		HeroTitle:       defaultHeroTitle,
		HeroDescription: defaultHeroDescription,
	}
}

func (s *SiteService) ApplyDefaults(item siteentity.Settings) siteentity.Settings {
	defaults := s.DefaultSettings()
	if strings.TrimSpace(item.BrandName) == "" {
		item.BrandName = defaults.BrandName
	}
	if strings.TrimSpace(item.BrandSubtitle) == "" {
		item.BrandSubtitle = defaults.BrandSubtitle
	}
	if strings.TrimSpace(item.HeroEyebrow) == "" {
		item.HeroEyebrow = defaults.HeroEyebrow
	}
	if strings.TrimSpace(item.HeroTitle) == "" {
		item.HeroTitle = defaults.HeroTitle
	}
	if strings.TrimSpace(item.HeroDescription) == "" {
		item.HeroDescription = defaults.HeroDescription
	}
	return item
}

func (s *SiteService) BuildSettings(input SettingsInput, iconPath *string, updatedBy int64) (siteentity.Settings, error) {
	brandName, err := requireField(input.BrandName, "brand name", 120)
	if err != nil {
		return siteentity.Settings{}, err
	}
	brandSubtitle, err := requireField(input.BrandSubtitle, "brand subtitle", 255)
	if err != nil {
		return siteentity.Settings{}, err
	}
	heroEyebrow, err := requireField(input.HeroEyebrow, "hero eyebrow", 120)
	if err != nil {
		return siteentity.Settings{}, err
	}
	heroTitle, err := requireField(input.HeroTitle, "hero title", 255)
	if err != nil {
		return siteentity.Settings{}, err
	}
	heroDescription, err := requireField(input.HeroDescription, "hero description", 500)
	if err != nil {
		return siteentity.Settings{}, err
	}
	return siteentity.Settings{
		ID:              1,
		BrandName:       brandName,
		BrandSubtitle:   brandSubtitle,
		BrandIconPath:   iconPath,
		HeroEyebrow:     heroEyebrow,
		HeroTitle:       heroTitle,
		HeroDescription: heroDescription,
		UpdatedByUserID: pointer(updatedBy),
	}, nil
}

func (s *SiteService) AttachIcon(item siteentity.Settings, iconPath string, updatedBy int64) siteentity.Settings {
	item = s.ApplyDefaults(item)
	item.ID = 1
	item.BrandIconPath = pointer(iconPath)
	item.UpdatedByUserID = pointer(updatedBy)
	return item
}

func requireField(value, field string, max int) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", field+" is required")
	}
	if len([]rune(trimmed)) > max {
		return "", shared.NewError(http.StatusBadRequest, "BAD_REQUEST", field+" is too long")
	}
	return trimmed, nil
}

func pointer[T any](value T) *T {
	return &value
}
