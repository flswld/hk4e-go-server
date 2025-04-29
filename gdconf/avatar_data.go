package gdconf

import (
	"fmt"
	"os"

	"hk4e/common/constant"

	"github.com/flswld/halo/logger"
	"github.com/hjson/hjson-go/v4"
)

// FightProp 战斗属性
type FightProp struct {
	FightPropId    int32
	FightPropValue float32
}

// PropGrow 属性成长
type PropGrow struct {
	Type  int32 // 类型
	Curve int32 // 曲线
}

// AvatarData 角色配置表
type AvatarData struct {
	AvatarId           int32    `csv:"ID"`
	QualityType        int32    `csv:"角色品质,omitempty"`
	ConfigJson         string   `csv:"战斗config,omitempty"`
	InitialWeapon      int32    `csv:"初始武器,omitempty"`
	WeaponType         int32    `csv:"武器种类,omitempty"`
	SkillDepotId       int32    `csv:"技能库ID,omitempty"`
	SkillDepotIdList   IntArray `csv:"候选技能库ID,omitempty"`
	PromoteId          int32    `csv:"角色突破ID,omitempty"`
	PromoteRewardLevel IntArray `csv:"角色突破奖励获取等阶,omitempty"`
	PromoteReward      IntArray `csv:"角色突破奖励,omitempty"`
	// 战斗属性
	HpBase           float32 `csv:"基础生命值,omitempty"`
	AttackBase       float32 `csv:"基础攻击力,omitempty"`
	DefenseBase      float32 `csv:"基础防御力,omitempty"`
	Critical         float32 `csv:"暴击率,omitempty"`
	CriticalHurt     float32 `csv:"暴击伤害,omitempty"`
	FireSubHurt      float32 `csv:"火元素抗性,omitempty"`
	GrassSubHurt     float32 `csv:"草元素抗性,omitempty"`
	WaterSubHurt     float32 `csv:"水元素抗性,omitempty"`
	ElecSubHurt      float32 `csv:"电元素抗性,omitempty"`
	WindSubHurt      float32 `csv:"风元素抗性,omitempty"`
	IceSubHurt       float32 `csv:"冰元素抗性,omitempty"`
	RockSubHurt      float32 `csv:"岩元素抗性,omitempty"`
	FireAddHurt      float32 `csv:"火元素伤害加成,omitempty"`
	GrassAddHurt     float32 `csv:"草元素伤害加成,omitempty"`
	WaterAddHurt     float32 `csv:"水元素伤害加成,omitempty"`
	ElecAddHurt      float32 `csv:"电元素伤害加成,omitempty"`
	WindAddHurt      float32 `csv:"风元素伤害加成,omitempty"`
	IceAddHurt       float32 `csv:"冰元素伤害加成,omitempty"`
	RockAddHurt      float32 `csv:"岩元素伤害加成,omitempty"`
	ElementMastery   float32 `csv:"元素精通,omitempty"`
	PhysicalSubHurt  float32 `csv:"物理抗性,omitempty"`
	PhysicalAddHurt  float32 `csv:"物理伤害加成,omitempty"`
	ChargeEfficiency float32 `csv:"充能效率,omitempty"`
	// 属性成长
	PropGrow1Type  int32 `csv:"[属性成长]1类型,omitempty"`
	PropGrow1Curve int32 `csv:"[属性成长]1曲线,omitempty"`
	PropGrow2Type  int32 `csv:"[属性成长]2类型,omitempty"`
	PropGrow2Curve int32 `csv:"[属性成长]2曲线,omitempty"`
	PropGrow3Type  int32 `csv:"[属性成长]3类型,omitempty"`
	PropGrow3Curve int32 `csv:"[属性成长]3曲线,omitempty"`

	ConfigAbility    *ConfigAbilityJson // 能力配置
	PromoteRewardMap map[uint32]uint32  // 突破奖励集合
	FightPropList    []*FightProp       // 战斗属性列表
	PropGrowList     []*PropGrow        // 属性成长列表
}

func (g *GameDataConfig) loadAvatarData() {
	g.AvatarDataMap = make(map[int32]*AvatarData)
	avatarDataList := make([]*AvatarData, 0)
	readTable[AvatarData](g.txtPrefix+"AvatarData.txt", &avatarDataList)
	for _, avatarData := range avatarDataList {
		fileData, err := os.ReadFile(g.jsonPrefix + "avatar/" + avatarData.ConfigJson + ".json")
		if err != nil {
			info := fmt.Sprintf("open file error: %v", err)
			panic(info)
		}
		if fileData[0] == 0xEF && fileData[1] == 0xBB && fileData[2] == 0xBF {
			fileData = fileData[3:]
		}
		configAbilityJson := new(ConfigAbilityJson)
		err = hjson.Unmarshal(fileData, configAbilityJson)
		if err != nil {
			info := fmt.Sprintf("parse file error: %v, avatarId: %v", err, avatarData.AvatarId)
			panic(info)
		}
		avatarData.ConfigAbility = configAbilityJson
		// 突破奖励转换列表
		if len(avatarData.PromoteRewardLevel) != 0 && len(avatarData.PromoteReward) != 0 {
			avatarData.PromoteRewardMap = make(map[uint32]uint32, len(avatarData.PromoteReward))
			for index, rewardId := range avatarData.PromoteReward {
				promoteLevel := avatarData.PromoteRewardLevel[index]
				avatarData.PromoteRewardMap[uint32(promoteLevel)] = uint32(rewardId)
			}
		}
		// 战斗属性列表
		avatarData.FightPropList = make([]*FightProp, 0)
		if avatarData.HpBase != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_BASE_HP,
				FightPropValue: avatarData.HpBase,
			})
		}
		if avatarData.AttackBase != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_BASE_ATTACK,
				FightPropValue: avatarData.AttackBase,
			})
		}
		if avatarData.DefenseBase != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_BASE_DEFENSE,
				FightPropValue: avatarData.DefenseBase,
			})
		}
		if avatarData.Critical != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_CRITICAL,
				FightPropValue: avatarData.Critical,
			})
		}
		if avatarData.CriticalHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_CRITICAL_HURT,
				FightPropValue: avatarData.CriticalHurt,
			})
		}
		if avatarData.FireSubHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_FIRE_SUB_HURT,
				FightPropValue: avatarData.FireSubHurt,
			})
		}
		if avatarData.GrassSubHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_GRASS_SUB_HURT,
				FightPropValue: avatarData.GrassSubHurt,
			})
		}
		if avatarData.WaterSubHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_WATER_SUB_HURT,
				FightPropValue: avatarData.WaterSubHurt,
			})
		}
		if avatarData.ElecSubHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ELEC_SUB_HURT,
				FightPropValue: avatarData.ElecSubHurt,
			})
		}
		if avatarData.WindSubHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_WIND_SUB_HURT,
				FightPropValue: avatarData.WindSubHurt,
			})
		}
		if avatarData.IceSubHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ICE_SUB_HURT,
				FightPropValue: avatarData.IceSubHurt,
			})
		}
		if avatarData.RockSubHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ROCK_SUB_HURT,
				FightPropValue: avatarData.RockSubHurt,
			})
		}
		if avatarData.FireAddHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_FIRE_ADD_HURT,
				FightPropValue: avatarData.FireAddHurt,
			})
		}
		if avatarData.GrassAddHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_GRASS_ADD_HURT,
				FightPropValue: avatarData.GrassAddHurt,
			})
		}
		if avatarData.WaterAddHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_WATER_ADD_HURT,
				FightPropValue: avatarData.WaterAddHurt,
			})
		}
		if avatarData.ElecAddHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ELEC_ADD_HURT,
				FightPropValue: avatarData.ElecAddHurt,
			})
		}
		if avatarData.WindAddHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_WIND_ADD_HURT,
				FightPropValue: avatarData.WindAddHurt,
			})
		}
		if avatarData.IceAddHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ICE_ADD_HURT,
				FightPropValue: avatarData.IceAddHurt,
			})
		}
		if avatarData.RockAddHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ROCK_ADD_HURT,
				FightPropValue: avatarData.RockAddHurt,
			})
		}
		if avatarData.ElementMastery != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_ELEMENT_MASTERY,
				FightPropValue: avatarData.ElementMastery,
			})
		}
		if avatarData.PhysicalSubHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_PHYSICAL_SUB_HURT,
				FightPropValue: avatarData.PhysicalSubHurt,
			})
		}
		if avatarData.PhysicalAddHurt != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_PHYSICAL_ADD_HURT,
				FightPropValue: avatarData.PhysicalAddHurt,
			})
		}
		if avatarData.ChargeEfficiency != 0.0 {
			avatarData.FightPropList = append(avatarData.FightPropList, &FightProp{
				FightPropId:    constant.FIGHT_PROP_CHARGE_EFFICIENCY,
				FightPropValue: avatarData.ChargeEfficiency,
			})
		}
		// 属性成长列表
		propGrowList := make([]*PropGrow, 0)
		if avatarData.PropGrow1Type != 0 {
			propGrowList = append(propGrowList, &PropGrow{
				Type:  avatarData.PropGrow1Type,
				Curve: avatarData.PropGrow1Curve,
			})
		}
		if avatarData.PropGrow2Type != 0 {
			propGrowList = append(propGrowList, &PropGrow{
				Type:  avatarData.PropGrow2Type,
				Curve: avatarData.PropGrow2Curve,
			})
		}
		if avatarData.PropGrow3Type != 0 {
			propGrowList = append(propGrowList, &PropGrow{
				Type:  avatarData.PropGrow3Type,
				Curve: avatarData.PropGrow3Curve,
			})
		}
		avatarData.PropGrowList = propGrowList
		g.AvatarDataMap[avatarData.AvatarId] = avatarData
	}
	logger.Info("AvatarData Count: %v", len(g.AvatarDataMap))
}

func GetAvatarDataById(avatarId int32) *AvatarData {
	return CONF.AvatarDataMap[avatarId]
}

func GetAvatarDataMap() map[int32]*AvatarData {
	return CONF.AvatarDataMap
}

func GetAvatarFightPropMap(avatarId uint32, level uint8, promote uint8) map[uint32]float32 {
	fightPropMap := make(map[uint32]float32)
	avatarConfig := GetAvatarDataById(int32(avatarId))
	if avatarConfig == nil {
		logger.Error("avatar config is nil, avatarId: %v", avatarId)
		return fightPropMap
	}
	for _, fightProp := range avatarConfig.FightPropList {
		fightPropId := fightProp.FightPropId
		fightPropValue := fightProp.FightPropValue
		for _, propGrow := range avatarConfig.PropGrowList {
			if propGrow.Type == fightPropId {
				avatarCurveConfig := GetAvatarCurveByLevelAndType(int32(level), propGrow.Curve)
				if avatarCurveConfig == nil {
					logger.Error("avatar curve config is nil, level: %v, curveType: %v", level, propGrow.Curve)
					return fightPropMap
				}
				fightPropValue *= avatarCurveConfig.Value
			}
		}
		avatarPromoteConfig := GetAvatarPromoteDataByIdAndLevel(avatarConfig.PromoteId, int32(promote))
		if avatarPromoteConfig == nil {
			logger.Error("avatar promote config is nil, promoteId: %v, promoteLevel: %v", avatarConfig.PromoteId, promote)
			return fightPropMap
		}
		for _, addProp := range avatarPromoteConfig.AddPropList {
			if addProp.Type == fightPropId {
				fightPropValue += addProp.Value
			}
		}
		fightPropMap[uint32(fightPropId)] = fightPropValue
	}
	return fightPropMap
}
