package dao

import (
	"hk4e/gs/model"
)

type PlayerGorm struct {
	Uid  uint32 `gorm:"column:uid;type:bigint(20);primaryKey"`
	Data []byte `gorm:"column:data;type:longblob"`
}

func (p PlayerGorm) TableName() string {
	return "player"
}

type ChatMsgGorm struct {
	ID       uint32 `gorm:"column:id;type:bigint(20);primaryKey"`
	Sequence uint32 `gorm:"column:sequence;type:bigint(20)"`
	Time     uint32 `gorm:"column:time;type:bigint(20)"`
	Uid      uint32 `gorm:"column:uid;type:bigint(20)"`
	ToUid    uint32 `gorm:"column:to_uid;type:bigint(20)"`
	IsRead   bool   `gorm:"column:is_read;type:tinyint(1)"`
	MsgType  uint8  `gorm:"column:msg_type;type:tinyint(1)"`
	Text     string `gorm:"column:text;type:text"`
	Icon     uint32 `gorm:"column:icon;type:bigint(20)"`
	IsDelete bool   `gorm:"column:is_delete;type:tinyint(1)"`
}

func (c ChatMsgGorm) TableName() string {
	return "chat_msg"
}

type SceneBlockGorm struct {
	Uid     uint32 `gorm:"column:uid;type:bigint(20)"`
	BlockId uint32 `gorm:"column:block_id;type:bigint(20)"`
	Data    []byte `gorm:"column:data;type:longblob"`
}

func (s SceneBlockGorm) TableName() string {
	return "scene_block"
}

func (d *Dao) InsertPlayerGorm(player *model.Player) error {
	return nil
}

func (d *Dao) InsertPlayerListGorm(playerList []*model.Player) error {
	return nil
}

func (d *Dao) DeletePlayerGorm(playerId uint32) error {
	return nil
}

func (d *Dao) DeletePlayerListGorm(playerIdList []uint32) error {
	return nil
}

func (d *Dao) UpdatePlayerGorm(player *model.Player) error {
	return nil
}

func (d *Dao) UpdatePlayerListGorm(playerList []*model.Player) error {
	return nil
}

func (d *Dao) QueryPlayerByIdGorm(playerId uint32) (*model.Player, error) {
	return nil, nil
}

func (d *Dao) QueryPlayerListGorm() ([]*model.Player, error) {
	return nil, nil
}

func (d *Dao) InsertChatMsgGorm(chatMsg *model.ChatMsg) error {
	return nil
}

func (d *Dao) DeleteUpdateChatMsgByUidGorm(uid uint32) error {
	return nil
}

func (d *Dao) UpdateChatMsgByUidAndToUidActionReadGorm(uid uint32, toUid uint32) error {
	return nil
}

func (d *Dao) QueryChatMsgListByUidGorm(uid uint32) ([]*model.ChatMsg, error) {
	return nil, nil
}

func (d *Dao) InsertSceneBlockGorm(sceneBlock *model.SceneBlock) error {
	return nil
}

func (d *Dao) UpdateSceneBlockGorm(sceneBlock *model.SceneBlock) error {
	return nil
}

func (d *Dao) QuerySceneBlockByUidAndBlockIdGorm(uid uint32, blockId uint32) (*model.SceneBlock, error) {
	return nil, nil
}
