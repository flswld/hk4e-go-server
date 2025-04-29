package gdconf

import (
	"github.com/flswld/halo/logger"
)

// ReliquaryLevelData 圣遗物等级配置表
type ReliquaryLevelData struct {
	Stage int32 `csv:"阶数"`
	Level int32 `csv:"等级,omitempty"`
	Exp   int32 `csv:"成长到下一级所需经验,omitempty"`

	AddProp1Type   int32   `csv:"[增加属性]1类型,omitempty"`
	AddProp1Value  float32 `csv:"[增加属性]1值,omitempty"`
	AddProp2Type   int32   `csv:"[增加属性]2类型,omitempty"`
	AddProp2Value  float32 `csv:"[增加属性]2值,omitempty"`
	AddProp3Type   int32   `csv:"[增加属性]3类型,omitempty"`
	AddProp3Value  float32 `csv:"[增加属性]3值,omitempty"`
	AddProp4Type   int32   `csv:"[增加属性]4类型,omitempty"`
	AddProp4Value  float32 `csv:"[增加属性]4值,omitempty"`
	AddProp5Type   int32   `csv:"[增加属性]5类型,omitempty"`
	AddProp5Value  float32 `csv:"[增加属性]5值,omitempty"`
	AddProp6Type   int32   `csv:"[增加属性]6类型,omitempty"`
	AddProp6Value  float32 `csv:"[增加属性]6值,omitempty"`
	AddProp7Type   int32   `csv:"[增加属性]7类型,omitempty"`
	AddProp7Value  float32 `csv:"[增加属性]7值,omitempty"`
	AddProp8Type   int32   `csv:"[增加属性]8类型,omitempty"`
	AddProp8Value  float32 `csv:"[增加属性]8值,omitempty"`
	AddProp9Type   int32   `csv:"[增加属性]9类型,omitempty"`
	AddProp9Value  float32 `csv:"[增加属性]9值,omitempty"`
	AddProp10Type  int32   `csv:"[增加属性]10类型,omitempty"`
	AddProp10Value float32 `csv:"[增加属性]10值,omitempty"`
	AddProp11Type  int32   `csv:"[增加属性]11类型,omitempty"`
	AddProp11Value float32 `csv:"[增加属性]11值,omitempty"`
	AddProp12Type  int32   `csv:"[增加属性]12类型,omitempty"`
	AddProp12Value float32 `csv:"[增加属性]12值,omitempty"`
	AddProp13Type  int32   `csv:"[增加属性]13类型,omitempty"`
	AddProp13Value float32 `csv:"[增加属性]13值,omitempty"`
	AddProp14Type  int32   `csv:"[增加属性]14类型,omitempty"`
	AddProp14Value float32 `csv:"[增加属性]14值,omitempty"`
	AddProp15Type  int32   `csv:"[增加属性]15类型,omitempty"`
	AddProp15Value float32 `csv:"[增加属性]15值,omitempty"`
	AddProp16Type  int32   `csv:"[增加属性]16类型,omitempty"`
	AddProp16Value float32 `csv:"[增加属性]16值,omitempty"`
	AddProp17Type  int32   `csv:"[增加属性]17类型,omitempty"`
	AddProp17Value float32 `csv:"[增加属性]17值,omitempty"`
	AddProp18Type  int32   `csv:"[增加属性]18类型,omitempty"`
	AddProp18Value float32 `csv:"[增加属性]18值,omitempty"`
	AddProp19Type  int32   `csv:"[增加属性]19类型,omitempty"`
	AddProp19Value float32 `csv:"[增加属性]19值,omitempty"`
	AddProp20Type  int32   `csv:"[增加属性]20类型,omitempty"`
	AddProp20Value float32 `csv:"[增加属性]20值,omitempty"`

	AddPropMap map[int32]*AddProp // 增加属性集合
}

func (g *GameDataConfig) loadReliquaryLevelData() {
	g.ReliquaryLevelDataMap = make(map[int32]map[int32]*ReliquaryLevelData)
	reliquaryLevelDataList := make([]*ReliquaryLevelData, 0)
	readTable[ReliquaryLevelData](g.txtPrefix+"ReliquaryLevelData.txt", &reliquaryLevelDataList)
	for _, reliquaryLevelData := range reliquaryLevelDataList {
		_, ok := g.ReliquaryLevelDataMap[reliquaryLevelData.Stage]
		if !ok {
			g.ReliquaryLevelDataMap[reliquaryLevelData.Stage] = make(map[int32]*ReliquaryLevelData)
		}
		// 增加属性集合
		addPropMap := make(map[int32]*AddProp)
		if reliquaryLevelData.AddProp1Type != 0 {
			addPropMap[reliquaryLevelData.AddProp1Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp1Type,
				Value: reliquaryLevelData.AddProp1Value,
			}
		}
		if reliquaryLevelData.AddProp2Type != 0 {
			addPropMap[reliquaryLevelData.AddProp2Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp2Type,
				Value: reliquaryLevelData.AddProp2Value,
			}
		}
		if reliquaryLevelData.AddProp3Type != 0 {
			addPropMap[reliquaryLevelData.AddProp3Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp3Type,
				Value: reliquaryLevelData.AddProp3Value,
			}
		}
		if reliquaryLevelData.AddProp4Type != 0 {
			addPropMap[reliquaryLevelData.AddProp4Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp4Type,
				Value: reliquaryLevelData.AddProp4Value,
			}
		}
		if reliquaryLevelData.AddProp5Type != 0 {
			addPropMap[reliquaryLevelData.AddProp5Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp5Type,
				Value: reliquaryLevelData.AddProp5Value,
			}
		}
		if reliquaryLevelData.AddProp6Type != 0 {
			addPropMap[reliquaryLevelData.AddProp6Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp6Type,
				Value: reliquaryLevelData.AddProp6Value,
			}
		}
		if reliquaryLevelData.AddProp7Type != 0 {
			addPropMap[reliquaryLevelData.AddProp7Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp7Type,
				Value: reliquaryLevelData.AddProp7Value,
			}
		}
		if reliquaryLevelData.AddProp8Type != 0 {
			addPropMap[reliquaryLevelData.AddProp8Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp8Type,
				Value: reliquaryLevelData.AddProp8Value,
			}
		}
		if reliquaryLevelData.AddProp9Type != 0 {
			addPropMap[reliquaryLevelData.AddProp9Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp9Type,
				Value: reliquaryLevelData.AddProp9Value,
			}
		}
		if reliquaryLevelData.AddProp10Type != 0 {
			addPropMap[reliquaryLevelData.AddProp10Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp10Type,
				Value: reliquaryLevelData.AddProp10Value,
			}
		}
		if reliquaryLevelData.AddProp11Type != 0 {
			addPropMap[reliquaryLevelData.AddProp11Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp11Type,
				Value: reliquaryLevelData.AddProp11Value,
			}
		}
		if reliquaryLevelData.AddProp12Type != 0 {
			addPropMap[reliquaryLevelData.AddProp12Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp12Type,
				Value: reliquaryLevelData.AddProp12Value,
			}
		}
		if reliquaryLevelData.AddProp13Type != 0 {
			addPropMap[reliquaryLevelData.AddProp13Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp13Type,
				Value: reliquaryLevelData.AddProp13Value,
			}
		}
		if reliquaryLevelData.AddProp14Type != 0 {
			addPropMap[reliquaryLevelData.AddProp14Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp14Type,
				Value: reliquaryLevelData.AddProp14Value,
			}
		}
		if reliquaryLevelData.AddProp15Type != 0 {
			addPropMap[reliquaryLevelData.AddProp15Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp15Type,
				Value: reliquaryLevelData.AddProp15Value,
			}
		}
		if reliquaryLevelData.AddProp16Type != 0 {
			addPropMap[reliquaryLevelData.AddProp16Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp16Type,
				Value: reliquaryLevelData.AddProp16Value,
			}
		}
		if reliquaryLevelData.AddProp17Type != 0 {
			addPropMap[reliquaryLevelData.AddProp17Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp17Type,
				Value: reliquaryLevelData.AddProp17Value,
			}
		}
		if reliquaryLevelData.AddProp18Type != 0 {
			addPropMap[reliquaryLevelData.AddProp18Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp18Type,
				Value: reliquaryLevelData.AddProp18Value,
			}
		}
		if reliquaryLevelData.AddProp19Type != 0 {
			addPropMap[reliquaryLevelData.AddProp19Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp19Type,
				Value: reliquaryLevelData.AddProp19Value,
			}
		}
		if reliquaryLevelData.AddProp20Type != 0 {
			addPropMap[reliquaryLevelData.AddProp20Type] = &AddProp{
				Type:  reliquaryLevelData.AddProp20Type,
				Value: reliquaryLevelData.AddProp20Value,
			}
		}
		reliquaryLevelData.AddPropMap = addPropMap
		// 通过突破等级找到突破数据
		g.ReliquaryLevelDataMap[reliquaryLevelData.Stage][reliquaryLevelData.Level] = reliquaryLevelData
	}
	logger.Info("ReliquaryLevelData Count: %v", len(g.ReliquaryLevelDataMap))
}

func GetReliquaryLevelDataByStageAndLevel(stage int32, level int32) *ReliquaryLevelData {
	value, exist := CONF.ReliquaryLevelDataMap[stage]
	if !exist {
		return nil
	}
	return value[level]
}
