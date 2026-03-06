package ebooks

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	reArchiveURL = regexp.MustCompile(`^https://archive\.org/download/[^/]+/[^/]+$`)
	reLiber3ID   = regexp.MustCompile(`^L[a-fA-F0-9]{32}$`)
	reAnnasID    = regexp.MustCompile(`^A[a-fA-F0-9]{32}$`)
	reZlibHash   = regexp.MustCompile(`^[a-fA-F0-9]{6}$`)
)

func defaultConfig() Config {
	return Config{
		EnableCalibre: false,
		EnableLiber3:  true,
		EnableArchive: true,
		EnableZlib:    false,
		EnableAnnas:   false,
		MaxResults:    20,
		DownloadMode:  "both",
	}
}

func sanitizeConfig(c *Config) {
	if c.MaxResults < 1 || c.MaxResults > 50 {
		c.MaxResults = 20
	}
	switch c.DownloadMode {
	case "link", "upload", "both":
	default:
		c.DownloadMode = "both"
	}
	c.CalibreWebURL = strings.TrimSpace(c.CalibreWebURL)
	if c.CalibreWebURL == "" {
		c.EnableCalibre = false
	}
}

func normalizeLimit(in string, fallback int, min, max int, clampMax bool) (int, error) {
	v := fallback
	if strings.TrimSpace(in) != "" {
		i, err := strconv.Atoi(strings.TrimSpace(in))
		if err != nil {
			return 0, err
		}
		v = i
	}
	if v < min {
		return 0, strconv.ErrRange
	}
	if v > max {
		if clampMax {
			return max, nil
		}
		return 0, strconv.ErrRange
	}
	return v, nil
}

func isCalibreDownloadURL(s string) bool {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return false
	}
	return strings.Contains(s, "/opds/download/")
}

func isArchiveDownloadURL(s string) bool {
	return reArchiveURL.MatchString(s)
}

func isLiber3ID(s string) bool {
	return reLiber3ID.MatchString(s)
}

func isAnnasID(s string) bool {
	return reAnnasID.MatchString(s)
}

func isZlibID(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func isZlibHash(s string) bool {
	return reZlibHash.MatchString(s)
}

func detectDownloadRoute(arg1, arg2 string) DownloadRoute {
	if arg1 == "" {
		return RouteUnknown
	}
	if arg2 != "" && isZlibID(arg1) && isZlibHash(arg2) {
		return RouteZlibIDHash
	}
	if isCalibreDownloadURL(arg1) {
		return RouteCalibreURL
	}
	if isArchiveDownloadURL(arg1) {
		return RouteArchiveURL
	}
	if isLiber3ID(arg1) {
		return RouteLiber3ID
	}
	if isAnnasID(arg1) {
		return RouteAnnasID
	}
	return RouteUnknown
}

func formatBookLine(i int, b BookItem) string {
	parts := []string{
		strconv.Itoa(i) + ". [" + b.Source + "] " + nz(b.Title),
	}
	if b.Authors != "" {
		parts = append(parts, "作者: "+b.Authors)
	}
	if b.Year != "" {
		parts = append(parts, "年份: "+b.Year)
	}
	if b.Format != "" {
		parts = append(parts, "格式: "+b.Format)
	}
	if b.ID != "" {
		parts = append(parts, "ID: "+b.ID)
	}
	if b.Hash != "" {
		parts = append(parts, "Hash: "+b.Hash)
	}
	if b.DownloadURL != "" {
		parts = append(parts, "下载: "+b.DownloadURL)
	}
	return strings.Join(parts, "\n")
}

func nz(s string) string {
	if strings.TrimSpace(s) == "" {
		return "未知"
	}
	return strings.TrimSpace(s)
}

func joinURL(base, rel string) string {
	b, err := url.Parse(base)
	if err != nil {
		return ""
	}
	r, err := url.Parse(rel)
	if err != nil {
		return ""
	}
	return b.ResolveReference(r).String()
}

func buildSearchBlockTexts(source string, items []BookItem, err error, maxLen int) []string {
	if maxLen <= 0 {
		maxLen = 3000
	}
	if err != nil {
		return []string{fmt.Sprintf("[%s] 查询失败: %v", source, err)}
	}
	if len(items) == 0 {
		return []string{fmt.Sprintf("[%s] 无结果", source)}
	}

	entries := make([]string, 0, len(items))
	for i, item := range items {
		entries = append(entries, formatBookLine(i+1, item))
	}
	return paginateSourceEntries(source, entries, maxLen)
}

func paginateSourceEntries(source string, entries []string, maxLen int) []string {
	if len(entries) == 0 {
		return []string{fmt.Sprintf("[%s] 无结果", source)}
	}

	bodyLimit := maxLen - len(source) - 16
	if bodyLimit < 200 {
		bodyLimit = 200
	}

	pages := make([]string, 0, 4)
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			pages = append(pages, current.String())
			current.Reset()
		}
	}

	for _, entry := range entries {
		chunks := splitLongText(entry, bodyLimit)
		for i, chunk := range chunks {
			segment := chunk
			if i < len(chunks)-1 {
				segment += "\n(continued)"
			}
			if current.Len() == 0 {
				current.WriteString(segment)
				continue
			}
			if current.Len()+2+len(segment) > bodyLimit {
				flush()
				current.WriteString(segment)
				continue
			}
			current.WriteString("\n\n")
			current.WriteString(segment)
		}
	}
	flush()

	if len(pages) == 1 {
		return []string{fmt.Sprintf("[%s]\n%s", source, pages[0])}
	}

	result := make([]string, 0, len(pages))
	for i, page := range pages {
		result = append(result, fmt.Sprintf("[%s %d/%d]\n%s", source, i+1, len(pages), page))
	}
	return result
}

func splitLongText(s string, maxLen int) []string {
	if maxLen <= 0 || len(s) <= maxLen {
		return []string{s}
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return []string{s}
	}
	out := make([]string, 0, (len(runes)/maxLen)+1)
	for start := 0; start < len(runes); start += maxLen {
		end := start + maxLen
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[start:end]))
	}
	return out
}
