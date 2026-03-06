package ebooks

type Config struct {
	EnableCalibre bool   `json:"enable_calibre"`
	EnableLiber3  bool   `json:"enable_liber3"`
	EnableArchive bool   `json:"enable_archive"`
	EnableZlib    bool   `json:"enable_zlib"`
	EnableAnnas   bool   `json:"enable_annas"`
	CalibreWebURL string `json:"calibre_web_url"`
	ZlibEmail     string `json:"zlib_email"`
	ZlibPassword  string `json:"zlib_password"`
	MaxResults    int    `json:"max_results"`
	DownloadMode  string `json:"download_mode"` // link|upload|both
}

type BookItem struct {
	Source      string
	Title       string
	Authors     string
	Year        string
	Publisher   string
	Language    string
	Description string
	Format      string
	FileSize    string
	CoverURL    string
	DownloadURL string
	ID          string
	Hash        string
}

type DownloadResult struct {
	Source   string
	FileName string
	URL      string
	Data     []byte
}

type Source interface {
	Name() string
	Enabled(*Config) bool
	Search(query string, limit int) ([]BookItem, error)
	Download(arg1, arg2 string) (*DownloadResult, error)
}

type DownloadRoute int

const (
	RouteUnknown DownloadRoute = iota
	RouteZlibIDHash
	RouteCalibreURL
	RouteArchiveURL
	RouteLiber3ID
	RouteAnnasID
)
