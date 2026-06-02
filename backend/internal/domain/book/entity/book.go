package entity

import "time"

type Book struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	Author             string    `json:"author"`
	OwnerUserID        *int64    `json:"-"`
	CategoryID         *int64    `json:"categoryId,omitempty"`
	Category           string    `json:"category"`
	Description        string    `json:"description"`
	ChapterCount       int       `json:"chapterCount"`
	LatestChapterTitle string    `json:"latestChapterTitle"`
	RecommendScore     int       `json:"recommendScore"`
	CoverURL           string    `json:"coverUrl"`
	CoverPath          *string   `json:"-"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"-"`
}
