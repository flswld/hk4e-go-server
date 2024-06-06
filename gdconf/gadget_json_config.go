package gdconf

import (
	"fmt"
	"hk4e/pkg/logger"
	"os"
	"strings"

	"github.com/hjson/hjson-go/v4"
)

type ConfigGadget struct {
	Type string `json:"$type"`
	ConfigGadgetVehicle
	ConfigAbilityJson
}

type ConfigGadgetVehicle struct {
	Vehicle *ConfigVehicle       `json:"vehicle"` // 载具
	Combat  *ConfigVehicleCombat `json:"combat"`  // 战斗
}

type ConfigVehicle struct {
	VehicleType  string                `json:"vehicleType"`  // 载具类型
	DefaultLevel int32                 `json:"defaultLevel"` // 默认等级
	MaxSeatCount int32                 `json:"maxSeatCount"` // 最大座位数
	Stamina      *ConfigVehicleStamina `json:"stamina"`      // 耐力
}

type ConfigVehicleStamina struct {
	StaminaUpperLimit      float32 `json:"staminaUpperLimit"`      // 耐力上限
	StaminaRecoverSpeed    float32 `json:"staminaRecoverSpeed"`    // 耐力回复速度
	StaminaRecoverWaitTime float32 `json:"staminaRecoverWaitTime"` // 耐力回复等待时间
	ExtraStaminaUpperLimit float32 `json:"extraStaminaUpperLimit"` // 额外耐力上限
	SprintStaminaCost      float32 `json:"sprintStaminaCost"`      // 冲刺时耐力消耗
	DashStaminaCost        float32 `json:"dashStaminaCost"`        // 猛冲时耐力消耗
}

type ConfigVehicleCombat struct {
	Property *ConfigVehicleCombatProperty `json:"property"` // 属性
}

type ConfigVehicleCombatProperty struct {
	HP          float32 `json:"HP"`          // 血量
	Attack      float32 `json:"attack"`      // 攻击力
	DefenseBase float32 `json:"defenseBase"` // 防御力
	Weight      float32 `json:"weight"`      // 重量
}

func (g *GameDataConfig) loadGadgetJsonConfig() {
	g.GadgetJsonConfigMap = make(map[string]*ConfigGadget)
	g.loadGadgetJsonConfigLoop(g.jsonPrefix + "gadget")
	logger.Info("GadgetJsonConfig Count: %v", len(g.GadgetJsonConfigMap))
}

func (g *GameDataConfig) loadGadgetJsonConfigLoop(path string) {
	fileList, err := os.ReadDir(path)
	if err != nil {
		info := fmt.Sprintf("open file error: %v, path: %v", err, path)
		panic(info)
	}
	for _, file := range fileList {
		fileName := file.Name()
		filePath := path + "/" + fileName
		if file.IsDir() {
			g.loadGadgetJsonConfigLoop(filePath)
		}
		if split := strings.Split(fileName, "."); split[len(split)-1] != "json" {
			continue
		}
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			info := fmt.Sprintf("open file error: %v", err)
			panic(info)
		}
		if fileData[0] == 0xEF && fileData[1] == 0xBB && fileData[2] == 0xBF {
			fileData = fileData[3:]
		}
		var gadgetJsonConfigMap map[string]*ConfigGadget = nil
		err = hjson.Unmarshal(fileData, &gadgetJsonConfigMap)
		if err != nil {
			logger.Info("parse file error: %v, path: %v", err, filePath)
			continue
		}
		for k, v := range gadgetJsonConfigMap {
			g.GadgetJsonConfigMap[k] = v
		}
	}
}

func GetGadgetJsonConfigByName(name string) *ConfigGadget {
	return CONF.GadgetJsonConfigMap[name]
}
