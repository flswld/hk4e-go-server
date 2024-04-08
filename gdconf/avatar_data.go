package gdconf

import (
	"fmt"
	"hk4e/common/constant"
	"os"

	"hk4e/pkg/endec"
	"hk4e/pkg/logger"

	"github.com/hjson/hjson-go/v4"
)

// PropGrow 属性成长
type PropGrow struct {
	Type  int32 // 类型
	Curve int32 // 曲线
}

// AvatarData 角色配置表
type AvatarData struct {
	AvatarId           int32    `csv:"ID"`
	HpBase             float32  `csv:"基础生命值,omitempty"`
	AttackBase         float32  `csv:"基础攻击力,omitempty"`
	DefenseBase        float32  `csv:"基础防御力,omitempty"`
	Critical           float32  `csv:"暴击率,omitempty"`
	CriticalHurt       float32  `csv:"暴击伤害,omitempty"`
	QualityType        int32    `csv:"角色品质,omitempty"`
	ConfigJson         string   `csv:"战斗config,omitempty"`
	InitialWeapon      int32    `csv:"初始武器,omitempty"`
	WeaponType         int32    `csv:"武器种类,omitempty"`
	SkillDepotId       int32    `csv:"技能库ID,omitempty"`
	SkillDepotIdList   IntArray `csv:"候选技能库ID,omitempty"`
	PromoteId          int32    `csv:"角色突破ID,omitempty"`
	PromoteRewardLevel IntArray `csv:"角色突破奖励获取等阶,omitempty"`
	PromoteReward      IntArray `csv:"角色突破奖励,omitempty"`

	PropGrow1Type  int32 `csv:"[属性成长]1类型,omitempty"`
	PropGrow1Curve int32 `csv:"[属性成长]1曲线,omitempty"`
	PropGrow2Type  int32 `csv:"[属性成长]2类型,omitempty"`
	PropGrow2Curve int32 `csv:"[属性成长]2曲线,omitempty"`
	PropGrow3Type  int32 `csv:"[属性成长]3类型,omitempty"`
	PropGrow3Curve int32 `csv:"[属性成长]3曲线,omitempty"`

	AbilityHashCodeList []int32
	PromoteRewardMap    map[uint32]uint32
	PropGrowList        []*PropGrow
}

type ConfigAvatar struct {
	Abilities       []*ConfigAvatarAbility `json:"abilities"`
	TargetAbilities []*ConfigAvatarAbility `json:"targetAbilities"`
}

type ConfigAvatarAbility struct {
	AbilityName string `json:"abilityName"`
}

func (g *GameDataConfig) loadAvatarData() {
	g.AvatarDataMap = make(map[int32]*AvatarData)
	avatarDataList := make([]*AvatarData, 0)
	readTable[AvatarData](g.txtPrefix+"AvatarData.txt", &avatarDataList)
	for _, avatarData := range avatarDataList {
		// 读取战斗config解析技能并转化为哈希码
		fileData, err := os.ReadFile(g.jsonPrefix + "avatar/" + avatarData.ConfigJson + ".json")
		if err != nil {
			info := fmt.Sprintf("open file error: %v, AvatarId: %v", err, avatarData.AvatarId)
			panic(info)
		}
		configAvatar := new(ConfigAvatar)
		err = hjson.Unmarshal(fileData, configAvatar)
		if err != nil {
			info := fmt.Sprintf("parse file error: %v, AvatarId: %v", err, avatarData.AvatarId)
			panic(info)
		}
		if len(configAvatar.Abilities) == 0 {
			logger.Info("can not find any ability of avatar, AvatarId: %v", avatarData.AvatarId)
		}
		for _, configAvatarAbility := range configAvatar.Abilities {
			abilityHashCode := endec.Hk4eAbilityHashCode(configAvatarAbility.AbilityName)
			avatarData.AbilityHashCodeList = append(avatarData.AbilityHashCodeList, abilityHashCode)
		}
		// 突破奖励转换列表
		if len(avatarData.PromoteRewardLevel) != 0 && len(avatarData.PromoteReward) != 0 {
			avatarData.PromoteRewardMap = make(map[uint32]uint32, len(avatarData.PromoteReward))
			for index, rewardId := range avatarData.PromoteReward {
				promoteLevel := avatarData.PromoteRewardLevel[index]
				avatarData.PromoteRewardMap[uint32(promoteLevel)] = uint32(rewardId)
			}
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
	logger.Info("AvatarData count: %v", len(g.AvatarDataMap))
}

func GetAvatarDataById(avatarId int32) *AvatarData {
	return CONF.AvatarDataMap[avatarId]
}

func GetAvatarDataMap() map[int32]*AvatarData {
	return CONF.AvatarDataMap
}

// GetAvatarBaseADH 获取角色基础攻防血
func GetAvatarBaseADH(avatarId uint32, level uint8, promote uint8, fightProp int) float32 {
	adh := float32(0.0)
	avatarConfig := GetAvatarDataById(int32(avatarId))
	if avatarConfig == nil {
		logger.Error("avatar config is nil, avatarId: %v", avatarId)
		return adh
	}
	switch fightProp {
	case constant.FIGHT_PROP_BASE_ATTACK:
		adh += avatarConfig.AttackBase
	case constant.FIGHT_PROP_BASE_DEFENSE:
		adh += avatarConfig.DefenseBase
	case constant.FIGHT_PROP_BASE_HP:
		adh += avatarConfig.HpBase
	}
	for _, propGrow := range avatarConfig.PropGrowList {
		if propGrow.Type == int32(fightProp) {
			avatarCurveConfig := GetAvatarCurveByLevelAndType(int32(level), propGrow.Curve)
			if avatarCurveConfig == nil {
				logger.Error("avatar curve config is nil, level: %v, curveType: %v", level, propGrow.Curve)
				return adh
			}
			adh *= avatarCurveConfig.Value
		}
	}
	avatarPromoteConfig := GetAvatarPromoteDataByIdAndLevel(avatarConfig.PromoteId, int32(promote))
	if avatarPromoteConfig == nil {
		logger.Error("avatar promote config is nil, promoteId: %v, promoteLevel: %v", avatarConfig.PromoteId, promote)
		return adh
	}
	for _, addProp := range avatarPromoteConfig.AddPropList {
		if addProp.Type == int32(fightProp) {
			adh += addProp.Value
		}
	}
	return adh
}
