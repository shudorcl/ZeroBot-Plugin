// Package coc coc插件
package coc

import (
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/file"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnPrefixGroup([]string{".set", "。set", ".SET"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		infoFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + DefaultYamlFile
		if file.IsNotExist(infoFile) {
			ctx.SendChain(message.Text("你群还没有布置coc,请相关人员后台布置coc.(详情看用法)"))
			return
		}
		if file.IsNotExist(engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + strconv.FormatInt(uid, 10) + ".yml") {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你还没有创建角色"))
			return
		}

		cocInfo, err := loadPanel(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		baseMsg := strings.Split(strings.TrimSpace(ctx.State["args"].(string)), "/")
		if baseMsg == nil || len(baseMsg) < 1 {
			ctx.SendChain(message.Text("[ERROR setting.go.37]:参数错误"))
			return
		}
		change := false
		for _, msgInfo := range baseMsg {
			msgValue := strings.Split(msgInfo, "#")
			if msgValue[0] == "描述" {
				cocInfo.Other = append(cocInfo.Other, msgValue[1])
				continue
			}
			for i, info := range cocInfo.BaseInfo {
				if msgValue[0] == info.Name {
					change = true
					munberValue, err := strconv.Atoi(msgValue[1])
					if err != nil {
						cocInfo.BaseInfo[i].Value = msgValue[1]
					} else {
						cocInfo.BaseInfo[i].Value = munberValue
					}
				}
			}
		}
		if change {
			err = savePanel(cocInfo, gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("更改内容不匹配"))
	})
	engine.OnPrefixGroup([]string{".sst", "。sst", ".SST"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		settingMsg := strings.Split(strings.TrimSpace(ctx.State["args"].(string)), " ")
		if len(settingMsg) < 3 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("参数错误"))
			return
		}
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		infoFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + DefaultYamlFile
		if file.IsNotExist(infoFile) {
			ctx.SendChain(message.Text("你群还没有布置coc,请相关人员后台布局coc.(详情看用法)"))
			return
		}
		if file.IsNotExist(engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + strconv.FormatInt(uid, 10) + ".yml") {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你还没有创建角色"))
			return
		}

		cocInfo, err := loadPanel(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		switch settingMsg[0] {
		case "add":
			newOther := make([]string, 0, len(cocInfo.Other)+1)
			site, err := strconv.Atoi(settingMsg[1])
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("参数错误"))
				return
			}
			newOther = append(newOther, cocInfo.Other[:site]...)
			newOther = append(newOther, settingMsg[2])
			newOther = append(newOther, cocInfo.Other[site:]...)
			cocInfo.Other = newOther
		case "del":
			newOther := make([]string, 0, len(cocInfo.Other)-1)
			site, err := strconv.Atoi(settingMsg[1])
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("参数错误"))
				return
			}
			newOther = append(newOther, cocInfo.Other[:site]...)
			newOther = append(newOther, cocInfo.Other[site+1:]...)
			cocInfo.Other = newOther
		case "clr":
			site, err := strconv.Atoi(settingMsg[1])
			if err != nil || site <= 0 || site > len(cocInfo.Other)-1 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("参数错误"))
				return
			}
			cocInfo.Other[site-1] = settingMsg[2]
		}
		err = savePanel(cocInfo, gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
	})
	engine.OnRegex(`^(\.|。)(s|S)(a|A) ([1-9]\d*)?(d|D)([1-9]\d*)?(a(\S+))? (\S+) ((-|\+)?[1-9]\d*)(\s+([1-9]\d*))?$`, getsetting).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		mu.Lock()
		cocSetting := settingGoup[gid]
		mu.Unlock()
		uid := ctx.Event.UserID
		infoFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + DefaultYamlFile
		if file.IsNotExist(infoFile) {
			ctx.SendChain(message.Text("你群还没有布置coc,请相关人员后台布局coc.(详情看用法)"))
			return
		}
		if file.IsNotExist(engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + strconv.FormatInt(uid, 10) + ".yml") {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你还没有创建角色"))
			return
		}

		cocInfo, err := loadPanel(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}

		var (
			times       = 1    // 4
			defaultDice = 100  // 6
			limit       = -1   // 8 -> name
			atrr        string // 9 -> name
			money       = 0
			success     = false
		)

		if ctx.State["regex_matched"].([]string)[4] != "" {
			times, err = strconv.Atoi(ctx.State["regex_matched"].([]string)[4])
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:骰子次数参数错误"))
				return
			}
		}

		if ctx.State["regex_matched"].([]string)[6] == "" && cocSetting.DefaultDice != 0 {
			defaultDice = cocSetting.DefaultDice
		} else if ctx.State["regex_matched"].([]string)[6] != "" {
			defaultDice, err = strconv.Atoi(ctx.State["regex_matched"].([]string)[6])
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:骰子面数参数错误"))
				return
			}
		}

		if ctx.State["regex_matched"].([]string)[8] == "" {
			limit = 50
		} else {
			for _, info := range cocInfo.Attribute {
				if ctx.State["regex_matched"].([]string)[8] == info.Name {
					limit = defaultDice * (info.Value - info.MinValue) / (info.MaxValue - info.MinValue)
				}
			}
		}
		if limit == -1 {
			ctx.SendChain(message.Text("[ERROR]:参数错误"))
			return
		}

		if ctx.State["regex_matched"].([]string)[9] == "" {
			ctx.SendChain(message.Text("[ERROR]:参数错误"))
			return
		}
		for _, info := range cocInfo.Attribute {
			if ctx.State["regex_matched"].([]string)[9] == info.Name {
				atrr = info.Name
			}
		}
		if atrr == "" {
			ctx.SendChain(message.Text("[ERROR]:参数错误"))
			return
		}

		atrrValue, err := strconv.Atoi(ctx.State["regex_matched"].([]string)[10])
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:属性增加值参数错误"))
			return
		}

		if ctx.State["regex_matched"].([]string)[13] != "" {
			money, err = strconv.Atoi(ctx.State["regex_matched"].([]string)[9])
			if err != nil || money < 0 {
				ctx.SendChain(message.Text("[ERROR]:金钱参数错误"))
				return
			}
			err = wallet.InsertWalletOf(uid, -money)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
		}

		msg := make(message.Message, 0, 3+times*2)
		msg = append(msg, message.Reply(ctx.Event.MessageID))
		msg = append(msg, message.Text("对[", atrr, "增幅", atrrValue, "]事件\n使用了金钱", money, "和", ctx.State["regex_matched"].([]string)[8], "为权重进行了\n"))
		sum := 0
		result := ""
		for i := times; i > 0; i-- {
			dice := rand.Intn(defaultDice) + 1
			result = diceRule(cocSetting.DiceRule, dice, limit, defaultDice)
			msg = append(msg, message.Text("🎲 => ", dice, result, "\n"))
			sum += dice
		}
		if times > 1 {
			result = diceRule(cocSetting.DiceRule, sum, limit*times, defaultDice*times)
			msg = append(msg, message.Text("合计=", sum, result))
		}
		if strings.Contains(result, "成功") {
			success = true
		}
		if success {
			for i, info := range cocInfo.Attribute {
				if atrr == info.Name {
					cocInfo.Attribute[i].Value += atrrValue
					if cocInfo.Attribute[i].Value > cocInfo.Attribute[i].MaxValue {
						cocInfo.Attribute[i].Value = cocInfo.Attribute[i].MaxValue
					}
				}
			}
			err = savePanel(cocInfo, gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
		}
		ctx.Send(msg)
	})
	engine.OnRegex(`^(\.|。)(pc|PC)(\[CQ:at,qq=)?(\d+)(\])?`, getsetting, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		mu.Lock()
		cocSetting := settingGoup[gid]
		mu.Unlock()
		uid, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[4], 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		cocSetting.CocPC = uid
		if cocSetting.DefaultDice == 0 {
			cocSetting.DefaultDice = 100
		}
		err = saveSetting(cocSetting, gid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
	})
	engine.OnPrefixGroup([]string{".dice", "。dice", ".DICE"}, getsetting, func(ctx *zero.Ctx) bool {
		mu.Lock()
		cocSetting := settingGoup[ctx.Event.GroupID]
		mu.Unlock()
		if cocSetting.CocPC == 0 {
			return zero.AdminPermission(ctx)
		} else if cocSetting.CocPC != 0 && ctx.Event.UserID != cocSetting.CocPC {
			ctx.SendChain(message.Text("[ERROR]:已指定PC,无权更改"))
			return false
		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		mu.Lock()
		cocSetting := settingGoup[gid]
		mu.Unlock()
		defaultDice, err := strconv.Atoi(strings.TrimSpace(ctx.State["args"].(string)))
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		cocSetting.DefaultDice = defaultDice
		err = saveSetting(cocSetting, gid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
	})
	engine.OnPrefixGroup([]string{".rule", "。rule", ".RULE"}, getsetting, func(ctx *zero.Ctx) bool {
		mu.Lock()
		cocSetting := settingGoup[ctx.Event.GroupID]
		mu.Unlock()
		if cocSetting.CocPC == 0 {
			return zero.AdminPermission(ctx)
		} else if cocSetting.CocPC != 0 && ctx.Event.UserID != cocSetting.CocPC {
			ctx.SendChain(message.Text("[ERROR]:已指定PC,无权更改"))
			return false
		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		mu.Lock()
		cocSetting := settingGoup[gid]
		mu.Unlock()
		defaultRule, err := strconv.Atoi(strings.TrimSpace(ctx.State["args"].(string)))
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		cocSetting.DiceRule = defaultRule
		err = saveSetting(cocSetting, gid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
	})
	// engine.OnRegex(`^(\.|。)(show|SHOW)\s*(\[CQ:at,qq=)?(\d+)(\])?`, getsetting, func(ctx *zero.Ctx) bool {
	// 	mu.Lock()
	// 	cocSetting := settingGoup[ctx.Event.GroupID]
	// 	mu.Unlock()
	// 	if cocSetting.CocPC == 0 {
	// 		return zero.AdminPermission(ctx)
	// 	} else if cocSetting.CocPC != 0 && ctx.Event.UserID != cocSetting.CocPC {
	// 		ctx.SendChain(message.Text("[ERROR]:已指定PC,无权操作"))
	// 		return false
	// 	}
	// 	return true
	// }).SetBlock(true).Handle(func(ctx *zero.Ctx) {
	// 	gid := ctx.Event.GroupID
	// 	uid, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[4], 10, 64)
	// 	if err != nil {
	// 		ctx.SendChain(message.Text("[ERROR]:", err))
	// 		return
	// 	}
	// 	if file.IsNotExist(engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + strconv.FormatInt(uid, 10) + ".yml") {
	// 		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("对方还没有创建角色"))
	// 		return
	// 	}
	// 	cocInfo, err := loadPanel(gid, uid)
	// 	if err != nil {
	// 		ctx.SendChain(message.Text("[ERROR]:", err))
	// 		return
	// 	}
	// 	pic, err := drawImage(cocInfo)
	// 	if err != nil {
	// 		ctx.SendChain(message.Text("[ERROR]:", err))
	// 		return
	// 	}
	// 	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.ImageBytes(pic))
	// })
	engine.OnRegex(`^(\.|。)(kill|KILL)(\[CQ:at,qq=)?(\d+)?(\])?`, getsetting, func(ctx *zero.Ctx) bool {
		mu.Lock()
		cocSetting := settingGoup[ctx.Event.GroupID]
		mu.Unlock()
		if cocSetting.CocPC == 0 {
			return zero.AdminPermission(ctx)
		} else if cocSetting.CocPC != 0 && ctx.Event.UserID != cocSetting.CocPC {
			ctx.SendChain(message.Text("[ERROR]:已指定PC,无权操作"))
			return false
		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[4], 10, 64)
		if err != nil || uid == 0 {
			uid = ctx.Event.UserID
		}
		infoFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + strconv.FormatInt(uid, 10) + ".yml"
		if file.IsNotExist(infoFile) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("还没有创建角色"))
			return
		}
		err = os.RemoveAll(infoFile)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
	})
	engine.OnRegex(`^(\.|。)(pcset|PCSET)\s*(\[CQ:at,qq=)?(\d+)(\])?\s*(.*)$`, getsetting, func(ctx *zero.Ctx) bool {
		mu.Lock()
		cocSetting := settingGoup[ctx.Event.GroupID]
		mu.Unlock()
		if cocSetting.CocPC == 0 {
			return zero.AdminPermission(ctx)
		} else if cocSetting.CocPC != 0 && ctx.Event.UserID != cocSetting.CocPC {
			ctx.SendChain(message.Text("[ERROR]:已指定PC,无权操作"))
			return false
		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[4], 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if file.IsNotExist(engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + strconv.FormatInt(uid, 10) + ".yml") {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("对方还没有创建角色"))
			return
		}
		cocInfo, err := loadPanel(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		baseMsg := strings.Split(ctx.State["regex_matched"].([]string)[6], "/")
		change := false
		for _, msgInfo := range baseMsg {
			msgValue := strings.Split(msgInfo, "#")
			if msgValue[0] == "描述" {
				cocInfo.Other = append(cocInfo.Other, msgValue[1])
				continue
			}
			for i, info := range cocInfo.BaseInfo {
				if msgValue[0] == info.Name {
					change = true
					munberValue, err := strconv.Atoi(msgValue[1])
					if err != nil {
						cocInfo.BaseInfo[i].Value = msgValue[1]
					} else {
						cocInfo.BaseInfo[i].Value = munberValue
					}
				}
			}
			if change {
				break
			}
			for i, info := range cocInfo.Attribute {
				if msgValue[0] == info.Name {
					change = true
					munberValue, err := strconv.Atoi(msgValue[1])
					if err != nil {
						ctx.SendChain(message.Text("[ERROR]:", err))
						return
					}
					cocInfo.Attribute[i].Value = munberValue
				}
			}
			if change {
				break
			}
		}
		if change {
			err = savePanel(cocInfo, gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("无数据匹配"))
	})
}
