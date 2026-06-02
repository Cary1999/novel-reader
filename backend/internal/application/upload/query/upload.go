package query

type ImportResult struct {
	BookID            int64  `json:"bookId"`
	ChapterCount      int    `json:"chapterCount"`
	UploadID          int64  `json:"uploadId"`
	FirstChapterTitle string `json:"firstChapterTitle"`
	LastChapterTitle  string `json:"lastChapterTitle"`
}
