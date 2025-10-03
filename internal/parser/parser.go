package parser

import (
	"os"

	ilog "github.com/hx-w/minidemo-encoder/internal/logger"
	dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
)

type TickPlayer struct {
	tick    int
	steamid uint64
}

func Start(filePath string) {
	iFile, err := os.Open(filePath)
	checkError(err)

	iParser := dem.NewParser(iFile)
	defer iParser.Close()

	// 先解析头部以获取地图信息
	header, err := iParser.ParseHeader()
	checkError(err)
	ilog.InfoLogger.Printf("========================================")
	ilog.InfoLogger.Printf("地图: %s", header.MapName)
	ilog.InfoLogger.Printf("========================================")

	// 处理特殊event构成的button表示
	var buttonTickMap map[TickPlayer]int32 = make(map[TickPlayer]int32)
	var (
		roundStarted = 0
		roundNum     = 0
	)

	iParser.RegisterEventHandler(func(e events.FrameDone) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()

		// 移除 roundInFreezetime 判断，从买枪阶段就开始录制
		if roundNum > 0 { // 只要回合已经开始就记录
			tPlayers := gs.TeamTerrorists().Members()
			ctPlayers := gs.TeamCounterTerrorists().Members()
			Players := append(tPlayers, ctPlayers...)
			for _, player := range Players {
				if player != nil {
					var addonButton int32 = 0
					key := TickPlayer{currentTick, player.SteamID64}
					if val, ok := buttonTickMap[key]; ok {
						addonButton = val
						delete(buttonTickMap, key)
					}
					parsePlayerFrame(player, addonButton, iParser.TickRate(), false)
				}
			}
		}
	})

	var weaponFireCount int = 0
	iParser.RegisterEventHandler(func(e events.WeaponFire) {
		if e.Shooter == nil {
			return
		}
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		key := TickPlayer{currentTick, e.Shooter.SteamID64}
		if _, ok := buttonTickMap[key]; ok {
			buttonTickMap[key] |= IN_ATTACK
		} else {
			buttonTickMap[key] = IN_ATTACK
		}
		weaponFireCount++
		if weaponFireCount%100 == 0 {
			ilog.InfoLogger.Printf("已记录 %d 次开枪事件\n", weaponFireCount)
		}
	})

	iParser.RegisterEventHandler(func(e events.PlayerJump) {
		if e.Player == nil {
			return
		}
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		key := TickPlayer{currentTick, e.Player.SteamID64}
		if _, ok := buttonTickMap[key]; ok {
			buttonTickMap[key] |= IN_JUMP
		} else {
			buttonTickMap[key] = IN_JUMP
		}
	})

	// 包括开局准备时间（买枪阶段）
	iParser.RegisterEventHandler(func(e events.RoundStart) {
		roundStarted = 1
		roundNum += 1
		ilog.InfoLogger.Printf("回合 %d 开始（包括买枪阶段）\n", roundNum)
		
		// 在买枪阶段开始时初始化录像文件
		// 写入所有选手的初始位置和角度
		gs := iParser.GameState()
		tPlayers := gs.TeamTerrorists().Members()
		ctPlayers := gs.TeamCounterTerrorists().Members()
		Players := append(tPlayers, ctPlayers...)
		for _, player := range Players {
			if player != nil {
				// parse player - 记录买枪阶段开始时的位置
				parsePlayerInitFrame(player)
			}
		}
	})

	// 准备时间结束，正式开始
	iParser.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		ilog.InfoLogger.Printf("回合 %d 买枪阶段结束，正式开始\n", roundNum)
	})

	// 回合结束，不包括自由活动时间
	iParser.RegisterEventHandler(func(e events.RoundEnd) {
		if roundStarted == 0 {
			roundStarted = 1
			roundNum = 0
		}
		ilog.InfoLogger.Println("回合结束：", roundNum)
		// 结束录像文件
		gs := iParser.GameState()
		tPlayers := gs.TeamTerrorists().Members()
		ctPlayers := gs.TeamCounterTerrorists().Members()
		Players := append(tPlayers, ctPlayers...)
		ilog.InfoLogger.Printf("回合%d结束，共%d名选手\n", roundNum, len(Players))
		for _, player := range Players {
			if player != nil {
				// save to rec file
				saveToRecFile(player, int32(roundNum))
			}
		}
	})
	err = iParser.ParseToEnd()
	checkError(err)
}
