package gdconf

import (
	"encoding/json"
	"fmt"
	"hk4e/pkg/endec"
	"hk4e/pkg/logger"
	"os"
	"sort"
	"strings"

	"github.com/hjson/hjson-go/v4"
)

const (
	LocalIdTypeAbilityAction  = 1
	LocalIdTypeAbilityMixin   = 2
	LocalIdTypeModifierAction = 3
	LocalIdTypeModifierMixin  = 4
)

type ConfigAbility struct {
	AbilityName string `json:"abilityName"`
}

type ConfigAbilityJson struct {
	Abilities       []*ConfigAbility `json:"abilities"`
	TargetAbilities []*ConfigAbility `json:"targetAbilities"`
}

type AbilityJsonConfig struct {
	Default *AbilityData `json:"Default"`
}

type AbilityData struct {
	AbilityName     string             `json:"abilityName"`
	Modifiers       ModifierOrderMap   `json:"modifiers"`
	AbilitySpecials map[string]float32 `json:"abilitySpecials"`

	OnAdded            []*ActionData `json:"onAdded"`
	OnRemoved          []*ActionData `json:"onRemoved"`
	OnAbilityStart     []*ActionData `json:"onAbilityStart"`
	OnKill             []*ActionData `json:"onKill"`
	OnFieldEnter       []*ActionData `json:"onFieldEnter"`
	OnExit             []*ActionData `json:"onExit"`
	OnAttach           []*ActionData `json:"onAttach"`
	OnDetach           []*ActionData `json:"onDetach"`
	OnAvatarIn         []*ActionData `json:"onAvatarIn"`
	OnAvatarOut        []*ActionData `json:"onAvatarOut"`
	OnTriggerAvatarRay []*ActionData `json:"onTriggerAvatarRay"`
	OnVehicleIn        []*ActionData `json:"onVehicleIn"`
	OnVehicleOut       []*ActionData `json:"onVehicleOut"`

	AbilityMixins []*MixinData `json:"abilityMixins"`

	LocalIdActionMap map[int32]*ActionData
	LocalIdMixinMap  map[int32]*MixinData
}

type ModifierData struct {
	State             string        `json:"state"`
	Duration          DynamicFloat  `json:"duration"`
	ThinkInterval     DynamicFloat  `json:"thinkInterval"`
	Stacking          string        `json:"stacking"`
	Properties        *PropertyData `json:"properties"`
	ElementType       string        `json:"elementType"`
	ElementDurability DynamicFloat  `json:"elementDurability"`

	OnAdded           []*ActionData `json:"onAdded"`
	OnRemoved         []*ActionData `json:"onRemoved"`
	OnBeingHit        []*ActionData `json:"onBeingHit"`
	OnAttackLanded    []*ActionData `json:"onAttackLanded"`
	OnHittingOther    []*ActionData `json:"onHittingOther"`
	OnThinkInterval   []*ActionData `json:"onThinkInterval"`
	OnKill            []*ActionData `json:"onKill"`
	OnCrash           []*ActionData `json:"onCrash"`
	OnAvatarIn        []*ActionData `json:"onAvatarIn"`
	OnAvatarOut       []*ActionData `json:"onAvatarOut"`
	OnReconnect       []*ActionData `json:"onReconnect"`
	OnChangeAuthority []*ActionData `json:"onChangeAuthority"`
	OnVehicleIn       []*ActionData `json:"onVehicleIn"`
	OnVehicleOut      []*ActionData `json:"onVehicleOut"`
	OnZoneEnter       []*ActionData `json:"onZoneEnter"`
	OnZoneExit        []*ActionData `json:"onZoneExit"`
	OnHeal            []*ActionData `json:"onHeal"`
	OnBeingHealed     []*ActionData `json:"onBeingHealed"`

	ModifierMixins []*MixinData `json:"modifierMixins"`
}

type ActionData struct {
	Type                         string       `json:"$type"`
	Target                       string       `json:"target"`
	Amount                       DynamicFloat `json:"amount"`
	AmountByCasterAttackRatio    DynamicFloat `json:"amountByCasterAttackRatio"`
	AmountByCasterCurrentHPRatio DynamicFloat `json:"amountByCasterCurrentHPRatio"`
	AmountByCasterMaxHPRatio     DynamicFloat `json:"amountByCasterMaxHPRatio"`
	AmountByGetDamage            DynamicFloat `json:"amountByGetDamage"`
	AmountByTargetCurrentHPRatio DynamicFloat `json:"amountByTargetCurrentHPRatio"`
	AmountByTargetMaxHPRatio     DynamicFloat `json:"amountByTargetMaxHPRatio"`
	LimboByTargetMaxHPRatio      DynamicFloat `json:"limboByTargetMaxHPRatio"`
	HealRatio                    DynamicFloat `json:"healRatio"`
	IgnoreAbilityProperty        bool         `json:"ignoreAbilityProperty"`
	ModifierName                 string       `json:"modifierName"`
	EnableLockHP                 bool         `json:"enableLockHP"`
	DisableWhenLoading           bool         `json:"disableWhenLoading"`
	Lethal                       bool         `json:"lethal"`
	MuteHealEffect               bool         `json:"muteHealEffect"`
	ByServer                     bool         `json:"byServer"`
	LifeByOwnerIsAlive           bool         `json:"lifeByOwnerIsAlive"`
	CampTargetType               string       `json:"campTargetType"`
	CampID                       int32        `json:"campID"`
	GadgetID                     int32        `json:"gadgetID"`
	OwnerIsTarget                bool         `json:"ownerIsTarget"`
	IsFromOwner                  bool         `json:"isFromOwner"`
	Key                          string       `json:"key"`
	GlobalValueKey               string       `json:"globalValueKey"`
	AbilityFormula               string       `json:"abilityFormula"`
	SrcTarget                    string       `json:"srcTarget"`
	DstTarget                    string       `json:"dstTarget"`
	SrcKey                       string       `json:"srcKey"`
	DstKey                       string       `json:"dstKey"`
	SkillID                      int32        `json:"skillID"`
	ResistanceListID             int32        `json:"resistanceListID"`
	MonsterID                    int32        `json:"monsterID"`
	SummonTag                    int32        `json:"summonTag"`
	DropType                     string       `json:"dropType"`
	BaseEnergy                   DynamicFloat `json:"baseEnergy"`
	Ratio                        DynamicFloat `json:"ratio"`
	ConfigID                     int32        `json:"configID"`
	ValueRangeMin                DynamicFloat `json:"valueRangeMin"`
	ValueRangeMax                DynamicFloat `json:"valueRangeMax"`
	OverrideMapKey               string       `json:"overrideMapKey"`
	Param1                       int32        `json:"param1"`
	Param2                       int32        `json:"param2"`
	Param3                       int32        `json:"param3"`
	FuncName                     string       `json:"funcName"`
	LuaCallType                  string       `json:"luaCallType"`
	CallParamList                []int32      `json:"callParamList"`
	Content                      string       `json:"content"`
	CostStaminaRatio             DynamicFloat `json:"costStaminaRatio"`

	Actions        []*ActionData `json:"actions"`
	SuccessActions []*ActionData `json:"successActions"`
	FailActions    []*ActionData `json:"failActions"`
}

type MixinData struct {
	Type             string       `json:"$type"`
	ModifierName     string       `json:"modifierName"`
	CostStaminaDelta DynamicFloat `json:"costStaminaDelta"`
}

type PropertyData struct {
	Actor_HpThresholdRatio DynamicFloat `json:"Actor_HpThresholdRatio"`
}

type ModifierOrderMap struct {
	KeyOrder []string
	Map      map[string]*ModifierData
}

func (m *ModifierOrderMap) UnmarshalJSON(data []byte) error {
	m.Map = make(map[string]*ModifierData)
	err := json.Unmarshal(data, &m.Map)
	if err != nil {
		return err
	}
	m.KeyOrder = make([]string, 0)
	for key := range m.Map {
		m.KeyOrder = append(m.KeyOrder, key)
	}
	sort.Slice(m.KeyOrder, func(i, j int) bool {
		return m.KeyOrder[i] < m.KeyOrder[j]
	})
	return nil
}

func (m *ModifierOrderMap) GetByName(modifierName string) *ModifierData {
	return m.Map[modifierName]
}

func (m *ModifierOrderMap) GetByLocalId(modifierLocalId uint32) *ModifierData {
	if int(modifierLocalId) >= len(m.KeyOrder) {
		return nil
	}
	key := m.KeyOrder[modifierLocalId]
	return m.Map[key]
}

type DynamicFloat interface {
}

func (g *GameDataConfig) loadAbilityJsonConfig() {
	g.AbilityDataMap = make(map[string]*AbilityData)
	g.AbilityDataHashMap = make(map[uint32]*AbilityData)
	g.loadAbilityJsonConfigLoop(g.jsonPrefix + "ability")
	logger.Info("AbilityData Count: %v, AbilityDataHash Count: %v", len(g.AbilityDataMap), len(g.AbilityDataHashMap))
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
		var abilityJsonConfigList []*AbilityJsonConfig = nil
		err = hjson.Unmarshal(fileData, &abilityJsonConfigList)
		if err != nil {
			logger.Info("parse file error: %v, path: %v", err, path+"/"+fileName)
			continue
		}
		for _, abilityJsonConfig := range abilityJsonConfigList {
			abilityData := abilityJsonConfig.Default
			abilityData.genLocalId()
			g.AbilityDataMap[abilityData.AbilityName] = abilityData
			g.AbilityDataHashMap[uint32(endec.Hk4eAbilityHashCode(abilityData.AbilityName))] = abilityData
		}
	}
}

func (a *AbilityData) genAbilitySubActionLocalId(actionIndex *int, configIndex int, actionList []*ActionData) {
	for _, action := range actionList {
		*actionIndex++
		localId := *actionIndex<<9 + configIndex<<3 + LocalIdTypeAbilityAction
		a.LocalIdActionMap[int32(localId)] = action
		a.genAbilitySubActionLocalId(actionIndex, configIndex, action.Actions)
		a.genAbilitySubActionLocalId(actionIndex, configIndex, action.SuccessActions)
		a.genAbilitySubActionLocalId(actionIndex, configIndex, action.FailActions)
	}
}

func (a *AbilityData) genModifierSubActionLocalId(actionIndex *int, configIndex int, modifierIndex int, actionList []*ActionData) {
	for _, action := range actionList {
		*actionIndex++
		localId := *actionIndex<<15 + configIndex<<9 + modifierIndex<<3 + LocalIdTypeModifierAction
		a.LocalIdActionMap[int32(localId)] = action
		a.genModifierSubActionLocalId(actionIndex, configIndex, modifierIndex, action.Actions)
		a.genModifierSubActionLocalId(actionIndex, configIndex, modifierIndex, action.SuccessActions)
		a.genModifierSubActionLocalId(actionIndex, configIndex, modifierIndex, action.FailActions)
	}
}

func (a *AbilityData) genLocalId() {
	a.LocalIdActionMap = make(map[int32]*ActionData)
	a.LocalIdMixinMap = make(map[int32]*MixinData)
	configIndex := 0
	genAbilityActionLocalId := func(actionList []*ActionData) {
		actionIndex := 0
		a.genAbilitySubActionLocalId(&actionIndex, configIndex, actionList)
		configIndex++
	}
	genAbilityActionLocalId(a.OnAdded)
	genAbilityActionLocalId(a.OnRemoved)
	genAbilityActionLocalId(a.OnAbilityStart)
	genAbilityActionLocalId(a.OnKill)
	genAbilityActionLocalId(a.OnFieldEnter)
	genAbilityActionLocalId(a.OnExit)
	genAbilityActionLocalId(a.OnAttach)
	genAbilityActionLocalId(a.OnDetach)
	genAbilityActionLocalId(a.OnAvatarIn)
	genAbilityActionLocalId(a.OnAvatarOut)
	genAbilityActionLocalId(a.OnTriggerAvatarRay)
	genAbilityActionLocalId(a.OnVehicleIn)
	genAbilityActionLocalId(a.OnVehicleOut)
	for mixinIndex, mixin := range a.AbilityMixins {
		localId := mixinIndex<<3 + LocalIdTypeAbilityMixin
		a.LocalIdMixinMap[int32(localId)] = mixin
	}
	for modifierIndex, key := range a.Modifiers.KeyOrder {
		modifier := a.Modifiers.Map[key]
		configIndex = 0
		genModifierActionLocalId := func(actionList []*ActionData) {
			actionIndex := 0
			a.genModifierSubActionLocalId(&actionIndex, configIndex, modifierIndex, actionList)
			configIndex++
		}
		genModifierActionLocalId(modifier.OnAdded)
		genModifierActionLocalId(modifier.OnRemoved)
		genModifierActionLocalId(modifier.OnBeingHit)
		genModifierActionLocalId(modifier.OnAttackLanded)
		genModifierActionLocalId(modifier.OnHittingOther)
		genModifierActionLocalId(modifier.OnThinkInterval)
		genModifierActionLocalId(modifier.OnKill)
		genModifierActionLocalId(modifier.OnCrash)
		genModifierActionLocalId(modifier.OnAvatarIn)
		genModifierActionLocalId(modifier.OnAvatarOut)
		genModifierActionLocalId(modifier.OnReconnect)
		genModifierActionLocalId(modifier.OnChangeAuthority)
		genModifierActionLocalId(modifier.OnVehicleIn)
		genModifierActionLocalId(modifier.OnVehicleOut)
		genModifierActionLocalId(modifier.OnZoneEnter)
		genModifierActionLocalId(modifier.OnZoneExit)
		genModifierActionLocalId(modifier.OnHeal)
		genModifierActionLocalId(modifier.OnBeingHealed)
		for mixinIndex, mixin := range modifier.ModifierMixins {
			localId := mixinIndex<<9 + modifierIndex<<3 + LocalIdTypeModifierMixin
			a.LocalIdMixinMap[int32(localId)] = mixin
		}
	}
}

func GetAbilityDataByName(abilityName string) *AbilityData {
	return CONF.AbilityDataMap[abilityName]
}

func GetAbilityDataByHash(abilityNameHash uint32) *AbilityData {
	return CONF.AbilityDataHashMap[abilityNameHash]
}

func (a *AbilityData) GetActionDataByLocalId(localId int32) *ActionData {
	return a.LocalIdActionMap[localId]
}

func (a *AbilityData) GetMixinDataByLocalId(localId int32) *MixinData {
	return a.LocalIdMixinMap[localId]
}
