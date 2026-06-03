package entity

import "time"

type Settings struct {
	ID              int64     `json:"id"`
	BrandName       string    `json:"brandName"`
	BrandSubtitle   string    `json:"brandSubtitle"`
	BrandIconURL    string    `json:"brandIconUrl"`
	BrandIconPath   *string   `json:"-"`
	HeroEyebrow     string    `json:"heroEyebrow"`
	HeroTitle       string    `json:"heroTitle"`
	HeroDescription string    `json:"heroDescription"`
	UpdatedByUserID *int64    `json:"-"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
