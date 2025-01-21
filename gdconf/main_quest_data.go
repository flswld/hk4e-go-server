package gdconf

import (
	"strconv"
	"strings"

	"hk4e/pkg/logger"
)

// MainQuestData 主线任务配置表
type MainQuestData struct {
	ParentQuestId int32  `csv:"父任务ID"`
	QuestReward   string `csv:"任务奖励RewardID,omitempty"`
	VideoKey      uint64 `csv:"VideoKey,omitempty"`

	RewardIdList []int32
}

func (g *GameDataConfig) loadMainQuestData() {
	g.MainQuestDataMap = make(map[int32]*MainQuestData)
	fileNameList := []string{
		"MainQuestData.txt",
		"MainQuestData_Exported.txt",
	}
	for _, fileName := range fileNameList {
		mainQuestDataList := make([]*MainQuestData, 0)
		readTable[MainQuestData](g.txtPrefix+fileName, &mainQuestDataList)
		for _, mainQuestData := range mainQuestDataList {
			mainQuestData.RewardIdList = make([]int32, 0)
			if mainQuestData.QuestReward != "" {
				for _, rewardIdStr := range strings.Split(mainQuestData.QuestReward, ",") {
					rewardId, err := strconv.Atoi(rewardIdStr)
					if err != nil {
						panic(err)
					}
					mainQuestData.RewardIdList = append(mainQuestData.RewardIdList, int32(rewardId))
				}
			}
			g.MainQuestDataMap[mainQuestData.ParentQuestId] = mainQuestData
		}
	}
	logger.Info("MainQuestData Count: %v", len(g.MainQuestDataMap))
}

func GetMainQuestDataById(parentQuestId int32) *MainQuestData {
	return CONF.MainQuestDataMap[parentQuestId]
}

func GetMainQuestDataMap() map[int32]*MainQuestData {
	return CONF.MainQuestDataMap
}
