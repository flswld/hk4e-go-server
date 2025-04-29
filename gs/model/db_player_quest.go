package model

import (
	"time"

	"hk4e/common/constant"
	"hk4e/gdconf"

	"github.com/flswld/halo/logger"
)

// DbQuest 玩家任务数据
type DbQuest struct {
	QuestMap       map[uint32]*Quest       // 任务列表 key:任务id value:任务
	ParentQuestMap map[uint32]*ParentQuest // 父任务列表 key:父任务id value:父任务
}

// Quest 任务
type Quest struct {
	QuestId         uint32   // 任务id
	State           uint8    // 任务状态
	AcceptTime      uint32   // 接取时间
	StartTime       uint32   // 开始执行时间
	FinishCountList []uint32 // 任务完成进度
}

// ParentQuest 父任务
type ParentQuest struct {
	ParentQuestId uint32   // 父任务id
	State         uint8    // 任务状态
	QuestVar      [5]int32 // 任务变量
}

func (p *Player) GetDbQuest() *DbQuest {
	if p.DbQuest == nil {
		p.DbQuest = new(DbQuest)
	}
	if p.DbQuest.QuestMap == nil {
		p.DbQuest.QuestMap = make(map[uint32]*Quest)
	}
	if p.DbQuest.ParentQuestMap == nil {
		p.DbQuest.ParentQuestMap = make(map[uint32]*ParentQuest)
	}
	return p.DbQuest
}

// GetQuestMap 获取全部任务
func (q *DbQuest) GetQuestMap() map[uint32]*Quest {
	return q.QuestMap
}

// GetQuestById 获取一个任务
func (q *DbQuest) GetQuestById(questId uint32) *Quest {
	return q.QuestMap[questId]
}

// AddQuest 添加一个任务
func (q *DbQuest) AddQuest(questId uint32) {
	_, exist := q.QuestMap[questId]
	if exist {
		logger.Error("quest is already exist, questId: %v", questId)
		return
	}
	questDataConfig := gdconf.GetQuestDataById(int32(questId))
	if questDataConfig == nil {
		logger.Error("get quest data config is nil, questId: %v", questId)
		return
	}
	q.QuestMap[questId] = &Quest{
		QuestId:         uint32(questDataConfig.QuestId),
		State:           constant.QUEST_STATE_UNSTARTED,
		AcceptTime:      uint32(time.Now().Unix()),
		StartTime:       0,
		FinishCountList: make([]uint32, len(questDataConfig.FinishCondList)),
	}
	q.AddParentQuest(uint32(questDataConfig.ParentQuestId))
}

// StartQuest 开始执行一个任务
func (q *DbQuest) StartQuest(questId uint32) {
	quest, exist := q.QuestMap[questId]
	if !exist {
		logger.Error("get quest is nil, questId: %v", questId)
		return
	}
	if quest.State != constant.QUEST_STATE_UNSTARTED {
		logger.Error("invalid quest state, questId: %v, state: %v", questId, quest.State)
		return
	}
	quest.State = constant.QUEST_STATE_UNFINISHED
	quest.StartTime = uint32(time.Now().Unix())
}

// DeleteQuest 删除一个任务
func (q *DbQuest) DeleteQuest(questId uint32) {
	_, exist := q.QuestMap[questId]
	if !exist {
		logger.Error("quest is not exist, questId: %v", questId)
		return
	}
	delete(q.QuestMap, questId)
}

// AddQuestFinishCount 添加一个任务的完成进度
func (q *DbQuest) AddQuestFinishCount(questId uint32, index int) {
	quest, exist := q.QuestMap[questId]
	if !exist {
		logger.Error("get quest is nil, questId: %v", questId)
		return
	}
	if quest.State != constant.QUEST_STATE_UNFINISHED {
		return
	}
	if index >= len(quest.FinishCountList) {
		logger.Error("invalid quest cond index, questId: %v, index: %v", questId, index)
		return
	}
	quest.FinishCountList[index] += 1
}

// CheckQuestFinish 检查任务是否完成
func (q *DbQuest) CheckQuestFinish(questId uint32) {
	quest, exist := q.QuestMap[questId]
	if !exist {
		logger.Error("get quest is nil, questId: %v", questId)
		return
	}
	if quest.State != constant.QUEST_STATE_UNFINISHED {
		return
	}
	questDataConfig := gdconf.GetQuestDataById(int32(questId))
	if questDataConfig == nil {
		logger.Error("get quest data config is nil, questId: %v", questId)
		return
	}
	resultList := make([]bool, 0)
	for index, finishCond := range questDataConfig.FinishCondList {
		result := false
		finishCount := finishCond.Count
		if finishCount == 0 {
			finishCount = 1
		}
		if quest.FinishCountList[index] >= uint32(finishCount) {
			result = true
		}
		resultList = append(resultList, result)
	}
	finish := false
	switch questDataConfig.FinishCondCompose {
	case constant.QUEST_LOGIC_TYPE_NONE:
		fallthrough
	case constant.QUEST_LOGIC_TYPE_AND:
		finish = true
		for _, result := range resultList {
			if !result {
				finish = false
				break
			}
		}
	case constant.QUEST_LOGIC_TYPE_OR:
		finish = false
		for _, result := range resultList {
			if result {
				finish = true
				break
			}
		}
	case constant.QUEST_LOGIC_TYPE_A_AND_ETCOR:
		if len(resultList) < 2 {
			finish = false
			break
		}
		finishA := resultList[0]
		finishEtc := false
		for _, result := range resultList[1:] {
			if result {
				finishEtc = true
				break
			}
		}
		finish = finishA && finishEtc
	case constant.QUEST_LOGIC_TYPE_A_AND_B_AND_ETCOR:
		if len(resultList) < 3 {
			finish = false
			break
		}
		finishA := resultList[0]
		finishB := resultList[1]
		finishEtc := false
		for _, result := range resultList[2:] {
			if result {
				finishEtc = true
				break
			}
		}
		finish = finishA && finishB && finishEtc
	case constant.QUEST_LOGIC_TYPE_A_OR_ETCAND:
		if len(resultList) < 2 {
			finish = false
			break
		}
		finishA := resultList[0]
		finishEtc := true
		for _, result := range resultList[1:] {
			if !result {
				finishEtc = false
				break
			}
		}
		finish = finishA || finishEtc
	case constant.QUEST_LOGIC_TYPE_A_OR_B_OR_ETCAND:
		if len(resultList) < 3 {
			finish = false
			break
		}
		finishA := resultList[0]
		finishB := resultList[1]
		finishEtc := true
		for _, result := range resultList[2:] {
			if !result {
				finishEtc = false
				break
			}
		}
		finish = finishA || finishB || finishEtc
	default:
		logger.Error("not support quest finish cond logic type: %v, questId: %v", questDataConfig.FinishCondCompose, questId)
	}
	if finish {
		quest.State = constant.QUEST_STATE_FINISHED
		q.CheckParentQuestFinish(uint32(questDataConfig.ParentQuestId))
	}
}

// ForceFinishQuest 强制完成一个任务
func (q *DbQuest) ForceFinishQuest(questId uint32) {
	quest, exist := q.QuestMap[questId]
	if !exist {
		logger.Error("get quest is nil, questId: %v", questId)
		return
	}
	quest.State = constant.QUEST_STATE_FINISHED
	questDataConfig := gdconf.GetQuestDataById(int32(questId))
	if questDataConfig == nil {
		logger.Error("get quest data config is nil, questId: %v", questId)
		return
	}
	q.CheckParentQuestFinish(uint32(questDataConfig.ParentQuestId))
}

// FailQuest 失败一个任务
func (q *DbQuest) FailQuest(questId uint32) {
	quest, exist := q.QuestMap[questId]
	if !exist {
		logger.Error("get quest is nil, questId: %v", questId)
		return
	}
	if quest.State != constant.QUEST_STATE_UNFINISHED {
		return
	}
	quest.State = constant.QUEST_STATE_FAILED
	questDataConfig := gdconf.GetQuestDataById(int32(questId))
	if questDataConfig == nil {
		logger.Error("get quest data config is nil, questId: %v", questId)
		return
	}
	quest.FinishCountList = make([]uint32, len(questDataConfig.FinishCondList))
}

// GetParentQuestMap 获取全部父任务
func (q *DbQuest) GetParentQuestMap() map[uint32]*ParentQuest {
	return q.ParentQuestMap
}

// GetParentQuestById 获取一个父任务
func (q *DbQuest) GetParentQuestById(parentQuestId uint32) *ParentQuest {
	return q.ParentQuestMap[parentQuestId]
}

// AddParentQuest 添加一个父任务
func (q *DbQuest) AddParentQuest(parentQuestId uint32) {
	_, exist := q.ParentQuestMap[parentQuestId]
	if exist {
		return
	}
	q.ParentQuestMap[parentQuestId] = &ParentQuest{
		ParentQuestId: parentQuestId,
		State:         constant.PARENT_QUEST_STATE_NONE,
		QuestVar:      [5]int32{0, 0, 0, 0, 0},
	}
}

// CheckParentQuestFinish 检查父任务是否完成
func (q *DbQuest) CheckParentQuestFinish(parentQuestId uint32) {
	parentQuest, exist := q.ParentQuestMap[parentQuestId]
	if !exist {
		logger.Error("get parent quest is nil, parentQuestId: %v", parentQuestId)
		return
	}
	finish := true
	questDataMap := gdconf.GetQuestDataMapByParentQuestId(int32(parentQuestId))
	for _, questData := range questDataMap {
		quest, exist := q.QuestMap[uint32(questData.QuestId)]
		if !exist {
			finish = false
			break
		}
		if quest.State != constant.QUEST_STATE_FINISHED {
			finish = false
			break
		}
	}
	if finish {
		parentQuest.State = constant.PARENT_QUEST_STATE_FINISHED
	}
}

// ForceFinishParentQuest 强制完成一个父任务
func (q *DbQuest) ForceFinishParentQuest(parentQuestId uint32) {
	parentQuest, exist := q.ParentQuestMap[parentQuestId]
	if !exist {
		logger.Error("get parent quest is nil, parentQuestId: %v", parentQuestId)
		return
	}
	parentQuest.State = constant.PARENT_QUEST_STATE_FINISHED
}
