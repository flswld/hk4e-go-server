package gdconf

import (
	"hk4e/pkg/logger"
)

// WeatherData 天气配置表
type WeatherData struct {
	WeatherAreaId     int32  `csv:"ID"`
	JsonWeatherAreaId int32  `csv:"JSON天气区域ID,omitempty"`
	MaxHeight         int32  `csv:"最大高度,omitempty"`
	GadgetID          int32  `csv:"GadgetID,omitempty"`
	DefaultOpen       int32  `csv:"默认是否开启,omitempty"`
	TemplateName      string `csv:"TemplateName,omitempty"`
	Priority          int32  `csv:"Priority,omitempty"`
	DefaultWeather    int32  `csv:"DefaultWeather,omitempty"`
	UseDefaultWeather int32  `csv:"是否固定使用DefaultWeather,omitempty"`
	SceneId           int32  `csv:"场景ID,omitempty"`
}

func (g *GameDataConfig) loadWeatherData() {
	g.WeatherDataMap = make(map[int32]*WeatherData)
	g.WeatherDataJsonMap = make(map[int32]map[int32]*WeatherData)
	weatherDataList := make([]*WeatherData, 0)
	readTable[WeatherData](g.txtPrefix+"WeatherData.txt", &weatherDataList)
	for _, weatherData := range weatherDataList {
		g.WeatherDataMap[weatherData.WeatherAreaId] = weatherData
		// json的天气区域id格式
		_, exist := g.WeatherDataJsonMap[weatherData.JsonWeatherAreaId]
		if !exist {
			g.WeatherDataJsonMap[weatherData.JsonWeatherAreaId] = make(map[int32]*WeatherData)
		}
		g.WeatherDataJsonMap[weatherData.JsonWeatherAreaId][weatherData.WeatherAreaId] = weatherData
	}
	logger.Info("WeatherData Count: %v", len(g.WeatherDataMap))
}

func GetWeatherDataMapByJsonWeatherAreaId(jsonWeatherAreaId int32) map[int32]*WeatherData {
	value, exist := CONF.WeatherDataJsonMap[jsonWeatherAreaId]
	if !exist {
		return nil
	}
	return value
}

func GetWeatherDataByWeatherAreaId(weatherAreaId int32) *WeatherData {
	value, exist := CONF.WeatherDataMap[weatherAreaId]
	if !exist {
		return nil
	}
	return value
}

func GetWeatherDataMap() map[int32]*WeatherData {
	return CONF.WeatherDataMap
}

func GetWeatherDataJsonMap() map[int32]map[int32]*WeatherData {
	return CONF.WeatherDataJsonMap
}
