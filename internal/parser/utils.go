package parser

import (
	"math"

	encoder "github.com/hx-w/minidemo-encoder/internal/encoder"
	ilog "github.com/hx-w/minidemo-encoder/internal/logger"
	common "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/common"
)

const Pi = 3.14159265358979323846

var bufWeaponMap map[string]int32 = make(map[string]int32)
var playerLastZ map[string]float32 = make(map[string]float32)
var playerTeamMap map[string]string = make(map[string]string) // Track player team: "t" or "ct"

// ResetState 重置所有全局状态（批量解析时需要）
func ResetState() {
	bufWeaponMap = make(map[string]int32)
	playerLastZ = make(map[string]float32)
	playerTeamMap = make(map[string]string)
}

// Function to handle errors
func checkError(err error) {
	if err != nil {
		ilog.ErrorLogger.Println(err.Error())
	}
}

func parsePlayerInitFrame(player *common.Player) {
	// 先清理旧数据，确保初始化前状态干净
	delete(bufWeaponMap, player.Name)
	delete(encoder.PlayerFramesMap, player.Name)
	delete(playerLastZ, player.Name)

	iFrameInit := encoder.FrameInitInfo{
		PlayerName: player.Name,
	}
	iFrameInit.Position[0] = float32(player.Position().X)
	iFrameInit.Position[1] = float32(player.Position().Y)
	iFrameInit.Position[2] = float32(player.Position().Z)
	iFrameInit.Angles[0] = float32(player.ViewDirectionY())
	iFrameInit.Angles[1] = float32(player.ViewDirectionX())

	encoder.InitPlayer(iFrameInit)

	playerLastZ[player.Name] = float32(player.Position().Z)
	
	// Track player team
	if player.Team == common.TeamTerrorists {
		playerTeamMap[player.Name] = "t"
	} else if player.Team == common.TeamCounterTerrorists {
		playerTeamMap[player.Name] = "ct"
	}
}

func normalizeDegree(degree float64) float64 {
	if degree < 0.0 {
		degree = degree + 360.0
	}
	return degree
}

// accept radian, return degree in [0, 360)
func radian2degree(radian float64) float64 {
	return normalizeDegree(radian * 180 / Pi)
}

func parsePlayerFrame(player *common.Player, addonButton int32, tickrate float64, fullsnap bool) {
	if !player.IsAlive() {
		return
	}
	iFrameInfo := new(encoder.FrameInfo)
	iFrameInfo.PredictedVelocity[0] = 0.0
	iFrameInfo.PredictedVelocity[1] = 0.0
	iFrameInfo.PredictedVelocity[2] = 0.0
	iFrameInfo.ActualVelocity[0] = float32(player.Velocity().X)
	iFrameInfo.ActualVelocity[1] = float32(player.Velocity().Y)
	iFrameInfo.ActualVelocity[2] = float32(player.Velocity().Z)
	iFrameInfo.PredictedAngles[0] = player.ViewDirectionY()
	iFrameInfo.PredictedAngles[1] = player.ViewDirectionX()
	iFrameInfo.PlayerImpulse = 0
	iFrameInfo.PlayerSeed = 0
	iFrameInfo.PlayerSubtype = 0
	// ----- button encode
	iFrameInfo.PlayerButtons = ButtonConvert(player, addonButton)

	// ---- weapon encode
	var currWeaponID int32 = 0
	if player.ActiveWeapon() != nil {
		currWeaponID = int32(WeaponStr2ID(player.ActiveWeapon().String()))
	}
	if len(encoder.PlayerFramesMap[player.Name]) == 0 {
		iFrameInfo.CSWeaponID = currWeaponID
		bufWeaponMap[player.Name] = currWeaponID
	} else if currWeaponID == bufWeaponMap[player.Name] {
		iFrameInfo.CSWeaponID = int32(CSWeapon_NONE)
	} else {
		iFrameInfo.CSWeaponID = currWeaponID
		bufWeaponMap[player.Name] = currWeaponID
	}

	lastIdx := len(encoder.PlayerFramesMap[player.Name]) - 1
	// addons
	if fullsnap || (lastIdx < 2000 && (lastIdx+1)%int(tickrate) == 0) || (lastIdx >= 2000 && (lastIdx+1)%int(tickrate) == 0) {
		// if false {
		iFrameInfo.AdditionalFields |= encoder.FIELDS_ORIGIN
		iFrameInfo.AtOrigin[0] = float32(player.Position().X)
		iFrameInfo.AtOrigin[1] = float32(player.Position().Y)
		iFrameInfo.AtOrigin[2] = float32(player.Position().Z)
		// iFrameInfo.AdditionalFields |= encoder.FIELDS_ANGLES
		// iFrameInfo.AtAngles[0] = float32(player.ViewDirectionY())
		// iFrameInfo.AtAngles[1] = float32(player.ViewDirectionX())
		iFrameInfo.AdditionalFields |= encoder.FIELDS_VELOCITY
		iFrameInfo.AtVelocity[0] = float32(player.Velocity().X)
		iFrameInfo.AtVelocity[1] = float32(player.Velocity().Y)
		iFrameInfo.AtVelocity[2] = float32(player.Velocity().Z)
	}
	// record Z velocity (no delta, use engine velocity)
	playerLastZ[player.Name] = float32(player.Position().Z)

	iFrameInfo.ActualVelocity[2] = float32(player.Velocity().Z)

	// Since I don't know how to get player's button bits in a tick frame,
	// I have to use *actual vels* and *angles* to generate *predicted vels* approximately
	// This will cause some error, but it's not a big deal
	if lastIdx >= 0 { // not first frame
		// We assume that actual velocity in tick N
		// is influenced by predicted velocity in tick N-1
		_preVel := &encoder.PlayerFramesMap[player.Name][lastIdx].PredictedVelocity

		// PV = 0.0 when AV(tick N-1) = 0.0 and AV(tick N) = 0.0 ?
		// Note: AV=Actual Velocity, PV=Predicted Velocity
		if !(iFrameInfo.ActualVelocity[0] == 0.0 &&
			iFrameInfo.ActualVelocity[1] == 0.0 &&
			encoder.PlayerFramesMap[player.Name][lastIdx].ActualVelocity[0] == 0.0 &&
			encoder.PlayerFramesMap[player.Name][lastIdx].ActualVelocity[1] == 0.0) {
			var velAngle float64 = 0.0
			if iFrameInfo.ActualVelocity[0] == 0.0 {
				if iFrameInfo.ActualVelocity[1] < 0.0 {
					velAngle = 270.0
				} else {
					velAngle = 90.0
				}
			} else {
				velAngle = radian2degree(math.Atan2(float64(iFrameInfo.ActualVelocity[1]), float64(iFrameInfo.ActualVelocity[0])))
			}
			faceFront := normalizeDegree(float64(iFrameInfo.PredictedAngles[1]))
			deltaAngle := normalizeDegree(velAngle - faceFront)

			const threshold = 30.0
			if 0.0+threshold < deltaAngle && deltaAngle < 180.0-threshold {
				_preVel[1] = -450.0 // left
			}
			if 90.0+threshold < deltaAngle && deltaAngle < 270.0-threshold {
				_preVel[0] = -450.0 // back
			}
			if 180.0+threshold < deltaAngle && deltaAngle < 360.0-threshold {
				_preVel[1] = 450.0 // right
			}
			if 270.0+threshold < deltaAngle || deltaAngle < 90.0-threshold {
				_preVel[0] = 450.0 // front
			}
		}

	}

	iFrameInfo.Origin[0] = float32(player.Position().X)
	iFrameInfo.Origin[1] = float32(player.Position().Y)
	iFrameInfo.Origin[2] = float32(player.Position().Z)

	encoder.PlayerFramesMap[player.Name] = append(encoder.PlayerFramesMap[player.Name], *iFrameInfo)
}

func saveToRecFile(player *common.Player, roundFolder string) {
	if player.Team == common.TeamTerrorists {
		encoder.WriteToRecFile(player.Name, roundFolder, "t")
	} else {
		encoder.WriteToRecFile(player.Name, roundFolder, "ct")
	}
}

// flushRecordedPlayers saves all recorded players to .rec files
func flushRecordedPlayers(roundFolder string) {
	ilog.InfoLogger.Printf("Flushing recorded players for folder: %s\n", roundFolder)
	
	for playerName := range encoder.PlayerFramesMap {
		frameCount := len(encoder.PlayerFramesMap[playerName])
		if frameCount > 0 {
			team := playerTeamMap[playerName]
			if team == "" {
				team = "unknown"
			}
			encoder.WriteToRecFile(playerName, roundFolder, team)
			ilog.InfoLogger.Printf("Saved %d frames for player: %s (team: %s)\n", frameCount, playerName, team)
		}
	}
}
