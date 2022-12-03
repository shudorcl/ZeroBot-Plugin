// Package tarot 塔罗牌
package sgs

import (
	"encoding/json"
	"math/rand"
	"strconv"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type cardInfo struct {
	Name  string   `json:"name"`
	URL   string   `json:"url"`
	Lines []string `json:"lines"`
}
type cardSet = map[string]cardInfo

var cardMap = make(cardSet, 80)
var infoMap = make(map[string]cardInfo, 80)
var datapath string

func init() {
	engine := control.Register("sgs", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "三国杀\n" +
			"- 抽武将",
		PublicDataFolder: "Sgs",
	}).ApplySingle(ctxext.DefaultSingle)
	getSgsCard := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		data, err := engine.GetLazyData("sgs.json", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		err = json.Unmarshal(data, &cardMap)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		for _, card := range cardMap {
			infoMap[card.Name] = card
		}
		datapath = file.BOTPATH + "/" + engine.DataFolder()
		logrus.Infof("[sgs]读取%d张三国杀武将", len(cardMap))
		return true
	})
	engine.OnFullMatchGroup([]string{"抽武将"}, getSgsCard).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			i := fcext.RandSenderPerDayN(ctx.Event.UserID, len(cardMap))
			card := cardMap[(strconv.Itoa(i))]

			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text("今天的守护武将是~【", card.Name, "】哒"),
				message.Image("file:///"+datapath+card.URL),
				message.Text("\n【", card.Lines[rand.Intn(len(card.Lines))], "】"),
			)
		})
	engine.OnRegex(`^看武将\s?(.*)`, getSgsCard).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		info, ok := infoMap[match]
		if ok {
			ctx.SendChain(
				message.Text("【", info.Name, "】"),
				message.Image("file:///"+datapath+info.URL),
				message.Text("\n【", info.Lines[rand.Intn(len(info.Lines))], "】"),
			)
			return
		}
		ctx.SendChain(
			message.Text("没有找到", match, "噢……"))
	})
}
