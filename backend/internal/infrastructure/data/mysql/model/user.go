package model

import (
	"database/sql"
	"time"

	identityentity "novel-reader/backend/internal/domain/identity/entity"
	"novel-reader/backend/internal/domain/shared"
)

type User struct {
	ID           int64
	Username     string
	Nickname     string
	AvatarPath   sql.NullString
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Operator struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AuthorApplication struct {
	ID                   int64
	UserID               int64
	Username             string
	Nickname             string
	PenName              string
	Reason               string
	Status               string
	ReviewNote           string
	ReviewedByOperatorID *int64
	ReviewedAt           *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type FrontUserSummary struct {
	ID                  int64
	Username            string
	Nickname            string
	Role                string
	LatestApplicationID *int64
	LatestApplication   string
	LatestApplicationAt *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (m *User) FromEntity(e *identityentity.User) {
	m.ID = e.ID
	m.Username = e.Username
	m.Nickname = e.Nickname
	if e.AvatarPath != nil {
		m.AvatarPath = sql.NullString{String: *e.AvatarPath, Valid: true}
	} else {
		m.AvatarPath = sql.NullString{}
	}
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
		AvatarPath:   nullStringToPointer(m.AvatarPath),
		PasswordHash: m.PasswordHash,
		Role:         shared.Role(m.Role),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func nullStringToPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func (m Operator) ToEntity() identityentity.Operator {
	return identityentity.Operator{
		ID:           m.ID,
		Username:     m.Username,
		PasswordHash: m.PasswordHash,
		Role:         shared.Role(m.Role),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (m AuthorApplication) ToEntity() identityentity.AuthorApplication {
	item := identityentity.AuthorApplication{
		ID:         m.ID,
		UserID:     m.UserID,
		Username:   m.Username,
		Nickname:   m.Nickname,
		PenName:    m.PenName,
		Reason:     m.Reason,
		Status:     m.Status,
		ReviewNote: m.ReviewNote,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
	if m.ReviewedByOperatorID != nil {
		item.ReviewedByOperatorID = m.ReviewedByOperatorID
	}
	if m.ReviewedAt != nil {
		item.ReviewedAt = m.ReviewedAt
	}
	return item
}

func (m FrontUserSummary) ToEntity() identityentity.FrontUserSummary {
	return identityentity.FrontUserSummary{
		ID:                  m.ID,
		Username:            m.Username,
		Nickname:            m.Nickname,
		Role:                m.Role,
		LatestApplicationID: m.LatestApplicationID,
		LatestApplication:   m.LatestApplication,
		LatestApplicationAt: m.LatestApplicationAt,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}
