package gdconf

import (
	"hk4e/common/constant"
	"hk4e/pkg/logger"
)

type HpDrop struct {
	Id        int32 // ID
	HpPercent int32 // 血量百分比
}

type FightProp struct {
	FightPropId    int32
	FightPropValue float32
}

// MonsterData 怪物配置表
type MonsterData struct {
	MonsterId    int32   `csv:"ID"`
	Name         string  `csv:"名称$text_name_Name,omitempty"`
	HpBase       float32 `csv:"基础生命值,omitempty"`
	AttackBase   float32 `csv:"基础攻击力,omitempty"`
	DefenseBase  float32 `csv:"基础防御力,omitempty"`
	Critical     float32 `csv:"暴击率,omitempty"`
	CriticalHurt float32 `csv:"暴击伤害,omitempty"`

	FireSubHurt     float32 `csv:"火元素抗性,omitempty"`
	GrassSubHurt    float32 `csv:"草元素抗性,omitempty"`
	WaterSubHurt    float32 `csv:"水元素抗性,omitempty"`
	ElecSubHurt     float32 `csv:"电元素抗性,omitempty"`
	WindSubHurt     float32 `csv:"风元素抗性,omitempty"`
	IceSubHurt      float32 `csv:"冰元素抗性,omitempty"`
	RockSubHurt     float32 `csv:"岩元素抗性,omitempty"`
	FireAddHurt     float32 `csv:"火元素伤害加成,omitempty"`
	GrassAddHurt    float32 `csv:"草元素伤害加成,omitempty"`
	WaterAddHurt    float32 `csv:"水元素伤害加成,omitempty"`
	ElecAddHurt     float32 `csv:"电元素伤害加成,omitempty"`
	WindAddHurt     float32 `csv:"风元素伤害加成,omitempty"`
	IceAddHurt      float32 `csv:"冰元素伤害加成,omitempty"`
	RockAddHurt     float32 `csv:"岩元素伤害加成,omitempty"`
	ElementMastery  float32 `csv:"元素精通,omitempty"`
	PhysicalSubHurt float32 `csv:"物理抗性,omitempty"`
	PhysicalAddHurt float32 `csv:"物理伤害加成,omitempty"`

	PropGrow1Type  int32 `csv:"[属性成长]1类型,omitempty"`
	PropGrow1Curve int32 `csv:"[属性成长]1曲线,omitempty"`
	PropGrow2Type  int32 `csv:"[属性成长]2类型,omitempty"`
	PropGrow2Curve int32 `csv:"[属性成长]2曲线,omitempty"`
	PropGrow3Type  int32 `csv:"[属性成长]3类型,omitempty"`
	PropGrow3Curve int32 `csv:"[属性成长]3曲线,omitempty"`

	Drop1Id        int32 `csv:"[掉落]1ID,omitempty"`
	Drop1HpPercent int32 `csv:"[掉落]1血量百分比,omitempty"`
	Drop2Id        int32 `csv:"[掉落]2ID,omitempty"`
	Drop2HpPercent int32 `csv:"[掉落]2血量百分比,omitempty"`
	Drop3Id        int32 `csv:"[掉落]3ID,omitempty"`
	Drop3HpPercent int32 `csv:"[掉落]3血量百分比,omitempty"`
	KillDropId     int32 `csv:"击杀掉落ID,omitempty"`

	FightPropList []*FightProp // 战斗属性列表
	PropGrowList  []*PropGrow  // 属性成长列表
	HpDropList    []*HpDrop    // 血量掉落列表
}

func (g *GameDataConfig) loadMonsterData() {
	g.MonsterDataMap = make(map[int32]*MonsterData)
	monsterDataList := make([]*MonsterData, 0)
	readTable[MonsterData](g.txtPrefix+"MonsterData.txt", &monsterDataList)
	for _, monsterData := range monsterDataList {
		// 战斗属性列表
		monsterData.FightPropList = make([]*FightProp, 0)
		if monsterData.FireSubHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_FIRE_SUB_HURT,
				FightPropValue: monsterData.FireSubHurt,
			})
		}
		if monsterData.GrassSubHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_GRASS_SUB_HURT,
				FightPropValue: monsterData.GrassSubHurt,
			})
		}
		if monsterData.WaterSubHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_WATER_SUB_HURT,
				FightPropValue: monsterData.WaterSubHurt,
			})
		}
		if monsterData.ElecSubHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ELEC_SUB_HURT,
				FightPropValue: monsterData.ElecSubHurt,
			})
		}
		if monsterData.WindSubHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_WIND_SUB_HURT,
				FightPropValue: monsterData.WindSubHurt,
			})
		}
		if monsterData.IceSubHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ICE_SUB_HURT,
				FightPropValue: monsterData.IceSubHurt,
			})
		}
		if monsterData.RockSubHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ROCK_SUB_HURT,
				FightPropValue: monsterData.RockSubHurt,
			})
		}
		if monsterData.FireAddHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_FIRE_ADD_HURT,
				FightPropValue: monsterData.FireAddHurt,
			})
		}
		if monsterData.GrassAddHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_GRASS_ADD_HURT,
				FightPropValue: monsterData.GrassAddHurt,
			})
		}
		if monsterData.WaterAddHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_WATER_ADD_HURT,
				FightPropValue: monsterData.WaterAddHurt,
			})
		}
		if monsterData.ElecAddHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ELEC_ADD_HURT,
				FightPropValue: monsterData.ElecAddHurt,
			})
		}
		if monsterData.WindAddHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_WIND_ADD_HURT,
				FightPropValue: monsterData.WindAddHurt,
			})
		}
		if monsterData.IceAddHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ICE_ADD_HURT,
				FightPropValue: monsterData.IceAddHurt,
			})
		}
		if monsterData.RockAddHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ROCK_ADD_HURT,
				FightPropValue: monsterData.RockAddHurt,
			})
		}
		if monsterData.ElementMastery != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ELEMENT_MASTERY,
				FightPropValue: monsterData.ElementMastery,
			})
		}
		if monsterData.PhysicalSubHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_PHYSICAL_SUB_HURT,
				FightPropValue: monsterData.PhysicalSubHurt,
			})
		}
		if monsterData.PhysicalAddHurt != 0.0 {
			monsterData.FightPropList = append(monsterData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_PHYSICAL_ADD_HURT,
				FightPropValue: monsterData.PhysicalAddHurt,
			})
		}
		// 属性成长列表
		propGrowList := make([]*PropGrow, 0)
		if monsterData.PropGrow1Type != 0 {
			propGrowList = append(propGrowList, &PropGrow{
				Type:  monsterData.PropGrow1Type,
				Curve: monsterData.PropGrow1Curve,
			})
		}
		if monsterData.PropGrow2Type != 0 {
			propGrowList = append(propGrowList, &PropGrow{
				Type:  monsterData.PropGrow2Type,
				Curve: monsterData.PropGrow2Curve,
			})
		}
		if monsterData.PropGrow3Type != 0 {
			propGrowList = append(propGrowList, &PropGrow{
				Type:  monsterData.PropGrow3Type,
				Curve: monsterData.PropGrow3Curve,
			})
		}
		monsterData.PropGrowList = propGrowList
		// 血量掉落列表
		monsterData.HpDropList = make([]*HpDrop, 0)
		if monsterData.Drop1Id != 0 {
			monsterData.HpDropList = append(monsterData.HpDropList, &HpDrop{
				Id:        monsterData.Drop1Id,
				HpPercent: monsterData.Drop1HpPercent,
			})
		}
		if monsterData.Drop2Id != 0 {
			monsterData.HpDropList = append(monsterData.HpDropList, &HpDrop{
				Id:        monsterData.Drop2Id,
				HpPercent: monsterData.Drop2HpPercent,
			})
		}
		if monsterData.Drop3Id != 0 {
			monsterData.HpDropList = append(monsterData.HpDropList, &HpDrop{
				Id:        monsterData.Drop3Id,
				HpPercent: monsterData.Drop3HpPercent,
			})
		}
		g.MonsterDataMap[monsterData.MonsterId] = monsterData
	}
	logger.Info("MonsterData count: %v", len(g.MonsterDataMap))
}

func GetMonsterDataById(monsterId int32) *MonsterData {
	return CONF.MonsterDataMap[monsterId]
}

func GetMonsterDataMap() map[int32]*MonsterData {
	return CONF.MonsterDataMap
}

// TODO 成长属性要读表

func (m *MonsterData) GetBaseHpByLevel(level uint8) float32 {
	return m.HpBase * float32(level)
}

func (m *MonsterData) GetBaseAttackByLevel(level uint8) float32 {
	return m.AttackBase * float32(level)
}

func (m *MonsterData) GetBaseDefenseByLevel(level uint8) float32 {
	return m.DefenseBase * float32(level)
}
