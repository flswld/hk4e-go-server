package gdconf

import (
	"hk4e/pkg/logger"
)

// WeaponCurveData 武器曲线配置表
type WeaponCurveData struct {
	Level int32 `csv:"等级"`

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
	Curve13Type       int32   `csv:"[曲线]13类型,omitempty"`
	Curve13Arithmetic int32   `csv:"[曲线]13运算,omitempty"`
	Curve13Value      float32 `csv:"[曲线]13值,omitempty"`
	Curve14Type       int32   `csv:"[曲线]14类型,omitempty"`
	Curve14Arithmetic int32   `csv:"[曲线]14运算,omitempty"`
	Curve14Value      float32 `csv:"[曲线]14值,omitempty"`
	Curve15Type       int32   `csv:"[曲线]15类型,omitempty"`
	Curve15Arithmetic int32   `csv:"[曲线]15运算,omitempty"`
	Curve15Value      float32 `csv:"[曲线]15值,omitempty"`
	Curve16Type       int32   `csv:"[曲线]16类型,omitempty"`
	Curve16Arithmetic int32   `csv:"[曲线]16运算,omitempty"`
	Curve16Value      float32 `csv:"[曲线]16值,omitempty"`
	Curve17Type       int32   `csv:"[曲线]17类型,omitempty"`
	Curve17Arithmetic int32   `csv:"[曲线]17运算,omitempty"`
	Curve17Value      float32 `csv:"[曲线]17值,omitempty"`
	Curve18Type       int32   `csv:"[曲线]18类型,omitempty"`
	Curve18Arithmetic int32   `csv:"[曲线]18运算,omitempty"`
	Curve18Value      float32 `csv:"[曲线]18值,omitempty"`

	CurveList []*Curve
}

func (g *GameDataConfig) loadWeaponCurveData() {
	g.WeaponCurveDataMap = make(map[int32]*WeaponCurveData)
	weaponCurveDataList := make([]*WeaponCurveData, 0)
	readTable[WeaponCurveData](g.txtPrefix+"WeaponCurveData.txt", &weaponCurveDataList)
	for _, weaponCurveData := range weaponCurveDataList {
		curveList := make([]*Curve, 0)
		if weaponCurveData.Curve1Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve1Type,
				Arithmetic: weaponCurveData.Curve1Arithmetic,
				Value:      weaponCurveData.Curve1Value,
			})
		}
		if weaponCurveData.Curve2Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve2Type,
				Arithmetic: weaponCurveData.Curve2Arithmetic,
				Value:      weaponCurveData.Curve2Value,
			})
		}
		if weaponCurveData.Curve3Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve3Type,
				Arithmetic: weaponCurveData.Curve3Arithmetic,
				Value:      weaponCurveData.Curve3Value,
			})
		}
		if weaponCurveData.Curve4Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve4Type,
				Arithmetic: weaponCurveData.Curve4Arithmetic,
				Value:      weaponCurveData.Curve4Value,
			})
		}
		if weaponCurveData.Curve5Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve5Type,
				Arithmetic: weaponCurveData.Curve5Arithmetic,
				Value:      weaponCurveData.Curve5Value,
			})
		}
		if weaponCurveData.Curve6Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve6Type,
				Arithmetic: weaponCurveData.Curve6Arithmetic,
				Value:      weaponCurveData.Curve6Value,
			})
		}
		if weaponCurveData.Curve7Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve7Type,
				Arithmetic: weaponCurveData.Curve7Arithmetic,
				Value:      weaponCurveData.Curve7Value,
			})
		}
		if weaponCurveData.Curve8Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve8Type,
				Arithmetic: weaponCurveData.Curve8Arithmetic,
				Value:      weaponCurveData.Curve8Value,
			})
		}
		if weaponCurveData.Curve9Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve9Type,
				Arithmetic: weaponCurveData.Curve9Arithmetic,
				Value:      weaponCurveData.Curve9Value,
			})
		}
		if weaponCurveData.Curve10Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve10Type,
				Arithmetic: weaponCurveData.Curve10Arithmetic,
				Value:      weaponCurveData.Curve10Value,
			})
		}
		if weaponCurveData.Curve11Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve11Type,
				Arithmetic: weaponCurveData.Curve11Arithmetic,
				Value:      weaponCurveData.Curve11Value,
			})
		}
		if weaponCurveData.Curve12Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve12Type,
				Arithmetic: weaponCurveData.Curve12Arithmetic,
				Value:      weaponCurveData.Curve12Value,
			})
		}
		if weaponCurveData.Curve13Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve13Type,
				Arithmetic: weaponCurveData.Curve13Arithmetic,
				Value:      weaponCurveData.Curve13Value,
			})
		}
		if weaponCurveData.Curve14Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve14Type,
				Arithmetic: weaponCurveData.Curve14Arithmetic,
				Value:      weaponCurveData.Curve14Value,
			})
		}
		if weaponCurveData.Curve15Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve15Type,
				Arithmetic: weaponCurveData.Curve15Arithmetic,
				Value:      weaponCurveData.Curve15Value,
			})
		}
		if weaponCurveData.Curve16Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve16Type,
				Arithmetic: weaponCurveData.Curve16Arithmetic,
				Value:      weaponCurveData.Curve16Value,
			})
		}
		if weaponCurveData.Curve17Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve17Type,
				Arithmetic: weaponCurveData.Curve17Arithmetic,
				Value:      weaponCurveData.Curve17Value,
			})
		}
		if weaponCurveData.Curve18Type != 0 {
			curveList = append(curveList, &Curve{
				Type:       weaponCurveData.Curve18Type,
				Arithmetic: weaponCurveData.Curve18Arithmetic,
				Value:      weaponCurveData.Curve18Value,
			})
		}
		weaponCurveData.CurveList = curveList
		g.WeaponCurveDataMap[weaponCurveData.Level] = weaponCurveData
	}
	logger.Info("WeaponCurveData Count: %v", len(g.WeaponCurveDataMap))
}

func GetWeaponCurveByLevelAndType(level int32, curveType int32) *Curve {
	weaponCurveData, exist := CONF.WeaponCurveDataMap[level]
	if !exist {
		return nil
	}
	for _, curve := range weaponCurveData.CurveList {
		if curve.Type == curveType {
			return curve
		}
	}
	return nil
}

func GetWeaponCurveDataMap() map[int32]*WeaponCurveData {
	return CONF.WeaponCurveDataMap
}
