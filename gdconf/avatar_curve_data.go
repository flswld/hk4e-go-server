package gdconf

import (
	"hk4e/pkg/logger"
)

// Curve 曲线
type Curve struct {
	Type       int32   // 类型
	Arithmetic int32   // 运算
	Value      float32 // 值
}

// AvatarCurveData 角色曲线配置表
type AvatarCurveData struct {
	Level int32 `csv:"等级"`

	Curve1Type       int32   `csv:"[曲线]1类型,omitempty"`
	Curve1Arithmetic int32   `csv:"[曲线]1运算,omitempty"`
	Curve1Value      float32 `csv:"[曲线]1值,omitempty"`
	Curve2Type       int32   `csv:"[曲线]2类型,omitempty"`
	Curve2Arithmetic int32   `csv:"[曲线]2运算,omitempty"`
	Curve2Value      float32 `csv:"[曲线]2值,omitempty"`
	Curve3Type       int32   `csv:"[曲线]3类型,omitempty"`
	Curve3Arithmetic int32   `csv:"[曲线]3运算,omitempty"`
	Curve3Value      float32 `csv:"[曲线]3值,omitempty"`
	Curve4Type       int32   `csv:"[曲线]4类型,omitempty"`
	Curve4Arithmetic int32   `csv:"[曲线]4运算,omitempty"`
	Curve4Value      float32 `csv:"[曲线]4值,omitempty"`

	CurveList []*Curve
}

func (g *GameDataConfig) loadAvatarCurveData() {
	g.AvatarCurveDataMap = make(map[int32]*AvatarCurveData)
	avatarCurveDataList := make([]*AvatarCurveData, 0)
	readTable[AvatarCurveData](g.txtPrefix+"AvatarCurveData.txt", &avatarCurveDataList)
	for _, avatarCurveData := range avatarCurveDataList {
		curveList := make([]*Curve, 0)
		if avatarCurveData.Curve1Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       avatarCurveData.Curve1Type,
				Arithmetic: avatarCurveData.Curve1Arithmetic,
				Value:      avatarCurveData.Curve1Value,
			})
		}
		if avatarCurveData.Curve2Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       avatarCurveData.Curve2Type,
				Arithmetic: avatarCurveData.Curve2Arithmetic,
				Value:      avatarCurveData.Curve2Value,
			})
		}
		if avatarCurveData.Curve3Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       avatarCurveData.Curve3Type,
				Arithmetic: avatarCurveData.Curve3Arithmetic,
				Value:      avatarCurveData.Curve3Value,
			})
		}
		if avatarCurveData.Curve4Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       avatarCurveData.Curve4Type,
				Arithmetic: avatarCurveData.Curve4Arithmetic,
				Value:      avatarCurveData.Curve4Value,
			})
		}
		avatarCurveData.CurveList = curveList
		g.AvatarCurveDataMap[avatarCurveData.Level] = avatarCurveData
	}
	logger.Info("AvatarCurveData count: %v", len(g.AvatarCurveDataMap))
}

func GetAvatarCurveByLevelAndType(level int32, curveType int32) *Curve {
	avatarCurveData, exist := CONF.AvatarCurveDataMap[level]
	if !exist {
		return nil
	}
	for _, curve := range avatarCurveData.CurveList {
		if curve.Type == curveType {
			return curve
		}
	}
	return nil
}

func GetAvatarCurveDataMap() map[int32]*AvatarCurveData {
	return CONF.AvatarCurveDataMap
}
