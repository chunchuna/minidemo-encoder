package encoder

import (
	"bytes"
	"fmt"
	"os"
	"time"
	"strings"
	"regexp"
	"strconv"

	ilog "github.com/hx-w/minidemo-encoder/internal/logger"
)

const __MAGIC__ int32 = -559038737
const __FORMAT_VERSION__ int8 = 2
const FIELDS_ORIGIN int32 = 1 << 0
const FIELDS_ANGLES int32 = 1 << 1
const FIELDS_VELOCITY int32 = 1 << 2

var bufMap map[string]*bytes.Buffer = make(map[string]*bytes.Buffer)
var PlayerFramesMap map[string][]FrameInfo = make(map[string][]FrameInfo)

var saveDir string = "./output"
var outputSubDir string = ""

// SetOutputSubDir 设置输出子目录（用于按demo名称分类）
func SetOutputSubDir(subDir string) {
	outputSubDir = subDir
}

// ResetState 重置所有全局状态（批量解析时需要）
func ResetState() {
	bufMap = make(map[string]*bytes.Buffer)
	PlayerFramesMap = make(map[string][]FrameInfo)
}

func init() {
	if ok, _ := PathExists(saveDir); !ok {
		os.Mkdir(saveDir, os.ModePerm)
		ilog.InfoLogger.Println("Save directory not found, created:", saveDir)
	} else {
		ilog.InfoLogger.Println("Save directory exists:", saveDir)
	}
}

// sanitize file name for Windows: replace \\ / : * ? " < > | and trim spaces/dots
func sanitizeFileName(name string) string {
	// replace invalid characters with underscore
	invalidPattern := regexp.MustCompile(`[\\/:*?"<>|]`)
	safe := invalidPattern.ReplaceAllString(name, "_")
	// trim trailing spaces and dots which are not allowed in file names
	safe = strings.TrimRight(safe, " .")
	if len(safe) == 0 {
		return "player"
	}
	return safe
}

func InitPlayer(initFrame FrameInitInfo) {
	if bufMap[initFrame.PlayerName] == nil {
		bufMap[initFrame.PlayerName] = new(bytes.Buffer)
	} else {
		bufMap[initFrame.PlayerName].Reset()
	}
	// step.1 MAGIC NUMBER
	WriteToBuf(initFrame.PlayerName, __MAGIC__)

	// step.2 VERSION
	WriteToBuf(initFrame.PlayerName, __FORMAT_VERSION__)

	// step.3 timestamp
	WriteToBuf(initFrame.PlayerName, int32(time.Now().Unix()))

	// step.4 name length
	WriteToBuf(initFrame.PlayerName, int8(len(initFrame.PlayerName)))

	// step.5 name
	WriteToBuf(initFrame.PlayerName, []byte(initFrame.PlayerName))

	// step.6 initial position
	for idx := 0; idx < 3; idx++ {
		WriteToBuf(initFrame.PlayerName, float32(initFrame.Position[idx]))
	}

	// step.7 initial angle
	for idx := 0; idx < 2; idx++ {
		WriteToBuf(initFrame.PlayerName, initFrame.Angles[idx])
	}
	// ilog.InfoLogger.Println("初始化成功: ", initFrame.PlayerName)
}

func WriteToRecFile(playerName string, roundFolder string, subdir string) {
	// 构建完整路径：output/[demoName]/roundX_TY-CTZ/t(或ct)
	var fullPath string
	if outputSubDir != "" {
		fullPath = saveDir + "/" + outputSubDir + "/" + roundFolder + "/" + subdir
	} else {
		fullPath = saveDir + "/" + roundFolder + "/" + subdir
	}
	
	if ok, _ := PathExists(fullPath); !ok {
		os.MkdirAll(fullPath, os.ModePerm)
	}
	// sanitize file name for windows
	safeName := sanitizeFileName(playerName)
	fileName := fullPath + "/" + safeName + ".rec"
	file, err := os.Create(fileName)
	if err != nil {
		ilog.ErrorLogger.Println("Failed to create file:", err.Error())
		return
	}
	defer file.Close()

	// step.8 tick count
	var tickCount int32 = int32(len(PlayerFramesMap[playerName]))
	WriteToBuf(playerName, tickCount)

	// step.9 bookmark count
	WriteToBuf(playerName, int32(0))

	// step.10 all bookmark
	// ignore

	// step.11 all tick frame
	for _, frame := range PlayerFramesMap[playerName] {
		WriteToBuf(playerName, frame.PlayerButtons)
		WriteToBuf(playerName, frame.PlayerImpulse)
		for idx := 0; idx < 3; idx++ {
			WriteToBuf(playerName, frame.ActualVelocity[idx])
		}
		for idx := 0; idx < 3; idx++ {
			WriteToBuf(playerName, frame.PredictedVelocity[idx])
		}
		for idx := 0; idx < 2; idx++ {
			WriteToBuf(playerName, frame.PredictedAngles[idx])
		}
		// write per-frame origin right after predicted angles
		for idx := 0; idx < 3; idx++ {
			WriteToBuf(playerName, frame.Origin[idx])
		}
		WriteToBuf(playerName, frame.CSWeaponID)
		WriteToBuf(playerName, frame.PlayerSubtype)
		WriteToBuf(playerName, frame.PlayerSeed)
		WriteToBuf(playerName, frame.AdditionalFields)
		// 附加信息
		if frame.AdditionalFields&FIELDS_ORIGIN != 0 {
			for idx := 0; idx < 3; idx++ {
				WriteToBuf(playerName, frame.AtOrigin[idx])
			}
		}
		if frame.AdditionalFields&FIELDS_ANGLES != 0 {
			for idx := 0; idx < 3; idx++ {
				WriteToBuf(playerName, frame.AtAngles[idx])
			}
		}
		if frame.AdditionalFields&FIELDS_VELOCITY != 0 {
			for idx := 0; idx < 3; idx++ {
				WriteToBuf(playerName, frame.AtVelocity[idx])
			}
		}
	}

	delete(PlayerFramesMap, playerName)
	file.Write(bufMap[playerName].Bytes())
	delete(bufMap, playerName) // 清理buffer map，避免多回合状态残留
	ilog.InfoLogger.Printf("[Save] RoundFolder=%s, Player=%s.rec\n", roundFolder, safeName)
}

// WriteTickrateNoteFile 在当前 demo 的输出目录下创建一个名字为 tickrate 的记事本
func WriteTickrateNoteFile(tickrate float64) {
	// 保存到 output/[demoName]/ 目录；如果没有 demoName 子目录，则保存到 output/
	var basePath string
	if outputSubDir != "" {
		basePath = saveDir + "/" + outputSubDir
	} else {
		basePath = saveDir
	}
	if ok, _ := PathExists(basePath); !ok {
		os.MkdirAll(basePath, os.ModePerm)
	}
	// 文件名使用真实 tickrate 字符串（不做四舍五入），例如 64 或 128.015625
	name := strconv.FormatFloat(tickrate, 'f', -1, 64)
	fileName := basePath + "/" + name + ".txt"
	file, err := os.Create(fileName)
	if err != nil {
		ilog.ErrorLogger.Println("Failed to create tickrate note:", err.Error())
		return
	}
	defer file.Close()
	file.WriteString(fmt.Sprintf("tickrate=%s", name))
	ilog.InfoLogger.Printf("[Save] Tickrate note: %s\n", fileName)
}
