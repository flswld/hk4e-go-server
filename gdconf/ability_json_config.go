package gdconf

import (
	"fmt"
	"hk4e/pkg/endec"
	"hk4e/pkg/logger"
	"os"
	"strings"

	"github.com/dengsgo/math-engine/engine"
	"github.com/hjson/hjson-go/v4"
)

type ConfigAbility struct {
	AbilityID   string `json:"abilityID"`
	AbilityName string `json:"abilityName"`
}

type AbilityConfigData struct {
	Default *AbilityData `json:"Default"`
}

type AbilityData struct {
	AbilityName        string                      `json:"abilityName"`
	Modifiers          map[string]*AbilityModifier `json:"modifiers"`
	AbilitySpecials    map[string]float32          `json:"abilitySpecials"`
	OnAdded            []*AbilityModifierAction    `json:"onAdded"`
	OnRemoved          []*AbilityModifierAction    `json:"onRemoved"`
	OnAbilityStart     []*AbilityModifierAction    `json:"onAbilityStart"`
	OnKill             []*AbilityModifierAction    `json:"onKill"`
	OnFieldEnter       []*AbilityModifierAction    `json:"onFieldEnter"`
	OnExit             []*AbilityModifierAction    `json:"onExit"`
	OnAttach           []*AbilityModifierAction    `json:"onAttach"`
	OnDetach           []*AbilityModifierAction    `json:"onDetach"`
	OnAvatarIn         []*AbilityModifierAction    `json:"onAvatarIn"`
	OnAvatarOut        []*AbilityModifierAction    `json:"onAvatarOut"`
	OnTriggerAvatarRay []*AbilityModifierAction    `json:"onTriggerAvatarRay"`
	OnVehicleIn        []*AbilityModifierAction    `json:"onVehicleIn"`
	OnVehicleOut       []*AbilityModifierAction    `json:"onVehicleOut"`
	AbilityMixins      []*AbilityMixinData         `json:"abilityMixins"`
}

type AbilityModifier struct {
	State             string                   `json:"state"`
	OnAdded           []*AbilityModifierAction `json:"onAdded"`
	OnThinkInterval   []*AbilityModifierAction `json:"onThinkInterval"`
	OnRemoved         []*AbilityModifierAction `json:"onRemoved"`
	OnBeingHit        []*AbilityModifierAction `json:"onBeingHit"`
	OnAttackLanded    []*AbilityModifierAction `json:"onAttackLanded"`
	OnHittingOther    []*AbilityModifierAction `json:"onHittingOther"`
	OnKill            []*AbilityModifierAction `json:"onKill"`
	OnCrash           []*AbilityModifierAction `json:"onCrash"`
	OnAvatarIn        []*AbilityModifierAction `json:"onAvatarIn"`
	OnAvatarOut       []*AbilityModifierAction `json:"onAvatarOut"`
	OnReconnect       []*AbilityModifierAction `json:"onReconnect"`
	OnChangeAuthority []*AbilityModifierAction `json:"onChangeAuthority"`
	OnVehicleIn       []*AbilityModifierAction `json:"onVehicleIn"`
	OnVehicleOut      []*AbilityModifierAction `json:"onVehicleOut"`
	OnZoneEnter       []*AbilityModifierAction `json:"onZoneEnter"`
	OnZoneExit        []*AbilityModifierAction `json:"onZoneExit"`
	OnHeal            []*AbilityModifierAction `json:"onHeal"`
	OnBeingHealed     []*AbilityModifierAction `json:"onBeingHealed"`
	Duration          DynamicFloat             `json:"duration"`
	ThinkInterval     DynamicFloat             `json:"thinkInterval"`
	Stacking          string                   `json:"stacking"`
	ModifierMixins    []*AbilityMixinData      `json:"modifierMixins"`
	Properties        *AbilityModifierProperty `json:"properties"`
	ElementType       string                   `json:"elementType"`
	ElementDurability DynamicFloat             `json:"elementDurability"`
}

type AbilityModifierAction struct {
	Type                         string                   `json:"$type"`
	Target                       string                   `json:"target"`
	Amount                       DynamicFloat             `json:"amount"`
	AmountByCasterAttackRatio    DynamicFloat             `json:"amountByCasterAttackRatio"`
	AmountByCasterCurrentHPRatio DynamicFloat             `json:"amountByCasterCurrentHPRatio"`
	AmountByCasterMaxHPRatio     DynamicFloat             `json:"amountByCasterMaxHPRatio"`
	AmountByGetDamage            DynamicFloat             `json:"amountByGetDamage"`
	AmountByTargetCurrentHPRatio DynamicFloat             `json:"amountByTargetCurrentHPRatio"`
	AmountByTargetMaxHPRatio     DynamicFloat             `json:"amountByTargetMaxHPRatio"`
	LimboByTargetMaxHPRatio      DynamicFloat             `json:"limboByTargetMaxHPRatio"`
	HealRatio                    DynamicFloat             `json:"healRatio"`
	IgnoreAbilityProperty        bool                     `json:"ignoreAbilityProperty"`
	ModifierName                 string                   `json:"modifierName"`
	EnableLockHP                 bool                     `json:"enableLockHP"`
	DisableWhenLoading           bool                     `json:"disableWhenLoading"`
	Lethal                       bool                     `json:"lethal"`
	MuteHealEffect               bool                     `json:"muteHealEffect"`
	ByServer                     bool                     `json:"byServer"`
	LifeByOwnerIsAlive           bool                     `json:"lifeByOwnerIsAlive"`
	CampTargetType               string                   `json:"campTargetType"`
	CampID                       int32                    `json:"campID"`
	GadgetID                     int32                    `json:"gadgetID"`
	OwnerIsTarget                bool                     `json:"ownerIsTarget"`
	IsFromOwner                  bool                     `json:"isFromOwner"`
	Key                          string                   `json:"key"`
	GlobalValueKey               string                   `json:"globalValueKey"`
	AbilityFormula               string                   `json:"abilityFormula"`
	SrcTarget                    string                   `json:"srcTarget"`
	DstTarget                    string                   `json:"dstTarget"`
	SrcKey                       string                   `json:"srcKey"`
	DstKey                       string                   `json:"dstKey"`
	SkillID                      int32                    `json:"skillID"`
	ResistanceListID             int32                    `json:"resistanceListID"`
	MonsterID                    int32                    `json:"monsterID"`
	SummonTag                    int32                    `json:"summonTag"`
	Actions                      []*AbilityModifierAction `json:"actions"`
	SuccessActions               []*AbilityModifierAction `json:"successActions"`
	FailActions                  []*AbilityModifierAction `json:"failActions"`
	DropType                     string                   `json:"dropType"`
	BaseEnergy                   DynamicFloat             `json:"baseEnergy"`
	Ratio                        DynamicFloat             `json:"ratio"`
	ConfigID                     int32                    `json:"configID"`
	ValueRangeMin                DynamicFloat             `json:"valueRangeMin"`
	ValueRangeMax                DynamicFloat             `json:"valueRangeMax"`
	OverrideMapKey               string                   `json:"overrideMapKey"`
	Param1                       int32                    `json:"param1"`
	Param2                       int32                    `json:"param2"`
	Param3                       int32                    `json:"param3"`
	FuncName                     string                   `json:"funcName"`
	LuaCallType                  string                   `json:"luaCallType"`
	CallParamList                []int32                  `json:"callParamList"`
	Content                      string                   `json:"content"`
}

type AbilityMixinData struct {
	Type         string `json:"$type"`
	ModifierName string `json:"modifierName"`
}

type AbilityModifierProperty struct {
	Actor_HpThresholdRatio DynamicFloat `json:"Actor_HpThresholdRatio"`
}

type DynamicFloat interface {
}

func (a *AbilityData) GetDynamicFloat(dynamicFloat DynamicFloat) float32 {
	switch dynamicFloat.(type) {
	case float64:
		return float32(dynamicFloat.(float64))
	case string:
		rawExp := dynamicFloat.(string)
		exp := ""
		for i := 0; i < len(rawExp); i++ {
			c := string(rawExp[i])
			if c == "%" {
				for j := i + 1; j < len(rawExp); j++ {
					cc := string(rawExp[j])
					end := j == len(rawExp)-1
					if cc == "+" || cc == "-" || cc == "*" || cc == "/" || end {
						key := ""
						if end {
							key = rawExp[i+1 : j+1]
						} else {
							key = rawExp[i+1 : j]
						}
						value, exist := a.AbilitySpecials[key]
						if !exist {
							logger.Error("ability special key not exist, key: %v", key)
							return 0.0
						}
						exp += fmt.Sprintf("%f", value)
						if end {
							i = j
						} else {
							i = j - 1
						}
						break
					}
				}
			} else {
				exp += c
			}
		}
		r, err := engine.ParseAndExec(exp)
		if err != nil {
			logger.Error("calc dynamic float error: %v", err)
			return 0.0
		}
		return float32(r)
	default:
		return 0.0
	}
}

func (g *GameDataConfig) loadAbilityJsonConfig() {
	g.AbilityDataMap = make(map[string]*AbilityData)
	g.AbilityDataHashMap = make(map[uint32]*AbilityData)
	g.loadAbilityJsonConfigLoop(g.jsonPrefix + "ability")
	logger.Info("AbilityDataMap Count: %v, AbilityDataHashMap Count: %v", len(g.AbilityDataMap), len(g.AbilityDataHashMap))
}

func (g *GameDataConfig) loadAbilityJsonConfigLoop(path string) {
	fileList, err := os.ReadDir(path)
	if err != nil {
		info := fmt.Sprintf("open file error: %v, path: %v", err, path)
		panic(info)
	}
	for _, file := range fileList {
		fileName := file.Name()
		if file.IsDir() {
			g.loadAbilityJsonConfigLoop(path + "/" + fileName)
		}
		if split := strings.Split(fileName, "."); split[len(split)-1] != "json" {
			continue
		}
		fileData, err := os.ReadFile(path + "/" + fileName)
		if err != nil {
			info := fmt.Sprintf("open file error: %v, path: %v", err, path+"/"+fileName)
			panic(info)
		}
		var abilityConfigDataList []*AbilityConfigData = nil
		err = hjson.Unmarshal(fileData, &abilityConfigDataList)
		if err != nil {
			logger.Info("parse file error: %v, path: %v", err, path+"/"+fileName)
			continue
		}
		for _, abilityConfigData := range abilityConfigDataList {
			abilityData := abilityConfigData.Default
			g.AbilityDataMap[abilityData.AbilityName] = abilityData
			g.AbilityDataHashMap[uint32(endec.Hk4eAbilityHashCode(abilityData.AbilityName))] = abilityData
		}
	}
}

func GetAbilityDataByName(abilityName string) *AbilityData {
	return CONF.AbilityDataMap[abilityName]
}

func GetAbilityDataByHash(abilityHashCode uint32) *AbilityData {
	return CONF.AbilityDataHashMap[abilityHashCode]
}
