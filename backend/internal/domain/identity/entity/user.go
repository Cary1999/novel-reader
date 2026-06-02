package entity

import (
	"time"

	"novel-reader/backend/internal/domain/shared"
)

type User struct {
	ID           int64       `json:"id"`
	Username     string      `json:"username"`
	Nickname     string      `json:"nickname"`
	PasswordHash string      `json:"-"`
	Role         shared.Role `json:"role"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}
