package ebooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type liber3Source struct {
	client *http.Client
}

func newLiber3Source() *liber3Source {
	return &liber3Source{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *liber3Source) Name() string { return "Liber3" }

func (s *liber3Source) Enabled(c *Config) bool { return c.EnableLiber3 }

func (s *liber3Source) Search(query string, limit int) ([]BookItem, error) {
	body, _ := json.Marshal(map[string]string{
		"address": "",
		"word":    query,
	})
	req, _ := http.NewRequest("POST", "https://lgate.glitternode.ru/v1/searchV2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("liber3 search status: %d", resp.StatusCode)
	}
	var out struct {
		Data struct {
			Book []struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Author string `json:"author"`
			} `json:"book"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Data.Book) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, limit)
	for i, b := range out.Data.Book {
		if i >= limit {
			break
		}
		ids = append(ids, b.ID)
	}
	detail, _ := s.fetchDetails(ids)
	items := make([]BookItem, 0, len(ids))
	for i, b := range out.Data.Book {
		if i >= limit {
			break
		}
		item := BookItem{
			Source:  s.Name(),
			Title:   nz(b.Title),
			Authors: b.Author,
			ID:      "L" + b.ID,
		}
		if d, ok := detail[b.ID]; ok {
			item.Year = d.Year
			item.Language = d.Language
			item.Publisher = d.Publisher
			item.Format = d.Extension
			item.FileSize = d.FileSize
			if d.IPFSCID != "" {
				name := sanitizeFilename(item.Title)
				item.DownloadURL = fmt.Sprintf("https://gateway-ipfs.st/ipfs/%s?filename=%s.%s", d.IPFSCID, url.QueryEscape(name), url.QueryEscape(d.Extension))
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *liber3Source) Download(arg1, _ string) (*DownloadResult, error) {
	if !isLiber3ID(arg1) {
		return nil, fmt.Errorf("invalid liber3 id")
	}
	id := strings.TrimPrefix(arg1, "L")
	detail, err := s.fetchDetails([]string{id})
	if err != nil {
		return nil, err
	}
	d, ok := detail[id]
	if !ok || d.IPFSCID == "" || d.Extension == "" {
		return nil, fmt.Errorf("liber3 detail not found")
	}
	name := sanitizeFilename(d.Title)
	return &DownloadResult{
		Source:   s.Name(),
		FileName: fmt.Sprintf("%s.%s", name, d.Extension),
		URL:      fmt.Sprintf("https://gateway-ipfs.st/ipfs/%s?filename=%s.%s", d.IPFSCID, url.QueryEscape(name), url.QueryEscape(d.Extension)),
	}, nil
}

type liber3Detail struct {
	Title     string
	Year      string
	Language  string
	Publisher string
	Extension string
	FileSize  string
	IPFSCID   string
}

func (s *liber3Source) fetchDetails(ids []string) (map[string]liber3Detail, error) {
	if len(ids) == 0 {
		return map[string]liber3Detail{}, nil
	}
	body, _ := json.Marshal(map[string][]string{"book_ids": ids})
	req, _ := http.NewRequest("POST", "https://lgate.glitternode.ru/v1/book", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("liber3 detail status: %d", resp.StatusCode)
	}
	var out struct {
		Data struct {
			Book map[string]struct {
				Book struct {
					Title     string `json:"title"`
					Year      string `json:"year"`
					Language  string `json:"language"`
					Publisher string `json:"publisher"`
					Extension string `json:"extension"`
					FileSize  string `json:"filesize"`
					IPFSCID   string `json:"ipfs_cid"`
				} `json:"book"`
			} `json:"book"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	res := make(map[string]liber3Detail, len(out.Data.Book))
	for k, v := range out.Data.Book {
		res[k] = liber3Detail{
			Title:     v.Book.Title,
			Year:      v.Book.Year,
			Language:  v.Book.Language,
			Publisher: v.Book.Publisher,
			Extension: v.Book.Extension,
			FileSize:  v.Book.FileSize,
			IPFSCID:   v.Book.IPFSCID,
		}
	}
	return res, nil
}

func sanitizeFilename(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "ebook"
	}
	r := strings.NewReplacer(`\`, "_", "/", "_", ":", "_", "*", "_", "?", "_", `"`, "_", "<", "_", ">", "_", "|", "_")
	return r.Replace(s)
}
