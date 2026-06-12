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
	Name    string
	Bits    string
	Nature  string // 卦象属性：天、地、雷、风、水、火、山、泽
	Quality string // 卦象特质
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
			"不附带问题时显示基础卦象解析；附带问题时会复用 AI 聊天配置调用大模型深度解析。",
	}).ApplySingle(ctxext.DefaultSingle)

	trigrams = map[string]trigram{
		"111": {Name: "乾", Bits: "111", Nature: "天", Quality: "刚健、主动、开创"},
		"110": {Name: "兑", Bits: "110", Nature: "泽", Quality: "悦纳、交流、取舍"},
		"101": {Name: "离", Bits: "101", Nature: "火", Quality: "明辨、依附、显现"},
		"100": {Name: "震", Bits: "100", Nature: "雷", Quality: "发动、惊醒、行动"},
		"011": {Name: "巽", Bits: "011", Nature: "风", Quality: "入微、顺势、渗透"},
		"010": {Name: "坎", Bits: "010", Nature: "水", Quality: "险阻、流动、试炼"},
		"001": {Name: "艮", Bits: "001", Nature: "山", Quality: "止息、边界、沉稳"},
		"000": {Name: "坤", Bits: "000", Nature: "地", Quality: "承载、顺应、养成"},
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
	// hexagramTexts 六十四卦卦辞
	hexagramTexts = map[int]string{
		1:  "元亨利贞。形势重在自强、清明与持续推进。",
		2:  "厚德载物。此时宜承接现实、稳住根基。",
		3:  "初生多难。先立秩序，再求展开。",
		4:  "蒙以养正。问题的关键在学习与求明。",
		5:  "云上于天。时机未熟，守正等待。",
		6:  "天水相违。分歧已现，宜慎言明界。",
		7:  "地中有水。需要纪律、组织与共同目标。",
		8:  "水在地上。亲比有道，先辨同路人。",
		9:  "风行天上。积小成势，暂不宜强攻。",
		10: "泽上于天。行事需知礼、知险、知分寸。",
		11: "天地交泰。上下相通，宜推进协作。",
		12: "天地不交。闭塞之时，先保全正道。",
		13: "天火同明。求同存异，公开透明。",
		14: "火在天上。资源可用，贵在不骄。",
		15: "地中有山。退让不是弱，能聚人心。",
		16: "雷出地奋。顺势而动，但勿沉于安逸。",
		17: "泽中有雷。随时变通，仍需守住主心。",
		18: "山下有风。旧弊需整治，先查根源。",
		19: "地泽相临。机会渐近，宜以诚接物。",
		20: "风行地上。先观察全局，再定取向。",
		21: "雷电交作。阻隔需决断处理。",
		22: "山下有火。文饰可助成事，勿掩其实。",
		23: "山附于地。消退之象，宜守不宜争。",
		24: "雷在地中。转机初回，贵在小步复正。",
		25: "天下雷行。顺其自然，勿妄动求速。",
		26: "山中蓄天。蓄力养德，待机而发。",
		27: "山下有雷。养正养身，慎其所入口。",
		28: "泽灭木。压力过重，需调整结构。",
		29: "重险相仍。守信行险，步步求实。",
		30: "明两作。看清依附关系，保持光明。",
		31: "泽山相感。感应相通，贵在真诚。",
		32: "雷风相与。长久之道，在节奏稳定。",
		33: "天下有山。退避不是逃，保存主动。",
		34: "雷在天上。势强宜正，不可逞强。",
		35: "火出地上。明德上行，适合显露成果。",
		36: "明入地中。光受遮蔽，宜韬晦守心。",
		37: "风自火出。内在秩序决定外部安定。",
		38: "火泽相违。差异明显，求小同避大争。",
		39: "水山蹇难。前路受阻，宜求援改道。",
		40: "雷雨作。困局可解，先松动关键结。",
		41: "山下有泽。有所减损，是为了成其重。",
		42: "风雷相益。利于行动与增益他人。",
		43: "泽上于天。决断已至，须明快而不暴。",
		44: "天下有风。偶遇有力，慎始慎入。",
		45: "泽上于地。众人聚合，需有中心。",
		46: "地中生木。循序上升，不急于求成。",
		47: "泽无水。资源受限，守心胜过求表。",
		48: "木上有水。回到根本，修井养人。",
		49: "泽中有火。变革当时，先正名分。",
		50: "火上有木。更新制度，化材成器。",
		51: "洊雷。震动带来警醒，动后需定。",
		52: "兼山。止于其所，边界即力量。",
		53: "山上有木。渐进成事，重在次第。",
		54: "泽上有雷。关系未正，慎重承诺。",
		55: "雷电皆至。盛大之时，防遮蔽与过满。",
		56: "山上有火。身在途中，守礼少求。",
		57: "随风。入而能柔，细处见功。",
		58: "丽泽。交流悦人，也要守信。",
		59: "风行水上。离散可化，先通其气。",
		60: "泽上有水。节制成度，过严亦失。",
		61: "风泽中孚。诚信在中，感而能通。",
		62: "雷在山上。小事可为，大事宜慎。",
		63: "水火既济。事已成形，防成后之乱。",
		64: "火水未济。尚未完成，调序即可过渡。",
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
	if r.Question == "" {
		build.WriteString("\n\n卦象解析:\n")
		build.WriteString(r.buildInterpretation())
	}
	return build.String()
}

func (r divinationResult) hexagramLine(label string, h hexagram) string {
	return label + ": 第" + strconv.Itoa(h.Number) + "卦 " + h.Name + "（" + h.Upper.Name + "上" + h.Lower.Name + "下）"
}

// buildInterpretation 构建卦象解析（不依赖 LLM）
func (r divinationResult) buildInterpretation() string {
	var build strings.Builder
	upper := r.Original.Upper
	lower := r.Original.Lower

	// 本卦卦辞
	if text, ok := hexagramTexts[r.Original.Number]; ok {
		build.WriteString("【本卦】")
		build.WriteString(r.Original.Name)
		build.WriteString("：")
		build.WriteString(text)
		build.WriteString("\n")
	}

	// 上下卦象关系
	build.WriteString("上卦")
	build.WriteString(upper.Name)
	build.WriteString("（")
	build.WriteString(upper.Nature)
	build.WriteString("）")
	build.WriteString("，特质为「")
	build.WriteString(upper.Quality)
	build.WriteString("」；下卦")
	build.WriteString(lower.Name)
	build.WriteString("（")
	build.WriteString(lower.Nature)
	build.WriteString("）")
	build.WriteString("，特质为「")
	build.WriteString(lower.Quality)
	build.WriteString("」。\n")

	// 动爻信息
	movingCount := 0
	for _, v := range r.Yao {
		if v == oldYin || v == oldYang {
			movingCount++
		}
	}

	if movingCount > 0 {
		build.WriteString("动爻")
		build.WriteString(formatMovingLines(r.Yao))
		build.WriteString("，表示变化集中于此，需重点关注。\n")
	}

	// 变卦信息
	if r.Original.Number != r.Changed.Number {
		build.WriteString("【变卦】由")
		build.WriteString(r.Original.Name)
		build.WriteString("趋向")
		build.WriteString(r.Changed.Name)
		build.WriteString("，表示处境中已有可转化的力量。\n")
		if text, ok := hexagramTexts[r.Changed.Number]; ok {
			build.WriteString(r.Changed.Name)
			build.WriteString("：")
			build.WriteString(text)
			build.WriteString("\n")
		}
	} else {
		build.WriteString("此卦无变，重点在守住当前结构，把问题看完整。\n")
	}

	// 综合建议
	if movingCount >= 3 {
		build.WriteString("动爻较多，环境变化快，先抓主要矛盾，避免同时处理所有问题。")
	} else if movingCount > 0 {
		build.WriteString("变化集中在少数位置，适合小幅调整、验证后再扩大行动。")
	} else {
		build.WriteString("局面暂稳，适合复盘动机、资源与边界。")
	}

	return build.String()
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
		return "━━━━"
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
