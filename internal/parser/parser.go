package parser

import (
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"strconv"

	encoder "github.com/hx-w/minidemo-encoder/internal/encoder"
	ilog "github.com/hx-w/minidemo-encoder/internal/logger"
	dem "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
)

type TickPlayer struct {
	tick    int
	steamid uint64
}

func Start(filePath string, skipFreezetime bool) {
	iFile, err := os.Open(filePath)
	checkError(err)

	iParser := dem.NewParser(iFile)
	defer iParser.Close()

	// Extract demo name from file path
	demoName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	var subdirSet bool = false  // Flag to track if output subdir is set
	
	// Reset chat messages for new demo
	ResetChatMessages()

	// Handle special events for button representation
	var buttonTickMap map[TickPlayer]int32 = make(map[TickPlayer]int32)
	var firstFrameFullsnap map[uint64]bool = make(map[uint64]bool)
	var (
		roundStarted        = 0
		roundNum            = 0
		recording           = 0
		pendingRoundFolder string
		demoTickrate       float64  // Store demo tickrate
	)
	
	if skipFreezetime {
		ilog.InfoLogger.Println("Skip freezetime mode enabled")
	}

	// Listen for match start event to ensure recording from preparation phase
	iParser.RegisterEventHandler(func(e events.MatchStart) {
		ilog.InfoLogger.Println("Match started")
		if !skipFreezetime {
			ilog.InfoLogger.Println("Recording begins")
			recording = 1
		}
	})

	iParser.RegisterEventHandler(func(e events.FrameDone) {
		// Set output subdir on first frame (tickrate is available at this point)
		if !subdirSet {
			demoTickrate = iParser.TickRate()
			rateStr := strconv.FormatFloat(demoTickrate, 'f', -1, 64)
			encoder.SetOutputSubDir(rateStr + demoName)
			subdirSet = true
		}

		gs := iParser.GameState()
		currentTick := gs.IngameTick()

		// Auto-start recording if players detected (for demos without MatchStart event)
		// When skipFreezetime is enabled, this auto-start is disabled
		if recording == 0 && !skipFreezetime {
			tPlayers := gs.TeamTerrorists().Members()
			ctPlayers := gs.TeamCounterTerrorists().Members()
			Players := append(tPlayers, ctPlayers...)
			for _, player := range Players {
				if player != nil && player.IsAlive() {
					ilog.InfoLogger.Println("Player detected, auto-start recording")
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
					// Check if player is initialized, if not initialize (ensure early preparation phase capture)
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
			flushRecordedPlayers(pendingRoundFolder)
			pendingRoundFolder = ""
		}
		
		// If skipFreezetime is enabled, don't start recording yet
		if !skipFreezetime {
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
		}
	})

	// 准备时间结束，正式开始
	iParser.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		roundNum += 1
		ilog.InfoLogger.Println("Round started:", roundNum)
		
		// 更新聊天消息的当前回合数
		SetCurrentChatRound(roundNum)
		
		// If skipFreezetime is enabled, start recording now
		if skipFreezetime {
			recording = 1
			gs := iParser.GameState()
			tPlayers := gs.TeamTerrorists().Members()
			ctPlayers := gs.TeamCounterTerrorists().Members()
			Players := append(tPlayers, ctPlayers...)
			for _, player := range Players {
				if player != nil {
					// Initialize players at freezetime end
					parsePlayerInitFrame(player)
					firstFrameFullsnap[player.SteamID64] = true
				}
			}
			ilog.InfoLogger.Println("Recording started after freezetime")
		}
	})

	// 回合结束，不包括自由活动时间
	iParser.RegisterEventHandler(func(e events.RoundEnd) {
		if roundStarted == 0 {
			roundStarted = 1
			roundNum = 0
		}
		// Keep recording to include post-round gap
		ilog.InfoLogger.Println("Round ended:", roundNum)
		gs := iParser.GameState()
		tScore := gs.TeamTerrorists().Score()
		ctScore := gs.TeamCounterTerrorists().Score()
		pendingRoundFolder = fmt.Sprintf("round%d_T%d-CT%d", roundNum, tScore, ctScore)
		
		// Count recorded players
		recordedCount := len(encoder.PlayerFramesMap)
		ilog.InfoLogger.Printf("Round %d ended, %d players recorded, pending folder: %s\n", roundNum, recordedCount, pendingRoundFolder)
	})

	// 聊天消息处理（ChatMessage 事件）
	iParser.RegisterEventHandler(func(e events.ChatMessage) {
		gs := iParser.GameState()
		currentTick := gs.IngameTick()
		
		// 获取发送者信息
		sender := "Unknown"
		team := "Unknown"
		
		// 尝试从发送者获取信息
		if e.Sender != nil {
			sender = e.Sender.Name
			switch e.Sender.Team {
			case 2:
				team = "T"
			case 3:
				team = "CT"
			default:
				team = "Spectator"
			}
		}
		
		// 判断是否为团队聊天（IsChatAll=false 表示团队聊天）
		isTeamChat := !e.IsChatAll
		
		// 添加聊天消息
		AddChatMessage(currentTick, sender, team, e.Text, isTeamChat)
	})

	err = iParser.ParseToEnd()
	checkError(err)
	// Final flush if demo ends without another RoundStart
	if pendingRoundFolder != "" {
		flushRecordedPlayers(pendingRoundFolder)
		pendingRoundFolder = ""
	}

	// 写入 tickrate 记事本（放在 demo 输出根目录下）
	encoder.WriteTickrateNoteFile(demoTickrate)
	
	// 保存聊天消息到文件
	outputDir := "./output"
	if encoder.GetOutputSubDir() != "" {
		outputDir = filepath.Join("./output", encoder.GetOutputSubDir())
	}
	
	// 保存为 TXT 格式（完整聊天记录）
	if err := SaveChatMessages(outputDir, demoName); err != nil {
		ilog.InfoLogger.Printf("Failed to save chat messages: %v\n", err)
	}
	
	// 保存为 CSV 格式（完整聊天记录）
	if err := SaveChatMessagesCSV(outputDir, demoName); err != nil {
		ilog.InfoLogger.Printf("Failed to save chat messages CSV: %v\n", err)
	}
	
	// 按回合保存聊天消息（用于游戏中复现）
	if err := SaveChatMessagesByRound(outputDir, ""); err != nil {
		ilog.InfoLogger.Printf("Failed to save chat messages by round: %v\n", err)
	}
}



