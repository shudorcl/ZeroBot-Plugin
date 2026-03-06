package ebooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type archiveSource struct {
	client *http.Client
}

func newArchiveSource() *archiveSource {
	return &archiveSource{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *archiveSource) Name() string { return "archive.org" }

func (s *archiveSource) Enabled(c *Config) bool { return c.EnableArchive }

func (s *archiveSource) Search(query string, limit int) ([]BookItem, error) {
	values := url.Values{
		"q":      []string{fmt.Sprintf(`title:"%s" mediatype:texts`, query)},
		"fl[]":   []string{"identifier,title"},
		"sort[]": []string{"downloads desc"},
		"rows":   []string{fmt.Sprintf("%d", limit+8)},
		"page":   []string{"1"},
		"output": []string{"json"},
	}
	api := "https://archive.org/advancedsearch.php?" + values.Encode()
	req, _ := http.NewRequest("GET", api, nil)
	req.Header.Set("User-Agent", "ZeroBot-Plugin/ebooks")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("archive search status: %d", resp.StatusCode)
	}
	var out struct {
		Response struct {
			Docs []struct {
				Identifier string `json:"identifier"`
				Title      string `json:"title"`
			} `json:"docs"`
		} `json:"response"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	items := make([]BookItem, 0, limit)
	for _, d := range out.Response.Docs {
		if len(items) >= limit {
			break
		}
		meta, e := s.fetchMeta(d.Identifier)
		if e != nil || meta.DownloadURL == "" {
			continue
		}
		meta.Title = nz(d.Title)
		meta.Source = s.Name()
		items = append(items, meta)
	}
	return items, nil
}

func (s *archiveSource) Download(arg1, _ string) (*DownloadResult, error) {
	if !isArchiveDownloadURL(arg1) {
		return nil, fmt.Errorf("invalid archive url")
	}
	return &DownloadResult{
		Source: s.Name(),
		URL:    arg1,
	}, nil
}

func (s *archiveSource) fetchMeta(identifier string) (BookItem, error) {
	var r BookItem
	req, _ := http.NewRequest("GET", "https://archive.org/metadata/"+identifier, nil)
	req.Header.Set("User-Agent", "ZeroBot-Plugin/ebooks")
	resp, err := s.client.Do(req)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return r, fmt.Errorf("archive metadata status: %d", resp.StatusCode)
	}
	var out struct {
		Metadata map[string]any `json:"metadata"`
		Files    []struct {
			Name string `json:"name"`
		} `json:"files"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return r, err
	}
	var file string
	var format string
	for _, f := range out.Files {
		n := strings.ToLower(f.Name)
		if strings.HasSuffix(n, ".pdf") {
			file = f.Name
			format = "pdf"
			break
		}
		if file == "" && strings.HasSuffix(n, ".epub") {
			file = f.Name
			format = "epub"
		}
	}
	if file == "" {
		return r, nil
	}
	r.Authors = anyToString(out.Metadata["creator"])
	r.Publisher = anyToString(out.Metadata["publisher"])
	r.Language = anyToString(out.Metadata["language"])
	r.Description = trimDesc(anyToString(out.Metadata["description"]))
	pubDate := anyToString(out.Metadata["publicdate"])
	if len(pubDate) >= 4 {
		r.Year = pubDate[:4]
	}
	r.Format = format
	r.ID = identifier
	r.DownloadURL = fmt.Sprintf("https://archive.org/download/%s/%s", identifier, url.PathEscape(file))
	r.CoverURL = "https://archive.org/services/img/" + identifier
	return r, nil
}

func anyToString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []any:
		if len(x) == 0 {
			return ""
		}
		return anyToString(x[0])
	default:
		return ""
	}
}

func trimDesc(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 180 {
		return s[:180] + "..."
	}
	return s
}

func fetchBinary(target string) ([]byte, string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("User-Agent", "ZeroBot-Plugin/ebooks")
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download status: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return b, resp.Header.Get("Content-Disposition"), nil
}
