package gdconf

import (
	"hk4e/pkg/logger"
)

// WeatherTemplateData 天气模版配置表
type WeatherTemplateData struct {
	TemplateName string `csv:"天气模板名"`
	Weather      int32  `csv:"天气,omitempty"`
	Sunny        int32  `csv:"晴,omitempty"`
	Cloudy       int32  `csv:"多云,omitempty"`
	Rain         int32  `csv:"雨,omitempty"`
	ThunderStorm int32  `csv:"雷雨,omitempty"`
	Snow         int32  `csv:"雪,omitempty"`
	Mist         int32  `csv:"雾,omitempty"`
	Desert       int32  `csv:"沙漠,omitempty"`
}

func (g *GameDataConfig) loadWeatherTemplateData() {
	g.WeatherTemplateDataMap = make(map[string]map[int32]*WeatherTemplateData)
	weatherTemplateDataList := make([]*WeatherTemplateData, 0)
	readTable[WeatherTemplateData](g.txtPrefix+"WeatherTemplate.txt", &weatherTemplateDataList)
	for _, weatherTemplateData := range weatherTemplateDataList {
		_, exist := g.WeatherTemplateDataMap[weatherTemplateData.TemplateName]
		if !exist {
			g.WeatherTemplateDataMap[weatherTemplateData.TemplateName] = make(map[int32]*WeatherTemplateData)
		}
		g.WeatherTemplateDataMap[weatherTemplateData.TemplateName][weatherTemplateData.Weather] = weatherTemplateData
	}
	logger.Info("WeatherTemplateData Count: %v", len(g.WeatherTemplateDataMap))
}

func GetWeatherTemplateDataByTemplateNameAndWeather(templateName string, weather int32) *WeatherTemplateData {
	value, exist := CONF.WeatherTemplateDataMap[templateName]
	if !exist {
		return nil
	}
	return value[weather]
}

func GetWeatherTemplateDataMap() map[string]map[int32]*WeatherTemplateData {
	return CONF.WeatherTemplateDataMap
}
