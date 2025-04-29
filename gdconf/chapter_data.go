package gdconf

import (
	"github.com/flswld/halo/logger"
)

// ChapterData 章节配置表
type ChapterData struct {
	ChapterId    int32 `csv:"章节ID"`
	StartQuestId int32 `csv:"开始子任务,omitempty"`
	EndQuestId   int32 `csv:"结束子任务,omitempty"`
}

func (g *GameDataConfig) loadChapterData() {
	g.ChapterDataMap = make(map[int32]*ChapterData)
	fileNameList := []string{
		"ChapterData.txt",
		"ChapterData_Exported.txt",
	}
	for _, fileName := range fileNameList {
		chapterDataList := make([]*ChapterData, 0)
		readTable[ChapterData](g.txtPrefix+fileName, &chapterDataList)
		for _, chapterData := range chapterDataList {
			g.ChapterDataMap[chapterData.ChapterId] = chapterData
		}
	}
	logger.Info("ChapterData Count: %v", len(g.ChapterDataMap))
}

func GetChapterDataById(chapterId int32) *ChapterData {
	return CONF.ChapterDataMap[chapterId]
}

func GetChapterDataMap() map[int32]*ChapterData {
	return CONF.ChapterDataMap
}
