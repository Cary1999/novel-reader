package entity

import "time"

type AuthorApplication struct {
	ID                   int64      `json:"id"`
	UserID               int64      `json:"userId"`
	Username             string     `json:"username,omitempty"`
	Nickname             string     `json:"nickname,omitempty"`
	PenName              string     `json:"penName"`
	Reason               string     `json:"reason"`
	Status               string     `json:"status"`
	ReviewNote           string     `json:"reviewNote,omitempty"`
	ReviewedByOperatorID *int64     `json:"reviewedByOperatorId,omitempty"`
	ReviewedAt           *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}
