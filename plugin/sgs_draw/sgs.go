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

var cardMap = make(cardSet, 50)
var datapath string

func init() {
	engine := control.Register("sgs", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "三国杀\n" +
			"- 抽武将",
		PublicDataFolder: "Sgs",
	}).ApplySingle(ctxext.DefaultSingle)
	engine.OnFullMatchGroup([]string{"抽武将"}, fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
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
			datapath = file.BOTPATH + "/" + engine.DataFolder()
			logrus.Infof("[sgs]读取%d张三国杀武将", len(cardMap))
			return true
		},
	)).SetBlock(true).
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
}
