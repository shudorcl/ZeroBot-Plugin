// Package ebooks multi-source ebook search plugin.
package ebooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zbpctxext "github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	helpText = "- 搜书 <关键词> [数量]\n" +
		"- 下书 <标识> [hash]\n" +
		"- 搜书帮助\n" +
		"- 查看搜书配置 (私聊超管)\n" +
		"- 设置搜书配置 <key> <value> (私聊超管)\n" +
		"- 开启书源 <archive|liber3|calibre|zlib|annas> (私聊超管)\n" +
		"- 关闭书源 <archive|liber3|calibre|zlib|annas> (私聊超管)"

	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Extra:             control.ExtraFromString("ebooks"),
		Brief:             "多书源搜书和下书",
		Help:              helpText,
		PrivateDataFolder: "ebooks",
	})

	cfg   = defaultConfig()
	cfgMu sync.RWMutex

	archiveS = newArchiveSource()
	liber3S  = newLiber3Source()
	calibreS = newCalibreSource(func() Config { return getCfg() })
	annasS   = newAnnasSource()
	zlibS    = newZlibSource()

	ensureCfg = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		loadCfg(ctx)
		return true
	})
)

type searchResult struct {
	source string
	items  []BookItem
	err    error
}

func init() {
	engine.OnFullMatch("搜书帮助").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text(helpText))
	})

	engine.OnPrefix("搜书", ensureCfg).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		arg := strings.TrimSpace(ctx.State["args"].(string))
		if arg == "" {
			ctx.SendChain(message.Text("用法: 搜书 <关键词> [数量]"))
			return
		}

		parts := strings.Fields(arg)
		limitText := ""
		query := arg
		if len(parts) >= 2 {
			last := parts[len(parts)-1]
			if allDigits(last) {
				limitText = last
				query = strings.TrimSpace(strings.TrimSuffix(arg, last))
			}
		}

		c := getCfg()
		limit, err := normalizeLimit(limitText, c.MaxResults, 1, 50, true)
		if err != nil {
			ctx.SendChain(message.Text("数量范围应为 1-50"))
			return
		}

		srcs := []Source{calibreS, liber3S, archiveS, zlibS, annasS}
		out := make(chan searchResult, len(srcs))
		var wg sync.WaitGroup
		for _, s := range srcs {
			if !s.Enabled(&c) {
				continue
			}
			wg.Add(1)
			go func(src Source) {
				defer wg.Done()
				items, searchErr := src.Search(query, limit)
				out <- searchResult{
					source: src.Name(),
					items:  items,
					err:    searchErr,
				}
			}(s)
		}
		wg.Wait()
		close(out)

		resultsBySource := make(map[string]searchResult, len(srcs))
		for res := range out {
			resultsBySource[res.source] = res
		}

		results := make([]searchResult, 0, len(resultsBySource))
		for _, src := range srcs {
			if !src.Enabled(&c) {
				continue
			}
			if res, ok := resultsBySource[src.Name()]; ok {
				results = append(results, res)
			}
		}
		if len(results) == 0 {
			ctx.SendChain(message.Text("无可用书源，请联系管理员检查插件配置"))
			return
		}

		msg := buildSearchForwardMessage(ctx, query, limit, results)
		if id := ctx.Send(msg).ID(); id == 0 {
			ctx.SendChain(message.Text(buildSearchFallbackSummary(query, results)))
		}
	})

	engine.OnPrefix("下书", ensureCfg).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		arg := strings.TrimSpace(ctx.State["args"].(string))
		if arg == "" {
			ctx.SendChain(message.Text("用法: 下书 <标识> [hash]"))
			return
		}
		parts := strings.Fields(arg)
		arg1 := parts[0]
		arg2 := ""
		if len(parts) > 1 {
			arg2 = parts[1]
		}

		var src Source
		switch detectDownloadRoute(arg1, arg2) {
		case RouteZlibIDHash:
			src = zlibS
		case RouteCalibreURL:
			src = calibreS
		case RouteArchiveURL:
			src = archiveS
		case RouteLiber3ID:
			src = liber3S
		case RouteAnnasID:
			src = annasS
		default:
			ctx.SendChain(message.Text("无法识别下载标识。支持 Calibre/Archive URL、Liber3 ID、Annas ID、Zlib ID+Hash"))
			return
		}

		c := getCfg()
		if !src.Enabled(&c) {
			ctx.SendChain(message.Text("[", src.Name(), "] 未启用"))
			return
		}

		res, err := src.Download(arg1, arg2)
		if err != nil {
			ctx.SendChain(message.Text("[", src.Name(), "] ", err))
			return
		}
		sendDownloadResult(ctx, &c, res)
	})

	engine.OnFullMatch("查看搜书配置", zero.OnlyPrivate, zero.SuperUserPermission, ensureCfg).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c := getCfg()
			ctx.SendChain(message.Text(fmt.Sprintf(
				"calibre=%v\nliber3=%v\narchive=%v\nzlib=%v\nannas=%v\ncalibre_web_url=%s\nmax_results=%d\ndownload_mode=%s",
				c.EnableCalibre,
				c.EnableLiber3,
				c.EnableArchive,
				c.EnableZlib,
				c.EnableAnnas,
				c.CalibreWebURL,
				c.MaxResults,
				c.DownloadMode,
			)))
		})

	engine.OnRegex(`^设置搜书配置\s+([a-zA-Z_]+)\s+(.+)$`, zero.OnlyPrivate, zero.SuperUserPermission, ensureCfg).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			c := getCfg()
			r := ctx.State["regex_matched"].([]string)
			key := strings.ToLower(strings.TrimSpace(r[1]))
			val := strings.TrimSpace(r[2])

			switch key {
			case "calibre_web_url":
				c.CalibreWebURL = val
				c.EnableCalibre = val != ""
			case "max_results":
				v, err := normalizeLimit(val, c.MaxResults, 1, 50, false)
				if err != nil {
					ctx.SendChain(message.Text("max_results 需为 1-50"))
					return
				}
				c.MaxResults = v
			case "download_mode":
				c.DownloadMode = val
			case "zlib_email":
				c.ZlibEmail = val
			case "zlib_password":
				c.ZlibPassword = val
			default:
				ctx.SendChain(message.Text("不支持的配置项"))
				return
			}

			sanitizeConfig(&c)
			if err := setCfg(m, c); err != nil {
				ctx.SendChain(message.Text("保存配置失败: ", err))
				return
			}
			ctx.SendChain(message.Text("已更新配置"))
		})

	engine.OnRegex(`^(开启|关闭)书源\s+(archive|liber3|calibre|zlib|annas)$`, zero.OnlyPrivate, zero.SuperUserPermission, ensureCfg).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			c := getCfg()
			r := ctx.State["regex_matched"].([]string)
			on := r[1] == "开启"

			switch r[2] {
			case "archive":
				c.EnableArchive = on
			case "liber3":
				c.EnableLiber3 = on
			case "calibre":
				c.EnableCalibre = on
			case "zlib":
				c.EnableZlib = on
			case "annas":
				c.EnableAnnas = on
			}

			sanitizeConfig(&c)
			if err := setCfg(m, c); err != nil {
				ctx.SendChain(message.Text("保存失败: ", err))
				return
			}
			ctx.SendChain(message.Text("已", r[1], "书源 ", r[2]))
		})
}

func getCfg() Config {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return cfg
}

func setCfg(m *ctrl.Control[*zero.Ctx], c Config) error {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	cfg = c
	return m.SetExtra(&cfg)
}

func loadCfg(ctx *zero.Ctx) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	cfg = defaultConfig()
	m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	_ = m.GetExtra(&cfg)
	sanitizeConfig(&cfg)
	if err := m.SetExtra(&cfg); err != nil {
		logrus.Warnln("[ebooks] save default config failed:", err)
	}
}

func sendDownloadResult(ctx *zero.Ctx, c *Config, res *DownloadResult) {
	if res == nil {
		ctx.SendChain(message.Text("下载结果为空"))
		return
	}
	sendLink := c.DownloadMode == "link" || c.DownloadMode == "both"
	sendUpload := c.DownloadMode == "upload" || c.DownloadMode == "both"

	if sendLink && res.URL != "" {
		ctx.SendChain(message.Text("[", res.Source, "] 下载链接: ", res.URL))
	}
	if !sendUpload {
		return
	}
	if res.URL == "" {
		ctx.SendChain(message.Text("[", res.Source, "] 当前无可上传链接"))
		return
	}
	if ctx.Event.GroupID == 0 {
		ctx.SendChain(message.Text("[", res.Source, "] 私聊暂不支持直传文件，请使用下载链接"))
		return
	}

	data, cd, err := fetchBinary(res.URL)
	if err != nil {
		ctx.SendChain(message.Text("[", res.Source, "] 文件下载失败: ", err))
		return
	}
	name := res.FileName
	if name == "" {
		name = parseFilename(cd)
	}
	if name == "" {
		name = "ebook.bin"
	}
	tmp := filepath.Join(engine.DataFolder(), "tmp")
	_ = os.MkdirAll(tmp, 0o755)
	p := filepath.Join(tmp, sanitizeFilename(name))
	if err = os.WriteFile(p, data, 0o644); err != nil {
		ctx.SendChain(message.Text("[", res.Source, "] 保存文件失败: ", err))
		return
	}
	ctx.UploadThisGroupFile(filepath.Join(file.BOTPATH, p), filepath.Base(p), "")
}

func parseFilename(contentDisposition string) string {
	if contentDisposition == "" {
		return ""
	}
	for _, part := range strings.Split(contentDisposition, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "filename=") {
			v := strings.TrimPrefix(part, "filename=")
			v = strings.Trim(v, `"`)
			return v
		}
	}
	return ""
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func buildSearchForwardMessage(ctx *zero.Ctx, query string, limit int, results []searchResult) message.Message {
	msg := make(message.Message, 0, len(results)+1)
	msg = append(msg, zbpctxext.FakeSenderForwardNode(ctx, message.Text(
		fmt.Sprintf("搜书结果\n关键词: %s\n请求数量: %d\n书源数: %d", query, limit, len(results)),
	)))
	for _, res := range results {
		for _, block := range buildSearchBlockTexts(res.source, res.items, res.err, 3000) {
			msg = append(msg, zbpctxext.FakeSenderForwardNode(ctx, message.Text(block)))
		}
	}
	return msg
}

func buildSearchFallbackSummary(query string, results []searchResult) string {
	var b strings.Builder
	b.WriteString("搜书结果发送失败，以下为摘要\n")
	b.WriteString("关键词: ")
	b.WriteString(query)
	for _, res := range results {
		b.WriteString("\n[")
		b.WriteString(res.source)
		b.WriteString("] ")
		if res.err != nil {
			b.WriteString("失败: ")
			b.WriteString(res.err.Error())
		} else {
			b.WriteString(fmt.Sprintf("%d 条结果", len(res.items)))
		}
	}
	return b.String()
}
