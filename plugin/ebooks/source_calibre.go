package ebooks

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type calibreSource struct {
	client *http.Client
	cfgFn  func() Config
}

func newCalibreSource(cfgFn func() Config) *calibreSource {
	return &calibreSource{
		client: &http.Client{Timeout: 60 * time.Second},
		cfgFn:  cfgFn,
	}
}

func (s *calibreSource) Name() string { return "Calibre-Web" }

func (s *calibreSource) Enabled(c *Config) bool {
	return c.EnableCalibre && strings.TrimSpace(c.CalibreWebURL) != ""
}

func (s *calibreSource) Search(query string, limit int) ([]BookItem, error) {
	cfg := s.cfgFn()
	base := strings.TrimRight(cfg.CalibreWebURL, "/")
	reqURL := base + "/opds/search/" + url.QueryEscape(query)
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("calibre search status: %d", resp.StatusCode)
	}
	var feed struct {
		Entries []struct {
			Title     string `xml:"title"`
			Summary   string `xml:"summary"`
			Published string `xml:"published"`
			Author    []struct {
				Name string `xml:"name"`
			} `xml:"author"`
			Link []struct {
				Rel  string `xml:"rel,attr"`
				Href string `xml:"href,attr"`
				Type string `xml:"type,attr"`
			} `xml:"link"`
		} `xml:"entry"`
	}
	if err = xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}
	items := make([]BookItem, 0, limit)
	for _, e := range feed.Entries {
		if len(items) >= limit {
			break
		}
		it := BookItem{
			Source:      s.Name(),
			Title:       nz(e.Title),
			Description: trimDesc(e.Summary),
		}
		if len(e.Published) >= 4 {
			it.Year = e.Published[:4]
		}
		if len(e.Author) > 0 {
			names := make([]string, 0, len(e.Author))
			for _, a := range e.Author {
				if a.Name != "" {
					names = append(names, a.Name)
				}
			}
			it.Authors = strings.Join(names, ", ")
		}
		for _, link := range e.Link {
			if strings.Contains(link.Rel, "acquisition") {
				it.DownloadURL = joinURL(base, link.Href)
				it.Format = link.Type
				break
			}
		}
		if it.DownloadURL == "" {
			continue
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *calibreSource) Download(arg1, _ string) (*DownloadResult, error) {
	if !isCalibreDownloadURL(arg1) {
		return nil, fmt.Errorf("invalid calibre download url")
	}
	return &DownloadResult{
		Source: s.Name(),
		URL:    arg1,
	}, nil
}
