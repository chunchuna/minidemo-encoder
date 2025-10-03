package encoder

type FrameInitInfo struct {
	PlayerName string
	Position   [3]float32
	Angles     [2]float32
}

// replay frame
type FrameInfo struct {
	PlayerButtons     int32
	PlayerImpulse     int32
	ActualVelocity    [3]float32
	PredictedVelocity [3]float32
	PredictedAngles   [2]float32
	Origin            [3]float32 // 每一帧的当前位置！必须包含，否则botmimic无法正确读取
	CSWeaponID        int32
	PlayerSubtype     int32
	PlayerSeed        int32
	AdditionalFields  int32
	// 附加信息（用于teleport等特殊情况）
	AtOrigin   [3]float32
	AtAngles   [3]float32
	AtVelocity [3]float32
}