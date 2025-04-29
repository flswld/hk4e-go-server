package gdconf

import (
	"fmt"
	"os"

	"github.com/flswld/halo/logger"
	"github.com/hjson/hjson-go/v4"
)

type ConfigGlobalCombat struct {
	DefaultAbilities struct {
		DefaultAvatarAbilities []string `json:"defaultAvatarAbilities"`
	} `json:"defaultAbilities"`
}

func (g *GameDataConfig) loadDefaultAbilityJsonConfig() {
	CONF.DefaultAbilityNameList = make([]string, 0)
	filePath := g.jsonPrefix + "common/ConfigGlobalCombat.json"
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		info := fmt.Sprintf("open file error: %v", err)
		panic(info)
	}
	if fileData[0] == 0xEF && fileData[1] == 0xBB && fileData[2] == 0xBF {
		fileData = fileData[3:]
	}
	configGlobalCombat := new(ConfigGlobalCombat)
	err = hjson.Unmarshal(fileData, configGlobalCombat)
	if err != nil {
		logger.Info("parse file error: %v, path: %v", err, filePath)
		panic(err)
	}
	for _, defaultAbilityName := range configGlobalCombat.DefaultAbilities.DefaultAvatarAbilities {
		CONF.DefaultAbilityNameList = append(CONF.DefaultAbilityNameList, defaultAbilityName)
	}
}

func GetDefaultAbilityNameList() []string {
	return CONF.DefaultAbilityNameList
}
