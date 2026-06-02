package model

import (
	"time"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	"novel-reader/backend/internal/domain/shared"
)

type User struct {
	ID           int64
	Username     string
	Nickname     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (m *User) FromEntity(e *identityentity.User) {
	m.ID = e.ID
	m.Username = e.Username
	m.Nickname = e.Nickname
	m.PasswordHash = e.PasswordHash
	m.Role = string(e.Role)
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
}

func (m User) ToEntity() identityentity.User {
	return identityentity.User{
		ID:           m.ID,
		Username:     m.Username,
		Nickname:     m.Nickname,
		PasswordHash: m.PasswordHash,
		Role:         shared.Role(m.Role),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
