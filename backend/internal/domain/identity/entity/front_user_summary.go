package entity

import "time"

type FrontUserSummary struct {
	ID                  int64      `json:"id"`
	Username            string     `json:"username"`
	Nickname            string     `json:"nickname"`
	Role                string     `json:"role"`
	LatestApplicationID *int64     `json:"latestApplicationId,omitempty"`
	LatestApplication   string     `json:"latestApplicationStatus,omitempty"`
	LatestApplicationAt *time.Time `json:"latestApplicationCreatedAt,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}
