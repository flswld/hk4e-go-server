package gdconf

import (
	"fmt"
	"hk4e/pkg/logger"
	"os"
	"strings"

	"github.com/hjson/hjson-go/v4"
)

type GadgetJsonConfig struct {
	Abilities []*ConfigAbility `json:"abilities"`
}

func (g *GameDataConfig) loadGadgetJsonConfig() {
	g.GadgetJsonConfigMap = make(map[string]*GadgetJsonConfig)
	g.loadGadgetJsonConfigLoop(g.jsonPrefix + "gadget")
	logger.Info("GadgetJsonConfigMap Count: %v", len(g.GadgetJsonConfigMap))
}

func (g *GameDataConfig) loadGadgetJsonConfigLoop(path string) {
	fileList, err := os.ReadDir(path)
	if err != nil {
		info := fmt.Sprintf("open file error: %v, path: %v", err, path)
		panic(info)
	}
	for _, file := range fileList {
		fileName := file.Name()
		if file.IsDir() {
			g.loadGadgetJsonConfigLoop(path + "/" + fileName)
		}
		if split := strings.Split(fileName, "."); split[len(split)-1] != "json" {
			continue
		}
		fileData, err := os.ReadFile(path + "/" + fileName)
		if err != nil {
			info := fmt.Sprintf("open file error: %v, path: %v", err, path+"/"+fileName)
			panic(info)
		}
		var configGadgetMap map[string]*GadgetJsonConfig = nil
		err = hjson.Unmarshal(fileData, &configGadgetMap)
		if err != nil {
			logger.Info("parse file error: %v, path: %v", err, path+"/"+fileName)
			continue
		}
		for k, v := range configGadgetMap {
			g.GadgetJsonConfigMap[k] = v
		}
	}
}

func GetGadgetJsonConfigByName(name string) *GadgetJsonConfig {
	return CONF.GadgetJsonConfigMap[name]
}
