package gdconf

import (
	"fmt"
	"os"
	"strconv"

	"hk4e/pkg/logger"

	"github.com/hjson/hjson-go/v4"
)

const (
	WidgetTagTypeActionPanel = "ActionPanel"
	WidgetTagTypeFlyAttach   = "FlyAttach"
)

type WidgetJsonConfig struct {
	WidgetConfigMap map[string]*ConfigWidget `json:"widgetConfigMap"`
}

type ConfigWidget struct {
	Type              string   `json:"$type"`
	GadgetId          int32    `json:"gadgetId"`
	Tags              []string `json:"tags"`
	IsConsumeMaterial bool     `json:"isConsumeMaterial"`
	CdGroup           int32    `json:"cdGroup"`
}

func (g *GameDataConfig) loadWidgetJsonConfig() {
	g.WidgetJsonConfigMap = make(map[string]*ConfigWidget)
	filePath := g.jsonPrefix + "widget_new/ConfigWidgetNew.json"
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		info := fmt.Sprintf("open file error: %v", err)
		panic(info)
	}
	if fileData[0] == 0xEF && fileData[1] == 0xBB && fileData[2] == 0xBF {
		fileData = fileData[3:]
	}
	widgetJsonConfig := new(WidgetJsonConfig)
	err = hjson.Unmarshal(fileData, &widgetJsonConfig)
	if err != nil {
		logger.Error("parse file error: %v, path: %v", err, filePath)
		return
	}
	for k, v := range widgetJsonConfig.WidgetConfigMap {
		g.WidgetJsonConfigMap[k] = v
	}
	logger.Info("WidgetJsonConfig Count: %v", len(g.WidgetJsonConfigMap))
}

func GetWidgetJsonConfigByMaterialId(materialId int32) *ConfigWidget {
	return CONF.WidgetJsonConfigMap[strconv.Itoa(int(materialId))]
}
