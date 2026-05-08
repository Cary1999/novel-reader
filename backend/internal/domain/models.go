package domain

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Nickname     string    `json:"nickname"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

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
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"-"`
}

type Chapter struct {
	ID      int64  `json:"id"`
	BookID  int64  `json:"bookId"`
	Index   int    `json:"index"`
	Title   string `json:"title"`
	Content string `json:"content,omitempty"`
}

type ChapterDraft struct {
	Index   int
	Title   string
	Content string
}

type Upload struct {
	ID               int64
	AdminUserID      int64
	OriginalFilename string
	StoredPath       string
	FileSize         int64
	Status           string
	ErrorMessage     string
	CreatedAt        time.Time
}

type UploadBookInput struct {
	Title       string
	CategoryID  int64
	Description string
	OwnerUserID int64
}

type BookMetadataInput struct {
	Title       string `json:"title"`
	CategoryID  int64  `json:"categoryId"`
	Description string `json:"description"`
}

type ChapterInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UploadResult struct {
	BookID            int64  `json:"bookId"`
	ChapterCount      int    `json:"chapterCount"`
	UploadID          int64  `json:"uploadId"`
	FirstChapterTitle string `json:"firstChapterTitle"`
	LastChapterTitle  string `json:"lastChapterTitle"`
}
