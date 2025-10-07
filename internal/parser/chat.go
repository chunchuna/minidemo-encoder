package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	ilog "github.com/hx-w/minidemo-encoder/internal/logger"
)

type ChatMessage struct {
	Tick        int
	RoundNum    int
	Sender      string
	SenderTeam  string
	Text        string
	IsTeamChat  bool
	Timestamp   time.Time
}

var chatMessages []ChatMessage
var currentRoundForChat int = 0

// ResetChatMessages 重置聊天消息列表（批量解析时需要）
func ResetChatMessages() {
	chatMessages = []ChatMessage{}
	currentRoundForChat = 0
}

// SetCurrentChatRound 设置当前回合数（用于标记聊天消息属于哪个回合）
func SetCurrentChatRound(roundNum int) {
	currentRoundForChat = roundNum
}

// AddChatMessage 添加聊天消息到列表
func AddChatMessage(tick int, sender string, team string, text string, isTeamChat bool) {
	msg := ChatMessage{
		Tick:       tick,
		RoundNum:   currentRoundForChat,
		Sender:     sender,
		SenderTeam: team,
		Text:       text,
		IsTeamChat: isTeamChat,
		Timestamp:  time.Now(),
	}
	chatMessages = append(chatMessages, msg)
	
	// 日志输出
	chatType := "All"
	if isTeamChat {
		chatType = "Team"
	}
	ilog.InfoLogger.Printf("[Chat][Round %d][%s] %s (%s): %s\n", currentRoundForChat, chatType, sender, team, text)
}

// SaveChatMessages 保存聊天消息到文件
func SaveChatMessages(outputDir string, demoName string) error {
	if len(chatMessages) == 0 {
		ilog.InfoLogger.Println("No chat messages to save")
		return nil
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 创建聊天记录文件
	chatFilePath := filepath.Join(outputDir, fmt.Sprintf("%s_chat.txt", demoName))
	file, err := os.Create(chatFilePath)
	if err != nil {
		return fmt.Errorf("failed to create chat file: %v", err)
	}
	defer file.Close()

	// 写入标题
	file.WriteString(fmt.Sprintf("=== Chat Messages from Demo: %s ===\n", demoName))
	file.WriteString(fmt.Sprintf("Total Messages: %d\n", len(chatMessages)))
	file.WriteString(strings.Repeat("=", 60) + "\n\n")

	// 写入每条聊天消息
	for i, msg := range chatMessages {
		chatType := "All Chat"
		if msg.IsTeamChat {
			chatType = "Team Chat"
		}
		
		file.WriteString(fmt.Sprintf("[%d] Tick: %d\n", i+1, msg.Tick))
		file.WriteString(fmt.Sprintf("    Type: %s\n", chatType))
		file.WriteString(fmt.Sprintf("    Sender: %s (Team: %s)\n", msg.Sender, msg.SenderTeam))
		file.WriteString(fmt.Sprintf("    Message: %s\n", msg.Text))
		file.WriteString("\n")
	}

	ilog.InfoLogger.Printf("Chat messages saved to: %s (Total: %d messages)\n", chatFilePath, len(chatMessages))
	return nil
}

// SaveChatMessagesCSV 保存聊天消息到 CSV 文件（可选的格式）
func SaveChatMessagesCSV(outputDir string, demoName string) error {
	if len(chatMessages) == 0 {
		return nil
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 创建 CSV 文件
	csvFilePath := filepath.Join(outputDir, fmt.Sprintf("%s_chat.csv", demoName))
	file, err := os.Create(csvFilePath)
	if err != nil {
		return fmt.Errorf("failed to create csv file: %v", err)
	}
	defer file.Close()

	// 写入 CSV 标题行
	file.WriteString("Tick,ChatType,Sender,Team,Message\n")

	// 写入每条聊天消息
	for _, msg := range chatMessages {
		chatType := "All"
		if msg.IsTeamChat {
			chatType = "Team"
		}
		
		// 转义 CSV 中的特殊字符
		text := strings.ReplaceAll(msg.Text, "\"", "\"\"")
		sender := strings.ReplaceAll(msg.Sender, "\"", "\"\"")
		
		file.WriteString(fmt.Sprintf("%d,%s,\"%s\",%s,\"%s\"\n", 
			msg.Tick, chatType, sender, msg.SenderTeam, text))
	}

	ilog.InfoLogger.Printf("Chat messages saved to CSV: %s\n", csvFilePath)
	return nil
}

// SaveChatMessagesByRound 按回合保存聊天消息（用于回放时重现）
func SaveChatMessagesByRound(outputDir string, roundFolder string) error {
	if len(chatMessages) == 0 {
		return nil
	}
	
	// 按回合分组聊天消息
	roundMessages := make(map[int][]ChatMessage)
	for _, msg := range chatMessages {
		roundMessages[msg.RoundNum] = append(roundMessages[msg.RoundNum], msg)
	}
	
	// 为每个回合保存聊天记录
	for roundNum, messages := range roundMessages {
		if len(messages) == 0 {
			continue
		}
		
		// 构建回合文件夹路径（如果指定了 roundFolder）
		var chatFilePath string
		if roundFolder != "" {
			// 创建回合文件夹（如果不存在）
			roundPath := filepath.Join(outputDir, roundFolder)
			if err := os.MkdirAll(roundPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create round directory: %v", err)
			}
			chatFilePath = filepath.Join(roundPath, "chat.txt")
		} else {
			chatFilePath = filepath.Join(outputDir, fmt.Sprintf("round%d_chat.txt", roundNum))
		}
		
		// 创建聊天记录文件
		file, err := os.Create(chatFilePath)
		if err != nil {
			ilog.InfoLogger.Printf("Failed to create chat file for round %d: %v\n", roundNum, err)
			continue
		}
		
		// 写入标题
		file.WriteString(fmt.Sprintf("=== Chat Messages - Round %d ===\n", roundNum))
		file.WriteString(fmt.Sprintf("Total Messages: %d\n", len(messages)))
		file.WriteString(strings.Repeat("=", 60) + "\n\n")
		
		// 写入每条聊天消息
		for i, msg := range messages {
			chatType := "All"
			if msg.IsTeamChat {
				chatType = "Team"
			}
			
			file.WriteString(fmt.Sprintf("[%d] Tick: %d\n", i+1, msg.Tick))
			file.WriteString(fmt.Sprintf("    Type: %s\n", chatType))
			file.WriteString(fmt.Sprintf("    Sender: %s (Team: %s)\n", msg.Sender, msg.SenderTeam))
			file.WriteString(fmt.Sprintf("    Message: %s\n", msg.Text))
			file.WriteString("\n")
		}
		
		file.Close()
		ilog.InfoLogger.Printf("Saved %d chat messages for round %d to: %s\n", len(messages), roundNum, chatFilePath)
	}
	
	return nil
} 