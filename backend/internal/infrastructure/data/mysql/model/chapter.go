package model

import bookentity "novel-reader/backend/internal/domain/book/entity"

type Chapter struct {
	ID      int64  `gorm:"column:id;primaryKey"`
	BookID  int64  `gorm:"column:book_id"`
	Index   int    `gorm:"column:chapter_index"`
	Title   string `gorm:"column:title"`
	Content string `gorm:"column:content"`
}

func (Chapter) TableName() string {
	return "chapters"
}

func (m *Chapter) FromEntity(e *bookentity.Chapter) {
	m.ID = e.ID
	m.BookID = e.BookID
	m.Index = e.Index
	m.Title = e.Title
	m.Content = e.Content
}

func (m Chapter) ToEntity() bookentity.Chapter {
	return bookentity.Chapter{
		ID:      m.ID,
		BookID:  m.BookID,
		Index:   m.Index,
		Title:   m.Title,
		Content: m.Content,
	}
}
