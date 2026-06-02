package model

import (
	"time"

	categoryentity "novel-reader/backend/internal/domain/category/entity"
)

type Category struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

func (m *Category) FromEntity(e *categoryentity.Category) {
	m.ID = e.ID
	m.Name = e.Name
	m.CreatedAt = e.CreatedAt
}

func (m Category) ToEntity() categoryentity.Category {
	return categoryentity.Category{
		ID:        m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
	}
}
