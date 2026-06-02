package entity

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
