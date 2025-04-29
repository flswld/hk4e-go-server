package gdconf

import (
	"github.com/flswld/halo/logger"
)

// MonsterCurveData 怪物曲线配置表
type MonsterCurveData struct {
	Level int32 `csv:"等级"`
	// 曲线
	Curve1Type        int32   `csv:"[曲线]1类型,omitempty"`
	Curve1Arithmetic  int32   `csv:"[曲线]1运算,omitempty"`
	Curve1Value       float32 `csv:"[曲线]1值,omitempty"`
	Curve2Type        int32   `csv:"[曲线]2类型,omitempty"`
	Curve2Arithmetic  int32   `csv:"[曲线]2运算,omitempty"`
	Curve2Value       float32 `csv:"[曲线]2值,omitempty"`
	Curve3Type        int32   `csv:"[曲线]3类型,omitempty"`
	Curve3Arithmetic  int32   `csv:"[曲线]3运算,omitempty"`
	Curve3Value       float32 `csv:"[曲线]3值,omitempty"`
	Curve4Type        int32   `csv:"[曲线]4类型,omitempty"`
	Curve4Arithmetic  int32   `csv:"[曲线]4运算,omitempty"`
	Curve4Value       float32 `csv:"[曲线]4值,omitempty"`
	Curve5Type        int32   `csv:"[曲线]5类型,omitempty"`
	Curve5Arithmetic  int32   `csv:"[曲线]5运算,omitempty"`
	Curve5Value       float32 `csv:"[曲线]5值,omitempty"`
	Curve6Type        int32   `csv:"[曲线]6类型,omitempty"`
	Curve6Arithmetic  int32   `csv:"[曲线]6运算,omitempty"`
	Curve6Value       float32 `csv:"[曲线]6值,omitempty"`
	Curve7Type        int32   `csv:"[曲线]7类型,omitempty"`
	Curve7Arithmetic  int32   `csv:"[曲线]7运算,omitempty"`
	Curve7Value       float32 `csv:"[曲线]7值,omitempty"`
	Curve8Type        int32   `csv:"[曲线]8类型,omitempty"`
	Curve8Arithmetic  int32   `csv:"[曲线]8运算,omitempty"`
	Curve8Value       float32 `csv:"[曲线]8值,omitempty"`
	Curve9Type        int32   `csv:"[曲线]9类型,omitempty"`
	Curve9Arithmetic  int32   `csv:"[曲线]9运算,omitempty"`
	Curve9Value       float32 `csv:"[曲线]9值,omitempty"`
	Curve10Type       int32   `csv:"[曲线]10类型,omitempty"`
	Curve10Arithmetic int32   `csv:"[曲线]10运算,omitempty"`
	Curve10Value      float32 `csv:"[曲线]10值,omitempty"`
	Curve11Type       int32   `csv:"[曲线]11类型,omitempty"`
	Curve11Arithmetic int32   `csv:"[曲线]11运算,omitempty"`
	Curve11Value      float32 `csv:"[曲线]11值,omitempty"`
	Curve12Type       int32   `csv:"[曲线]12类型,omitempty"`
	Curve12Arithmetic int32   `csv:"[曲线]12运算,omitempty"`
	Curve12Value      float32 `csv:"[曲线]12值,omitempty"`

	CurveList []*Curve // 曲线列表
}

func (g *GameDataConfig) loadMonsterCurveData() {
	g.MonsterCurveDataMap = make(map[int32]*MonsterCurveData)
	monsterCurveDataList := make([]*MonsterCurveData, 0)
	readTable[MonsterCurveData](g.txtPrefix+"MonsterCurveData.txt", &monsterCurveDataList)
	for _, monsterCurveData := range monsterCurveDataList {
		curveList := make([]*Curve, 0)
		if monsterCurveData.Curve1Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve1Type,
				Arithmetic: monsterCurveData.Curve1Arithmetic,
				Value:      monsterCurveData.Curve1Value,
			})
		}
		if monsterCurveData.Curve2Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve2Type,
				Arithmetic: monsterCurveData.Curve2Arithmetic,
				Value:      monsterCurveData.Curve2Value,
			})
		}
		if monsterCurveData.Curve3Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve3Type,
				Arithmetic: monsterCurveData.Curve3Arithmetic,
				Value:      monsterCurveData.Curve3Value,
			})
		}
		if monsterCurveData.Curve4Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve4Type,
				Arithmetic: monsterCurveData.Curve4Arithmetic,
				Value:      monsterCurveData.Curve4Value,
			})
		}
		if monsterCurveData.Curve5Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve5Type,
				Arithmetic: monsterCurveData.Curve5Arithmetic,
				Value:      monsterCurveData.Curve5Value,
			})
		}
		if monsterCurveData.Curve6Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve6Type,
				Arithmetic: monsterCurveData.Curve6Arithmetic,
				Value:      monsterCurveData.Curve6Value,
			})
		}
		if monsterCurveData.Curve7Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve7Type,
				Arithmetic: monsterCurveData.Curve7Arithmetic,
				Value:      monsterCurveData.Curve7Value,
			})
		}
		if monsterCurveData.Curve8Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve8Type,
				Arithmetic: monsterCurveData.Curve8Arithmetic,
				Value:      monsterCurveData.Curve8Value,
			})
		}
		if monsterCurveData.Curve9Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve9Type,
				Arithmetic: monsterCurveData.Curve9Arithmetic,
				Value:      monsterCurveData.Curve9Value,
			})
		}
		if monsterCurveData.Curve10Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve10Type,
				Arithmetic: monsterCurveData.Curve10Arithmetic,
				Value:      monsterCurveData.Curve10Value,
			})
		}
		if monsterCurveData.Curve11Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve11Type,
				Arithmetic: monsterCurveData.Curve11Arithmetic,
				Value:      monsterCurveData.Curve11Value,
			})
		}
		if monsterCurveData.Curve12Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       monsterCurveData.Curve12Type,
				Arithmetic: monsterCurveData.Curve12Arithmetic,
				Value:      monsterCurveData.Curve12Value,
			})
		}
		monsterCurveData.CurveList = curveList
		g.MonsterCurveDataMap[monsterCurveData.Level] = monsterCurveData
	}
	logger.Info("MonsterCurveData Count: %v", len(g.MonsterCurveDataMap))
}

func GetMonsterCurveByLevelAndType(level int32, curveType int32) *Curve {
	monsterCurveData, exist := CONF.MonsterCurveDataMap[level]
	if !exist {
		return nil
	}
	for _, curve := range monsterCurveData.CurveList {
		if curve.Type == curveType {
			return curve
		}
	}
	return nil
}

func GetMonsterCurveDataMap() map[int32]*MonsterCurveData {
	return CONF.MonsterCurveDataMap
}
