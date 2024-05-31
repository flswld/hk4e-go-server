package gdconf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hk4e/pkg/endec"
	"hk4e/pkg/logger"
	"os"
	"strings"

	"github.com/dengsgo/math-engine/engine"
	"github.com/hjson/hjson-go/v4"
)

type ConfigAbility struct {
	AbilityName string `json:"abilityName"`
}

type AbilityConfigData struct {
	Default *AbilityData `json:"Default"`
}

type AbilityModifierOrderMap struct {
	KeyOrder []string
	Map      map[string]*AbilityModifier
}

func (a *AbilityModifierOrderMap) UnmarshalJSON(data []byte) error {
	var formatJson bytes.Buffer
	err := json.Indent(&formatJson, data, "", "\t")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(formatJson.String(), "\n") {
		if strings.Count(line, "\t") == 1 && strings.Count(line, "\"") == 2 {
			key := strings.Split(line, "\"")[1]
			a.KeyOrder = append(a.KeyOrder, key)
		}
	}
	err = json.Unmarshal(data, &a.Map)
	if err != nil {
		return err
	}
	return nil
}

func (a *AbilityModifierOrderMap) GetByName(modifierName string) *AbilityModifier {
	return a.Map[modifierName]
}

func (a *AbilityModifierOrderMap) GetByLocalId(modifierLocalId uint32) *AbilityModifier {
	if int(modifierLocalId) >= len(a.KeyOrder) {
		return nil
	}
	key := a.KeyOrder[modifierLocalId]
	return a.Map[key]
}

type AbilityData struct {
	AbilityName        string                   `json:"abilityName"`
	Modifiers          *AbilityModifierOrderMap `json:"modifiers"`
	AbilitySpecials    map[string]float32       `json:"abilitySpecials"`
	OnAdded            []*AbilityAction         `json:"onAdded"`
	OnRemoved          []*AbilityAction         `json:"onRemoved"`
	OnAbilityStart     []*AbilityAction         `json:"onAbilityStart"`
	OnKill             []*AbilityAction         `json:"onKill"`
	OnFieldEnter       []*AbilityAction         `json:"onFieldEnter"`
	OnExit             []*AbilityAction         `json:"onExit"`
	OnAttach           []*AbilityAction         `json:"onAttach"`
	OnDetach           []*AbilityAction         `json:"onDetach"`
	OnAvatarIn         []*AbilityAction         `json:"onAvatarIn"`
	OnAvatarOut        []*AbilityAction         `json:"onAvatarOut"`
	OnTriggerAvatarRay []*AbilityAction         `json:"onTriggerAvatarRay"`
	OnVehicleIn        []*AbilityAction         `json:"onVehicleIn"`
	OnVehicleOut       []*AbilityAction         `json:"onVehicleOut"`
	AbilityMixins      []*AbilityMixinData      `json:"abilityMixins"`
}

type AbilityModifier struct {
	State             string                   `json:"state"`
	OnAdded           []*AbilityAction         `json:"onAdded"`
	OnThinkInterval   []*AbilityAction         `json:"onThinkInterval"`
	OnRemoved         []*AbilityAction         `json:"onRemoved"`
	OnBeingHit        []*AbilityAction         `json:"onBeingHit"`
	OnAttackLanded    []*AbilityAction         `json:"onAttackLanded"`
	OnHittingOther    []*AbilityAction         `json:"onHittingOther"`
	OnKill            []*AbilityAction         `json:"onKill"`
	OnCrash           []*AbilityAction         `json:"onCrash"`
	OnAvatarIn        []*AbilityAction         `json:"onAvatarIn"`
	OnAvatarOut       []*AbilityAction         `json:"onAvatarOut"`
	OnReconnect       []*AbilityAction         `json:"onReconnect"`
	OnChangeAuthority []*AbilityAction         `json:"onChangeAuthority"`
	OnVehicleIn       []*AbilityAction         `json:"onVehicleIn"`
	OnVehicleOut      []*AbilityAction         `json:"onVehicleOut"`
	OnZoneEnter       []*AbilityAction         `json:"onZoneEnter"`
	OnZoneExit        []*AbilityAction         `json:"onZoneExit"`
	OnHeal            []*AbilityAction         `json:"onHeal"`
	OnBeingHealed     []*AbilityAction         `json:"onBeingHealed"`
	Duration          DynamicFloat             `json:"duration"`
	ThinkInterval     DynamicFloat             `json:"thinkInterval"`
	Stacking          string                   `json:"stacking"`
	ModifierMixins    []*AbilityMixinData      `json:"modifierMixins"`
	Properties        *AbilityModifierProperty `json:"properties"`
	ElementType       string                   `json:"elementType"`
	ElementDurability DynamicFloat             `json:"elementDurability"`
}

type AbilityAction struct {
	Type                         string           `json:"$type"`
	Target                       string           `json:"target"`
	Amount                       DynamicFloat     `json:"amount"`
	AmountByCasterAttackRatio    DynamicFloat     `json:"amountByCasterAttackRatio"`
	AmountByCasterCurrentHPRatio DynamicFloat     `json:"amountByCasterCurrentHPRatio"`
	AmountByCasterMaxHPRatio     DynamicFloat     `json:"amountByCasterMaxHPRatio"`
	AmountByGetDamage            DynamicFloat     `json:"amountByGetDamage"`
	AmountByTargetCurrentHPRatio DynamicFloat     `json:"amountByTargetCurrentHPRatio"`
	AmountByTargetMaxHPRatio     DynamicFloat     `json:"amountByTargetMaxHPRatio"`
	LimboByTargetMaxHPRatio      DynamicFloat     `json:"limboByTargetMaxHPRatio"`
	HealRatio                    DynamicFloat     `json:"healRatio"`
	IgnoreAbilityProperty        bool             `json:"ignoreAbilityProperty"`
	ModifierName                 string           `json:"modifierName"`
	EnableLockHP                 bool             `json:"enableLockHP"`
	DisableWhenLoading           bool             `json:"disableWhenLoading"`
	Lethal                       bool             `json:"lethal"`
	MuteHealEffect               bool             `json:"muteHealEffect"`
	ByServer                     bool             `json:"byServer"`
	LifeByOwnerIsAlive           bool             `json:"lifeByOwnerIsAlive"`
	CampTargetType               string           `json:"campTargetType"`
	CampID                       int32            `json:"campID"`
	GadgetID                     int32            `json:"gadgetID"`
	OwnerIsTarget                bool             `json:"ownerIsTarget"`
	IsFromOwner                  bool             `json:"isFromOwner"`
	Key                          string           `json:"key"`
	GlobalValueKey               string           `json:"globalValueKey"`
	AbilityFormula               string           `json:"abilityFormula"`
	SrcTarget                    string           `json:"srcTarget"`
	DstTarget                    string           `json:"dstTarget"`
	SrcKey                       string           `json:"srcKey"`
	DstKey                       string           `json:"dstKey"`
	SkillID                      int32            `json:"skillID"`
	ResistanceListID             int32            `json:"resistanceListID"`
	MonsterID                    int32            `json:"monsterID"`
	SummonTag                    int32            `json:"summonTag"`
	Actions                      []*AbilityAction `json:"actions"`
	SuccessActions               []*AbilityAction `json:"successActions"`
	FailActions                  []*AbilityAction `json:"failActions"`
	DropType                     string           `json:"dropType"`
	BaseEnergy                   DynamicFloat     `json:"baseEnergy"`
	Ratio                        DynamicFloat     `json:"ratio"`
	ConfigID                     int32            `json:"configID"`
	ValueRangeMin                DynamicFloat     `json:"valueRangeMin"`
	ValueRangeMax                DynamicFloat     `json:"valueRangeMax"`
	OverrideMapKey               string           `json:"overrideMapKey"`
	Param1                       int32            `json:"param1"`
	Param2                       int32            `json:"param2"`
	Param3                       int32            `json:"param3"`
	FuncName                     string           `json:"funcName"`
	LuaCallType                  string           `json:"luaCallType"`
	CallParamList                []int32          `json:"callParamList"`
	Content                      string           `json:"content"`
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

func GetAbilityDataByHash(abilityNameHash uint32) *AbilityData {
	return CONF.AbilityDataHashMap[abilityNameHash]
}
