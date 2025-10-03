package parser

import (
	"os"
	"fmt"

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

	// 处理特殊event构成的button表示
	var buttonTickMap map[TickPlayer]int32 = make(map[TickPlayer]int32)
	var firstFrameFullsnap map[uint64]bool = make(map[uint64]bool)
	var (
		roundStarted        = 0
		roundNum            = 0
		recording           = 0
		pendingRoundFolder string
	)

	// 监听游戏开始事件，确保从准备阶段就开始录制
	iParser.RegisterEventHandler(func(e events.MatchStart) {
		ilog.InfoLogger.Println("比赛开始，开始录制")
		recording = 1
	})

	iParser.RegisterEventHandler(func(e events.FrameDone) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()

		// 如果还未开始录制但检测到玩家，自动开始录制（适用于没有MatchStart事件的demo）
		if recording == 0 {
			tPlayers := gs.TeamTerrorists().Members()
			ctPlayers := gs.TeamCounterTerrorists().Members()
			Players := append(tPlayers, ctPlayers...)
			for _, player := range Players {
				if player != nil && player.IsAlive() {
					ilog.InfoLogger.Println("检测到玩家，自动开始录制")
					recording = 1
					break
				}
			}
		}

		if recording == 1 { // record during freezetime, active round, and post-round gap
			tPlayers := gs.TeamTerrorists().Members()
			ctPlayers := gs.TeamCounterTerrorists().Members()
			Players := append(tPlayers, ctPlayers...)
			for _, player := range Players {
				if player != nil {
					// 检查玩家是否已初始化，如果没有则初始化（确保捕获准备阶段早期行为）
					if _, exists := playerLastZ[player.Name]; !exists {
						parsePlayerInitFrame(player)
						firstFrameFullsnap[player.SteamID64] = true
					}
					
					var addonButton int32 = 0
					key := TickPlayer{currentTick, player.SteamID64}
					if val, ok := buttonTickMap[key]; ok {
						addonButton = val
						delete(buttonTickMap, key)
					}
					fullsnap := false
					if firstFrameFullsnap[player.SteamID64] {
						fullsnap = true
						delete(firstFrameFullsnap, player.SteamID64)
					}
					parsePlayerFrame(player, addonButton, iParser.TickRate(), fullsnap)
				}
			}
		}
	})

	iParser.RegisterEventHandler(func(e events.WeaponFire) {
		// guard against nil shooter (can occur on certain events)
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
	})

	iParser.RegisterEventHandler(func(e events.PlayerJump) {
		// guard against nil player
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

	// 包括开局准备时间
	iParser.RegisterEventHandler(func(e events.RoundStart) {
		roundStarted = 1
		// If previous round is pending, flush it now (includes post-round gap)
		if pendingRoundFolder != "" {
			gs := iParser.GameState()
			tPlayers := gs.TeamTerrorists().Members()
			ctPlayers := gs.TeamCounterTerrorists().Members()
			Players := append(tPlayers, ctPlayers...)
			for _, player := range Players {
				if player != nil {
					saveToRecFile(player, pendingRoundFolder)
				}
			}
			pendingRoundFolder = ""
		}
		recording = 1
		// Initialize recording buffers and write initial positions/angles at round start
		gs := iParser.GameState()
		tPlayers := gs.TeamTerrorists().Members()
		ctPlayers := gs.TeamCounterTerrorists().Members()
		Players := append(tPlayers, ctPlayers...)
		for _, player := range Players {
			if player != nil {
				// 重新初始化玩家（新回合开始）
				parsePlayerInitFrame(player)
				firstFrameFullsnap[player.SteamID64] = true
			}
		}
	})

	// 准备时间结束，正式开始
	iParser.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		roundNum += 1
		ilog.InfoLogger.Println("回合开始：", roundNum)
	})

	// 回合结束，不包括自由活动时间
	iParser.RegisterEventHandler(func(e events.RoundEnd) {
		if roundStarted == 0 {
			roundStarted = 1
			roundNum = 0
		}
		// Keep recording to include post-round gap
		ilog.InfoLogger.Println("回合结束：", roundNum)
		gs := iParser.GameState()
		tPlayers := gs.TeamTerrorists().Members()
		ctPlayers := gs.TeamCounterTerrorists().Members()
		Players := append(tPlayers, ctPlayers...)
		tScore := gs.TeamTerrorists().Score()
		ctScore := gs.TeamCounterTerrorists().Score()
		pendingRoundFolder = fmt.Sprintf("round%d_T%d-CT%d", roundNum, tScore, ctScore)
		ilog.InfoLogger.Printf("回合%d结束，共%d名选手，暂存目录：%s\n", roundNum, len(Players), pendingRoundFolder)
	})
	err = iParser.ParseToEnd()
	checkError(err)
	// Final flush if demo ends without another RoundStart
	if pendingRoundFolder != "" {
		gs := iParser.GameState()
		tPlayers := gs.TeamTerrorists().Members()
		ctPlayers := gs.TeamCounterTerrorists().Members()
		Players := append(tPlayers, ctPlayers...)
		for _, player := range Players {
			if player != nil {
				saveToRecFile(player, pendingRoundFolder)
			}
		}
		pendingRoundFolder = ""
	}
}


