// Package bugua 周易卜卦
package bugua

import (
	"math/rand"
	"strconv"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	oldYin    = 6
	youngYang = 7
	youngYin  = 8
	oldYang   = 9
)

type yaoValues [6]int

type trigram struct {
	Name string
	Bits string
}

type hexagram struct {
	Number int
	Name   string
	Upper  trigram
	Lower  trigram
}

type divinationResult struct {
	Question string
	Original hexagram
	Changed  hexagram
	Yao      yaoValues
}

var (
	en = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "周易卜卦",
		Help: "- 卜卦 [询问的事情]\n" +
			"- 起卦 [询问的事情]\n" +
			"附带问题时会复用 AI 聊天配置调用大模型解析。",
	}).ApplySingle(ctxext.DefaultSingle)

	trigrams = map[string]trigram{
		"111": {Name: "乾", Bits: "111"},
		"110": {Name: "兑", Bits: "110"},
		"101": {Name: "离", Bits: "101"},
		"100": {Name: "震", Bits: "100"},
		"011": {Name: "巽", Bits: "011"},
		"010": {Name: "坎", Bits: "010"},
		"001": {Name: "艮", Bits: "001"},
		"000": {Name: "坤", Bits: "000"},
	}
	kingWenTable = map[string]map[string]int{
		"乾": {"乾": 1, "兑": 43, "离": 14, "震": 34, "巽": 9, "坎": 5, "艮": 26, "坤": 11},
		"兑": {"乾": 10, "兑": 58, "离": 38, "震": 54, "巽": 61, "坎": 60, "艮": 41, "坤": 19},
		"离": {"乾": 13, "兑": 49, "离": 30, "震": 55, "巽": 37, "坎": 63, "艮": 22, "坤": 36},
		"震": {"乾": 25, "兑": 17, "离": 21, "震": 51, "巽": 42, "坎": 3, "艮": 27, "坤": 24},
		"巽": {"乾": 44, "兑": 28, "离": 50, "震": 32, "巽": 57, "坎": 48, "艮": 18, "坤": 46},
		"坎": {"乾": 6, "兑": 47, "离": 64, "震": 40, "巽": 59, "坎": 29, "艮": 4, "坤": 7},
		"艮": {"乾": 33, "兑": 31, "离": 56, "震": 62, "巽": 53, "坎": 39, "艮": 52, "坤": 15},
		"坤": {"乾": 12, "兑": 45, "离": 35, "震": 16, "巽": 20, "坎": 8, "艮": 23, "坤": 2},
	}
	hexagramNames = map[int]string{
		1: "乾", 2: "坤", 3: "屯", 4: "蒙", 5: "需", 6: "讼", 7: "师", 8: "比",
		9: "小畜", 10: "履", 11: "泰", 12: "否", 13: "同人", 14: "大有", 15: "谦", 16: "豫",
		17: "随", 18: "蛊", 19: "临", 20: "观", 21: "噬嗑", 22: "贲", 23: "剥", 24: "复",
		25: "无妄", 26: "大畜", 27: "颐", 28: "大过", 29: "坎", 30: "离", 31: "咸", 32: "恒",
		33: "遁", 34: "大壮", 35: "晋", 36: "明夷", 37: "家人", 38: "睽", 39: "蹇", 40: "解",
		41: "损", 42: "益", 43: "夬", 44: "姤", 45: "萃", 46: "升", 47: "困", 48: "井",
		49: "革", 50: "鼎", 51: "震", 52: "艮", 53: "渐", 54: "归妹", 55: "丰", 56: "旅",
		57: "巽", 58: "兑", 59: "涣", 60: "节", 61: "中孚", 62: "小过", 63: "既济", 64: "未济",
	}
)

func init() {
	en.OnRegex(`^(卜卦|起卦)\s?(.*)$`).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		question := strings.TrimSpace(ctx.State["regex_matched"].([]string)[2])
		result := newDivination(question)
		ctx.SendChain(message.Text(result.summary()))
		if question == "" {
			return
		}
		if !chat.EnsureConfig(ctx) {
			ctx.SendChain(message.Text("卜卦解析失败: 无法读取 AI 聊天配置"))
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		stor, err := chat.NewStorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("卜卦解析失败: ", errors.Wrap(err, "读取 AI 聊天温度配置失败")))
			return
		}
		reply, err := result.analyze(stor.Temp())
		if err != nil {
			logrus.Warnln("[bugua]大模型解析失败:", err)
			ctx.SendChain(message.Text("卜卦解析失败: ", err))
			return
		}
		if reply == "" {
			ctx.SendChain(message.Text("卜卦解析失败: 大模型返回为空"))
			return
		}
		if id := ctx.Send(makeNodeMessage(reply, ctx.CardOrNickName(ctx.Event.UserID), ctx.Event.UserID)).ID(); id == 0 {
			ctx.SendChain(message.Text("ERROR: 可能被风控了"))
		}
	})
}

func newDivination(question string) divinationResult {
	yao := castYaoValues()
	return divinationResult{
		Question: question,
		Original: lookupHexagram(yao),
		Changed:  lookupHexagram(yao.changed()),
		Yao:      yao,
	}
}

func castYaoValues() (yao yaoValues) {
	for i := range yao {
		sum := 0
		for range 3 {
			sum += rand.Intn(2) + 2
		}
		yao[i] = sum
	}
	return
}

func lookupHexagram(yao yaoValues) hexagram {
	lower := trigrams[yao.trigramBits(0)]
	upper := trigrams[yao.trigramBits(3)]
	number := kingWenTable[lower.Name][upper.Name]
	return hexagram{
		Number: number,
		Name:   hexagramNames[number],
		Upper:  upper,
		Lower:  lower,
	}
}

func (yao yaoValues) trigramBits(start int) string {
	var build strings.Builder
	for i := start; i < start+3; i++ {
		if yao.isYang(i) {
			build.WriteByte('1')
		} else {
			build.WriteByte('0')
		}
	}
	return build.String()
}

func (yao yaoValues) isYang(i int) bool {
	return yao[i] == youngYang || yao[i] == oldYang
}

func (yao yaoValues) changed() (changed yaoValues) {
	for i, v := range yao {
		switch v {
		case oldYin:
			changed[i] = youngYang
		case oldYang:
			changed[i] = youngYin
		default:
			changed[i] = v
		}
	}
	return
}

func (yao yaoValues) movingLines() []string {
	lines := make([]string, 0, 6)
	for i, v := range yao {
		if v == oldYin || v == oldYang {
			lines = append(lines, movingLineName(i, v))
		}
	}
	return lines
}

func movingLineName(index, value int) string {
	num := "六"
	if value == oldYang {
		num = "九"
	}
	switch index {
	case 0:
		return "初" + num
	case 5:
		return "上" + num
	default:
		return num + []string{"", "二", "三", "四", "五"}[index]
	}
}

func (r divinationResult) summary() string {
	var build strings.Builder
	build.WriteString("卜卦结果:\n")
	build.WriteString(r.hexagramLine("本卦", r.Original))
	build.WriteByte('\n')
	build.WriteString(r.hexagramLine("变卦", r.Changed))
	build.WriteByte('\n')
	build.WriteString("动爻: ")
	build.WriteString(formatMovingLines(r.Yao))
	build.WriteString("\n卦象:\n")
	build.WriteString(r.Yao.format())
	return build.String()
}

func (r divinationResult) hexagramLine(label string, h hexagram) string {
	return label + ": 第" + strconv.Itoa(h.Number) + "卦 " + h.Name + "（" + h.Upper.Name + "上" + h.Lower.Name + "下）"
}

func formatMovingLines(yao yaoValues) string {
	lines := yao.movingLines()
	if len(lines) == 0 {
		return "无"
	}
	return strings.Join(lines, "、")
}

func (yao yaoValues) format() string {
	var build strings.Builder
	for i := len(yao) - 1; i >= 0; i-- {
		build.WriteString(movingLineName(i, yaoLineNameValue(yao[i])))
		build.WriteString(" ")
		build.WriteString(yaoLineArt(yao[i]))
		build.WriteString(" ")
		build.WriteString(yaoLineText(yao[i]))
		if i > 0 {
			build.WriteByte('\n')
		}
	}
	return build.String()
}

func yaoLineArt(value int) string {
	if value == youngYang || value == oldYang {
		return "━━━━━"
	}
	return "━  ━"
}

func yaoLineNameValue(value int) int {
	if value == youngYang || value == oldYang {
		return oldYang
	}
	return oldYin
}

func yaoLineText(value int) string {
	switch value {
	case oldYin:
		return "阴爻 动"
	case youngYang:
		return "阳爻"
	case youngYin:
		return "阴爻"
	case oldYang:
		return "阳爻 动"
	default:
		return "未知"
	}
}

func (r divinationResult) prompt() string {
	var build strings.Builder
	build.WriteString("你是一位谨慎的周易卜卦解读者。请围绕用户的问题和本次卦象解读，先说明本卦，再说明动爻与变卦，最后给出综合建议，字数控制在300-500字。")
	build.WriteString("不要把占卜结果表述为确定事实。\n")
	build.WriteString("用户问题: ")
	build.WriteString(r.Question)
	build.WriteByte('\n')
	build.WriteString(r.hexagramLine("本卦", r.Original))
	build.WriteByte('\n')
	build.WriteString(r.hexagramLine("变卦", r.Changed))
	build.WriteByte('\n')
	build.WriteString("动爻: ")
	build.WriteString(formatMovingLines(r.Yao))
	build.WriteByte('\n')
	build.WriteString("卦象:\n")
	build.WriteString(r.Yao.format())
	return build.String()
}

func (r divinationResult) analyze(temperature float32) (string, error) {
	topp, maxn := chat.AC.MParams()
	mod, err := chat.AC.Type.Protocol(chat.AC.ModelName, temperature, topp, maxn, chat.AC.ReasoningEffort)
	if err != nil {
		return "", errors.Wrap(err, "创建 AI 模型协议失败")
	}

	api := deepinfra.NewAPI(chat.AC.API, string(chat.AC.Key))
	data, err := api.Request(mod.User(model.NewContentText(r.prompt())))
	if err != nil {
		return "", errors.Wrap(err, "请求 AI 模型失败")
	}
	return strings.TrimSpace(data), nil
}

func makeNodeMessage(reply, nickname string, userID int64) message.Message {
	chunks := splitTextChunks("卜卦解析:\n"+reply, 1000)
	msg := make(message.Message, 0, len(chunks))
	for _, chunk := range chunks {
		msg = append(msg, message.CustomNode(nickname, userID, message.Message{message.Text(chunk)}))
	}
	return msg
}

func splitTextChunks(txt string, maxRunes int) []string {
	runes := []rune(txt)
	if maxRunes <= 0 || len(runes) <= maxRunes {
		return []string{txt}
	}
	chunks := make([]string, 0, (len(runes)+maxRunes-1)/maxRunes)
	for len(runes) > maxRunes {
		chunks = append(chunks, string(runes[:maxRunes]))
		runes = runes[maxRunes:]
	}
	if len(runes) > 0 {
		chunks = append(chunks, string(runes))
	}
	return chunks
}
