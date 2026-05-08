package parser

import "testing"

func TestParseChaptersSupportsChineseArabicAndEnglishTitles(t *testing.T) {
	input := `序言会被忽略
第1章 初遇
这一段是第一章正文

第一章 风起
中文数字标题正文

Chapter 3: Return
English chapter content
`
	chapters, err := ParseChapters(input)
	if err != nil {
		t.Fatalf("ParseChapters returned error: %v", err)
	}
	if len(chapters) != 3 {
		t.Fatalf("expected 3 chapters, got %d", len(chapters))
	}
	if chapters[0].Title != "第1章 初遇" || chapters[1].Title != "第一章 风起" || chapters[2].Title != "Chapter 3: Return" {
		t.Fatalf("unexpected titles: %#v", chapters)
	}
	if chapters[0].Index != 1 || chapters[2].Index != 3 {
		t.Fatalf("unexpected indexes: %#v", chapters)
	}
	if chapters[2].Content != "English chapter content" {
		t.Fatalf("unexpected content: %q", chapters[2].Content)
	}
}

func TestParseChaptersRejectsTextWithoutChapterTitles(t *testing.T) {
	_, err := ParseChapters("just a plain txt file\nwithout a chapter marker")
	if err == nil {
		t.Fatal("expected error")
	}
}
