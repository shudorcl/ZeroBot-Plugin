package ebooks

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type annasSource struct {
	client *http.Client
}

func newAnnasSource() *annasSource {
	return &annasSource{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *annasSource) Name() string { return "Anna's Archive" }

func (s *annasSource) Enabled(c *Config) bool { return c.EnableAnnas }

func (s *annasSource) Search(query string, limit int) ([]BookItem, error) {
	req, _ := http.NewRequest("GET", "https://annas-archive.org/search?q="+query, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("annas search status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	html := string(body)

	re := regexp.MustCompile(`/md5/([a-fA-F0-9]{32})`)
	found := re.FindAllStringSubmatch(html, limit*2)
	seen := map[string]struct{}{}
	items := make([]BookItem, 0, limit)
	for _, m := range found {
		if len(items) >= limit {
			break
		}
		id := strings.ToUpper(m[1])
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		items = append(items, BookItem{
			Source:      s.Name(),
			Title:       "Anna's Archive 资源",
			ID:          "A" + id,
			DownloadURL: "https://annas-archive.org/md5/" + strings.ToLower(id),
		})
	}
	return items, nil
}

func (s *annasSource) Download(arg1, _ string) (*DownloadResult, error) {
	if !isAnnasID(arg1) {
		return nil, fmt.Errorf("invalid annas id")
	}
	id := strings.ToLower(strings.TrimPrefix(arg1, "A"))
	return &DownloadResult{
		Source: s.Name(),
		URL:    "https://annas-archive.org/md5/" + id,
	}, nil
}
