package ebooks

import (
	"strconv"
	"testing"
)

func TestNormalizeLimit(t *testing.T) {
	v, err := normalizeLimit("", 20, 1, 50, true)
	if err != nil || v != 20 {
		t.Fatalf("expected 20,nil got %d,%v", v, err)
	}
	v, err = normalizeLimit("5", 20, 1, 50, true)
	if err != nil || v != 5 {
		t.Fatalf("expected 5,nil got %d,%v", v, err)
	}
	_, err = normalizeLimit("0", 20, 1, 50, true)
	if err == nil {
		t.Fatal("expected error for low bound")
	}
	v, err = normalizeLimit("100", 20, 1, 50, true)
	if err != nil || v != 50 {
		t.Fatalf("expected clamped 50,nil got %d,%v", v, err)
	}
	_, err = normalizeLimit("abc", 20, 1, 50, true)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestDetectDownloadRoute(t *testing.T) {
	cases := []struct {
		arg1 string
		arg2 string
		want DownloadRoute
	}{
		{"12345", "abcdef", RouteZlibIDHash},
		{"https://demo.local/opds/download/1/epub/", "", RouteCalibreURL},
		{"https://archive.org/download/id/file.pdf", "", RouteArchiveURL},
		{"L0123456789abcdef0123456789abcdef", "", RouteLiber3ID},
		{"A0123456789abcdef0123456789abcdef", "", RouteAnnasID},
		{"bad", "", RouteUnknown},
	}
	for _, c := range cases {
		got := detectDownloadRoute(c.arg1, c.arg2)
		if got != c.want {
			t.Fatalf("route(%q,%q)=%v want %v", c.arg1, c.arg2, got, c.want)
		}
	}
}

func TestBuildSearchBlockTextsPagination(t *testing.T) {
	items := make([]BookItem, 0, 8)
	for i := 0; i < 8; i++ {
		items = append(items, BookItem{
			Source:      "archive.org",
			Title:       "A Very Long Book Title " + string(rune('A'+i)),
			Authors:     "Author Name",
			Format:      "epub",
			ID:          "book-id",
			DownloadURL: "https://archive.org/download/id/file.epub",
		})
	}
	blocks := buildSearchBlockTexts("archive.org", items, nil, 220)
	if len(blocks) < 2 {
		t.Fatalf("expected pagination, got %d block(s)", len(blocks))
	}
	for _, block := range blocks {
		if len(block) > 260 {
			t.Fatalf("block too long: %d", len(block))
		}
	}
}

func TestBuildSearchBlockTextsErrorAndEmpty(t *testing.T) {
	errBlocks := buildSearchBlockTexts("Liber3", nil, strconv.ErrSyntax, 3000)
	if len(errBlocks) != 1 || errBlocks[0] == "" {
		t.Fatal("expected one error block")
	}
	emptyBlocks := buildSearchBlockTexts("Calibre-Web", nil, nil, 3000)
	if len(emptyBlocks) != 1 || emptyBlocks[0] == "" {
		t.Fatal("expected one empty block")
	}
}
