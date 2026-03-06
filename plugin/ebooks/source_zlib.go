package ebooks

import "fmt"

type zlibSource struct{}

func newZlibSource() *zlibSource { return &zlibSource{} }

func (s *zlibSource) Name() string { return "Z-Library" }

func (s *zlibSource) Enabled(c *Config) bool { return c.EnableZlib }

func (s *zlibSource) Search(query string, limit int) ([]BookItem, error) {
	_ = query
	_ = limit
	return nil, fmt.Errorf("Z-Library 需要账号接口支持，当前版本请使用下书 <ID> <Hash> 或关闭该源")
}

func (s *zlibSource) Download(arg1, arg2 string) (*DownloadResult, error) {
	if !isZlibID(arg1) || !isZlibHash(arg2) {
		return nil, fmt.Errorf("zlib 参数格式错误，正确格式: 下书 <ID> <Hash>")
	}
	return &DownloadResult{
		Source: s.Name(),
		URL:    fmt.Sprintf("https://z-library.sk/dl/%s/%s", arg1, arg2),
	}, nil
}
