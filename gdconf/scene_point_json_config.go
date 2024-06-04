package gdconf

import (
	"encoding/json"
	"os"
	"strconv"

	"hk4e/pkg/logger"
)

// 场景传送点配置数据

// 传送点类型
const (
	PointTypeStrTransPointNormal       = "TransPointNormal"
	PointTypeStrTransPointPortal       = "TransPointPortal"
	PointTypeStrTransPointStatue       = "TransPointStatue"
	PointTypeStrDungeonEntry           = "DungeonEntry"
	PointTypeStrDungeonExit            = "DungeonExit"
	PointTypeStrDungeonQuitPoint       = "DungeonQuitPoint"
	PointTypeStrDungeonWayPoint        = "DungeonWayPoint"
	PointTypeStrDungeonSlipRevivePoint = "DungeonSlipRevivePoint"
	PointTypeStrSceneBuildingPoint     = "SceneBuildingPoint"
	PointTypeStrPersonalSceneJumpPoint = "PersonalSceneJumpPoint"
	PointTypeStrVehicleSummonPoint     = "VehicleSummonPoint"
	PointTypeStrOther                  = "Other"
)

const (
	PointTypeTransPointNormal = iota // X
	PointTypeTransPointPortal
	PointTypeTransPointStatue // X
	PointTypeDungeonEntry     // X
	PointTypeDungeonExit
	PointTypeDungeonQuitPoint
	PointTypeDungeonWayPoint
	PointTypeDungeonSlipRevivePoint
	PointTypeSceneBuildingPoint     // X
	PointTypePersonalSceneJumpPoint // X
	PointTypeVehicleSummonPoint     // X
	PointTypeOther
)

type ScenePointJsonConfig struct {
	Points   map[string]*PointData `json:"points"`
	PointMap map[int32]*PointData  `json:"-"`
}

type PointData struct {
	Id                int32     `json:"-"`
	PointType         int       `json:"-"`
	PointTypeStr      string    `json:"pointType"`         // 传送点类型
	TranPos           *Position `json:"tranPos"`           // 传送后位置
	TranRot           *Position `json:"tranRot"`           // 传送后朝向
	DungeonIds        []int32   `json:"dungeonIds"`        // 地牢id列表
	DungeonRandomList []int32   `json:"dungeonRandomList"` // 随机地牢id列表
	TranSceneId       int32     `json:"tranSceneId"`       // 跳转到场景id
	IsModelHidden     bool      `json:"isModelHidden"`     // 是否为隐藏传送点
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func (g *GameDataConfig) loadScenePointJsonConfig() {
	g.ScenePointJsonConfigMap = make(map[int32]*ScenePointJsonConfig)
	sceneLuaPrefix := g.luaPrefix + "scene/"
	for _, sceneData := range g.SceneDataMap {
		sceneId := sceneData.SceneId
		sceneIdStr := strconv.Itoa(int(sceneId))
		fileData, err := os.ReadFile(sceneLuaPrefix + sceneIdStr + "/scene" + sceneIdStr + "_point.json")
		if err != nil {
			logger.Info("open file error: %v, sceneId: %v", err, sceneId)
			continue
		}
		scenePointJsonConfig := new(ScenePointJsonConfig)
		err = json.Unmarshal(fileData, scenePointJsonConfig)
		if err != nil {
			logger.Error("parse file error: %v", err)
			continue
		}
		scenePointJsonConfig.PointMap = make(map[int32]*PointData)
		for pointIdStr, pointData := range scenePointJsonConfig.Points {
			pointId, err := strconv.Atoi(pointIdStr)
			if err != nil {
				logger.Error("parse file error: %v", err)
				continue
			}
			pointData.Id = int32(pointId)
			switch pointData.PointTypeStr {
			case PointTypeStrTransPointNormal:
				pointData.PointType = PointTypeTransPointNormal
			case PointTypeStrTransPointPortal:
				pointData.PointType = PointTypeTransPointPortal
			case PointTypeStrTransPointStatue:
				pointData.PointType = PointTypeTransPointStatue
			case PointTypeStrDungeonEntry:
				pointData.PointType = PointTypeDungeonEntry
			case PointTypeStrDungeonExit:
				pointData.PointType = PointTypeDungeonExit
			case PointTypeStrDungeonQuitPoint:
				pointData.PointType = PointTypeDungeonQuitPoint
			case PointTypeStrDungeonWayPoint:
				pointData.PointType = PointTypeDungeonWayPoint
			case PointTypeStrDungeonSlipRevivePoint:
				pointData.PointType = PointTypeDungeonSlipRevivePoint
			case PointTypeStrSceneBuildingPoint:
				pointData.PointType = PointTypeSceneBuildingPoint
			case PointTypeStrPersonalSceneJumpPoint:
				pointData.PointType = PointTypePersonalSceneJumpPoint
			case PointTypeStrVehicleSummonPoint:
				pointData.PointType = PointTypeVehicleSummonPoint
			case PointTypeStrOther:
				pointData.PointType = PointTypeOther
			default:
				logger.Info("unknown scene point type: %v", pointData.PointTypeStr)
				pointData.PointType = PointTypeOther
			}
			scenePointJsonConfig.PointMap[int32(pointId)] = pointData
		}
		g.ScenePointJsonConfigMap[sceneId] = scenePointJsonConfig
	}
	scenePointCount := 0
	for _, scenePoint := range g.ScenePointJsonConfigMap {
		scenePointCount += len(scenePoint.PointMap)
	}
	logger.Info("ScenePointJsonConfig Count: %v", scenePointCount)
}

func GetScenePointBySceneIdAndPointId(sceneId int32, pointId int32) *PointData {
	value, exist := CONF.ScenePointJsonConfigMap[sceneId]
	if !exist {
		return nil
	}
	return value.PointMap[pointId]
}

func GetScenePointMapBySceneId(sceneId int32) map[int32]*PointData {
	value, exist := CONF.ScenePointJsonConfigMap[sceneId]
	if !exist {
		return nil
	}
	return value.PointMap
}
