package gdconf

import (
	"fmt"
	"hk4e/pkg/logger"
	"os"
	"strings"

	"github.com/hjson/hjson-go/v4"
)

type ConfigGadget struct {
	Abilities []*ConfigAbility `json:"abilities"`
}

func (g *GameDataConfig) loadGadgetJsonConfig() {
	g.ConfigGadgetMap = make(map[string]*ConfigGadget)
	g.loadGadgetJsonConfigLoop(g.jsonPrefix + "gadget")
	logger.Info("ConfigGadgetMap Count: %v", len(g.ConfigGadgetMap))
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
		var configGadgetMap map[string]*ConfigGadget = nil
		err = hjson.Unmarshal(fileData, &configGadgetMap)
		if err != nil {
			logger.Info("parse file error: %v, path: %v", err, path+"/"+fileName)
			continue
		}
		for k, v := range configGadgetMap {
			g.ConfigGadgetMap[k] = v
		}
	}
}
