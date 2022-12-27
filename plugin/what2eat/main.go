// Package what2eat 今天吃什么
// 数据来自https://github.com/MinatoAquaCrews/nonebot_plugin_what2eat 非常感谢！
package what2eat

import (
	"encoding/json"
	"math/rand"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	foodMap       map[string][]string
	drinkMap      map[string][]string
	drinkNameList []string
	introList     = [...]string{"根据大数据测算，您", "通过我高性能电子智能的分析，您", "魔法水晶球的揭示告诉我，您"}
)

func init() {
	engine := control.Register("what2eat", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "吃什么\n" +
			"- [今明后]天[早上|中午|晚上]吃什么\n" +
			"- [今明后]天[早上|中午|晚上]喝什么",
		PublicDataFolder: "What2eat",
	}).ApplySingle(ctxext.DefaultSingle)
	getWhat2eat := fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			data, err := engine.GetLazyData("eating.json", false)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = json.Unmarshal(data, &foodMap)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			logrus.Infof("[what2eat]读取%d个食物", len(foodMap["basic_food"]))
			data, err = engine.GetLazyData("drinks.json", false)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = json.Unmarshal(data, &drinkMap)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			for k := range drinkMap {
				drinkNameList = append(drinkNameList, k)
			}
			logrus.Infof("[what2eat]读取%d家饮料", len(drinkNameList))
			return true
		})
	engine.OnRegex(`^(大?[今明后]天)?([早中午晚][上饭餐午]|早上|夜宵|今晚)?吃(什么|啥|点啥)`, getWhat2eat).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			eatDay := ctx.State["regex_matched"].([]string)[1]
			eatTime := ctx.State["regex_matched"].([]string)[2]
			i := rand.Intn(len(foodMap["basic_food"]))
			food := foodMap["basic_food"][i]
			intro := introList[rand.Intn(len(introList))]
			if id := ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text(intro, eatDay, eatTime, "应该吃:【", food, "】"),
				// 或许应该搞点图
			); id.ID() == 0 {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text(intro, eatDay, eatTime, "应该吃:【", food, "】"))
			}
		})
	engine.OnRegex(`^(大?[今明后]天)?([早中午晚][上饭餐午]|早上|夜宵|今晚)?喝(什么|啥|点啥)`, getWhat2eat).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			eatDay := ctx.State["regex_matched"].([]string)[1]
			eatTime := ctx.State["regex_matched"].([]string)[2]
			key := drinkNameList[rand.Intn(len(drinkNameList))]
			drink := drinkMap[key][rand.Intn(len(drinkMap[key]))]
			intro := introList[rand.Intn(len(introList))]
			if id := ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text(intro, eatDay, eatTime, "应该喝【", key, "】的【", drink, "】"),
				// 或许应该搞点图
			); id.ID() == 0 {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text(intro, eatDay, eatTime, "应该喝【", key, "】的【", drink, "】"))
			}
		})
}
