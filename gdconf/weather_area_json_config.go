package gdconf

import (
	"hk4e/pkg/alg"
	"os"
	"strconv"

	"hk4e/pkg/logger"

	"github.com/hjson/hjson-go/v4"
)

type WeatherAreaJsonConfig struct {
	AreaId int32        `json:"area_id"` // 天气区域id
	Points []*AreaPoint `json:"points"`  // 多边形平面顶点数组

	VectorPoints []*alg.Vector2 `json:"-"` // 多边形平面顶点二维向量数组
}

type AreaPoint struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

func (g *GameDataConfig) loadWeatherAreaJsonConfig() {
	g.WeatherAreaJsonConfigMap = make(map[int32]map[int32]*WeatherAreaJsonConfig)
	sceneLuaPrefix := g.luaPrefix + "scene/"
	count := 0
	for _, sceneData := range g.SceneDataMap {
		sceneId := sceneData.SceneId
		sceneIdStr := strconv.Itoa(int(sceneId))
		// 读取场景天气区域
		fileData, err := os.ReadFile(sceneLuaPrefix + sceneIdStr + "/scene" + sceneIdStr + "_weather_areas.json")
		if err != nil {
			// 有些场景没有天气区域是正常情况
			// logger.Error("open file error: %v, sceneId: %v", err, sceneId)
			continue
		}
		weatherAreaJsonConfigList := make([]*WeatherAreaJsonConfig, 0)
		err = hjson.Unmarshal(fileData, &weatherAreaJsonConfigList)
		if err != nil {
			logger.Error("parse file error: %v, sceneId: %v", err, sceneId)
			continue
		}
		// 记录每个天气区域
		for _, weatherAreaJsonConfig := range weatherAreaJsonConfigList {
			weatherAreaJsonConfig.VectorPoints = make([]*alg.Vector2, 0, len(weatherAreaJsonConfig.Points))
			// 多边形平面顶点数组转换
			for _, point := range weatherAreaJsonConfig.Points {
				weatherAreaJsonConfig.VectorPoints = append(weatherAreaJsonConfig.VectorPoints, &alg.Vector2{
					X: point.X,
					Z: point.Y,
				})
			}
			_, exist := g.WeatherAreaJsonConfigMap[sceneId]
			if !exist {
				g.WeatherAreaJsonConfigMap[sceneId] = make(map[int32]*WeatherAreaJsonConfig, len(weatherAreaJsonConfig.Points))
			}
			g.WeatherAreaJsonConfigMap[sceneId][weatherAreaJsonConfig.AreaId] = weatherAreaJsonConfig
			count++
		}
	}
	logger.Info("WeatherAreaJsonConfig Count: %v", count)
}

func GetWeatherAreaMapBySceneIdAndWeatherAreaId(sceneId int32, weatherAreaId int32) *WeatherAreaJsonConfig {
	value, exist := CONF.WeatherAreaJsonConfigMap[sceneId]
	if !exist {
		return nil
	}
	return value[weatherAreaId]
}

func GetWeatherAreaMap() map[int32]map[int32]*WeatherAreaJsonConfig {
	return CONF.WeatherAreaJsonConfigMap
}
