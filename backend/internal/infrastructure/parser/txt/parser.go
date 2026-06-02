package txt

import (
	"errors"
	"regexp"
	"strings"

	bookentity "novel-reader/backend/internal/domain/book/entity"
)

var chapterTitlePattern = regexp.MustCompile(`(?i)^\s*((第\s*[0-9０-９一二三四五六七八九十百千万两]+\s*[章节回卷部].*)|(chapter\s+[0-9]+[\s:：.-]?.*))\s*$`)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (Parser) ParseChapters(text string) ([]bookentity.ChapterDraft, error) {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")

	var chapters []bookentity.ChapterDraft
	current := -1
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isChapterTitle(trimmed) {
			chapters = append(chapters, bookentity.ChapterDraft{
				Index: len(chapters) + 1,
				Title: trimmed,
			})
			current = len(chapters) - 1
			continue
		}
		if current >= 0 {
			chapters[current].Content += line + "\n"
		}
	}

	for i := range chapters {
		chapters[i].Content = strings.TrimSpace(chapters[i].Content)
	}
	if len(chapters) == 0 {
		return nil, errors.New("no chapters found")
	}
	return chapters, nil
}

func isChapterTitle(line string) bool {
	if line == "" || len([]rune(line)) > 80 {
		return false
	}
	return chapterTitlePattern.MatchString(line)
}
