package ebooks

import "testing"

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
