package model

type WeatherInfo struct {
	WeatherAreaId     uint32 // 天气区域id
	JsonWeatherAreaId uint32 // 天气区域id json场景天气区域的id
	ClimateType       uint32 // 气候类型
}

func NewWeatherInfo() *WeatherInfo {
	return &WeatherInfo{
		WeatherAreaId:     0,
		JsonWeatherAreaId: 0,
		ClimateType:       0,
	}
}
