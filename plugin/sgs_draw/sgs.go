// Package tarot 塔罗牌
package tarot

import (
	"encoding/json"
	"math/rand"
	"strconv"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
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

var cardMap = make(cardSet, 30)
var datapath string

func init() {
	engine := control.Register("sgs", &control.Options{
		DisableOnDefault: false,
		Help: "三国杀\n" +
			"- 抽武将",
		PublicDataFolder: "Sgs",
	}).ApplySingle(ctxext.DefaultSingle)
	engine.OnFullMatchGroup([]string{"抽武将"}, ctxext.DoOnceOnSuccess(
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
			i := ctxext.RandSenderPerDayN(ctx, len(cardMap))
			card := cardMap[(strconv.Itoa(i))]

			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text("今天的守护武将是~【", card.Name, "】哒"),
				message.Image("file:///"+datapath+card.URL),
				message.Text("\n【", card.Lines[rand.Intn(len(card.Lines))], "】"),
			)
		})
}
