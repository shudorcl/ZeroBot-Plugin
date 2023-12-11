package pokemon

import (
	"encoding/json"
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

var cardMap = make(cardSet, 200)
var infoMap = make(map[string]cardInfo, 200)
var datapath string

func init() {
	engine := control.Register("pokemon", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help: "抽宝可梦\n" +
			"- 抽宝可梦",
		PublicDataFolder: "Pokemon",
	}).ApplySingle(ctxext.DefaultSingle)
	getWife := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		data, err := engine.GetLazyData("pokemon.json", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		err = json.Unmarshal(data, &cardMap)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		datapath = file.BOTPATH + "/" + engine.DataFolder()
		logrus.Infof("[pokemon]读取%d只宝可梦", len(cardMap))
		return true
	})
	engine.OnFullMatchGroup([]string{"抽宝可梦"}, getWife).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			i := fcext.RandSenderPerDayN(ctx.Event.UserID, len(cardMap))
			card := cardMap[(strconv.Itoa(i))]

			if id := ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text("今天的守护宝可梦是~【", card.Name, "】哒"),
				message.Image("file:///"+datapath+card.URL),
				// message.Text("\n【", card.Lines[rand.Intn(len(card.Lines))], "】"),
			); id.ID() == 0 {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("今天的守护宝可梦是~【", card.Name, "】哒\n【图片发送失败，请联系维护者~】"))
			}
		})
	engine.OnRegex(`^看宝可梦\s?(.*)`, getWife).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		info, ok := infoMap[match]
		if ok {
			if id := ctx.SendChain(
				message.Text("【", info.Name, "】"),
				message.Image("file:///"+datapath+info.URL),
			); id.ID() == 0 {
				ctx.SendChain(
					message.Text("【", info.Name, "】\n【图片发送失败，请联系维护者~】"))
				return
			}
			ctx.SendChain(
				message.Text("没有找到", match, "噢……"))
		}
	})
}
