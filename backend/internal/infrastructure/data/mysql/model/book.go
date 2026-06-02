package model

import (
	"database/sql"
	"time"

	bookentity "novel-reader/backend/internal/domain/book/entity"
)

type Book struct {
	ID                 int64          `gorm:"column:id;primaryKey"`
	Title              string         `gorm:"column:title"`
	Author             string         `gorm:"column:author"`
	OwnerUserID        sql.NullInt64  `gorm:"column:owner_user_id"`
	CategoryID         sql.NullInt64  `gorm:"column:category_id"`
	SourceUploadID     sql.NullInt64  `gorm:"column:source_upload_id"`
	CategoryName       string         `gorm:"column:category_name;-:all"`
	Description        string         `gorm:"column:description"`
	ChapterCount       int            `gorm:"column:chapter_count"`
	LatestChapterTitle string         `gorm:"column:latest_chapter_title"`
	RecommendScore     int            `gorm:"column:recommend_score"`
	CoverPath          sql.NullString `gorm:"column:cover_path"`
	CreatedAt          time.Time      `gorm:"column:created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at"`
}

func (Book) TableName() string {
	return "books"
}

func (m *Book) FromEntity(e *bookentity.Book) {
	m.ID = e.ID
	m.Title = e.Title
	m.Author = e.Author
	if e.OwnerUserID != nil {
		m.OwnerUserID = sql.NullInt64{Int64: *e.OwnerUserID, Valid: true}
	}
	if e.CategoryID != nil {
		m.CategoryID = sql.NullInt64{Int64: *e.CategoryID, Valid: true}
	}
	m.CategoryName = e.Category
	m.Description = e.Description
	m.ChapterCount = e.ChapterCount
	m.LatestChapterTitle = e.LatestChapterTitle
	m.RecommendScore = e.RecommendScore
	if e.CoverPath != nil {
		m.CoverPath = sql.NullString{String: *e.CoverPath, Valid: true}
	}
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
}

func (m Book) ToEntity() bookentity.Book {
	item := bookentity.Book{
		ID:                 m.ID,
		Title:              m.Title,
		Author:             m.Author,
		Category:           m.CategoryName,
		Description:        m.Description,
		ChapterCount:       m.ChapterCount,
		LatestChapterTitle: m.LatestChapterTitle,
		RecommendScore:     m.RecommendScore,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
	if m.OwnerUserID.Valid {
		item.OwnerUserID = &m.OwnerUserID.Int64
	}
	if m.CategoryID.Valid {
		item.CategoryID = &m.CategoryID.Int64
	}
	if m.CoverPath.Valid {
		item.CoverPath = &m.CoverPath.String
	}
	return item
}
