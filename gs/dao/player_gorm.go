package dao

import (
	"errors"

	"hk4e/gs/model"

	"github.com/vmihailenco/msgpack/v5"
	"gorm.io/gorm"
)

type PlayerGorm struct {
	Uid  uint32 `gorm:"column:uid;type:bigint(20);primaryKey"`
	Data []byte `gorm:"column:data;type:longblob"`
}

func (p PlayerGorm) TableName() string {
	return "player"
}

type ChatMsgGorm struct {
	ID       uint32 `gorm:"column:id;type:bigint(20);primaryKey;autoIncrement"`
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
	data, err := msgpack.Marshal(player)
	if err != nil {
		return err
	}
	err = d.gormDb.Create(&PlayerGorm{
		Uid:  player.PlayerId,
		Data: data,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) InsertPlayerListGorm(playerList []*model.Player) error {
	for _, player := range playerList {
		err := d.InsertPlayerGorm(player)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dao) DeletePlayerGorm(playerId uint32) error {
	d.gormDb.Where("uid = ?", playerId).Delete(&PlayerGorm{})
	return nil
}

func (d *Dao) DeletePlayerListGorm(playerIdList []uint32) error {
	for _, playerId := range playerIdList {
		err := d.DeletePlayerGorm(playerId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dao) UpdatePlayerGorm(player *model.Player) error {
	data, err := msgpack.Marshal(player)
	if err != nil {
		return err
	}
	err = d.gormDb.Updates(&PlayerGorm{
		Uid:  player.PlayerId,
		Data: data,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdatePlayerListGorm(playerList []*model.Player) error {
	for _, player := range playerList {
		err := d.UpdatePlayerGorm(player)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dao) QueryPlayerByIdGorm(playerId uint32) (*model.Player, error) {
	playerGorm := new(PlayerGorm)
	err := d.gormDb.Where("uid = ?", playerId).First(playerGorm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	player := new(model.Player)
	err = msgpack.Unmarshal(playerGorm.Data, player)
	if err != nil {
		return nil, err
	}
	return player, nil
}

func (d *Dao) QueryPlayerListGorm() ([]*model.Player, error) {
	var playerGormList []*PlayerGorm = nil
	err := d.gormDb.Find(&playerGormList).Error
	if err != nil {
		return nil, err
	}
	playerList := make([]*model.Player, 0)
	for _, playerGorm := range playerGormList {
		player := new(model.Player)
		err = msgpack.Unmarshal(playerGorm.Data, player)
		if err != nil {
			return nil, err
		}
		playerList = append(playerList, player)
	}
	return playerList, nil
}

func (d *Dao) InsertChatMsgGorm(chatMsg *model.ChatMsg) error {
	err := d.gormDb.Create(&ChatMsgGorm{
		Sequence: chatMsg.Sequence,
		Time:     chatMsg.Time,
		Uid:      chatMsg.Uid,
		ToUid:    chatMsg.ToUid,
		IsRead:   chatMsg.IsRead,
		MsgType:  chatMsg.MsgType,
		Text:     chatMsg.Text,
		Icon:     chatMsg.Icon,
		IsDelete: chatMsg.IsDelete,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) DeleteUpdateChatMsgByUidGorm(uid uint32) error {
	err := d.gormDb.Model(&ChatMsgGorm{}).Where("to_uid = ? or uid = ?", uid, uid).Update("is_delete", true).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateChatMsgByUidAndToUidActionReadGorm(uid uint32, toUid uint32) error {
	err := d.gormDb.Model(&ChatMsgGorm{}).Where("to_uid = ? and uid = ?", uid, toUid).Update("is_read", true).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QueryChatMsgListByUidGorm(uid uint32) ([]*model.ChatMsg, error) {
	var chatMsgGormList []*ChatMsgGorm = nil
	err := d.gormDb.Where("to_uid = ? or uid = ? and is_delete = ?", uid, uid, false).Find(&chatMsgGormList).
		Order("time DESC").Limit(MaxQueryChatMsgLen).Error
	if err != nil {
		return nil, err
	}
	chatMsgList := make([]*model.ChatMsg, 0)
	for _, chatMsgGorm := range chatMsgGormList {
		chatMsgList = append(chatMsgList, &model.ChatMsg{
			Sequence: chatMsgGorm.Sequence,
			Time:     chatMsgGorm.Time,
			Uid:      chatMsgGorm.Uid,
			ToUid:    chatMsgGorm.ToUid,
			IsRead:   chatMsgGorm.IsRead,
			MsgType:  chatMsgGorm.MsgType,
			Text:     chatMsgGorm.Text,
			Icon:     chatMsgGorm.Icon,
			IsDelete: chatMsgGorm.IsDelete,
		})
	}
	return chatMsgList, nil
}

func (d *Dao) InsertSceneBlockGorm(sceneBlock *model.SceneBlock) error {
	data, err := msgpack.Marshal(sceneBlock)
	if err != nil {
		return err
	}
	err = d.gormDb.Create(&SceneBlockGorm{
		Uid:     sceneBlock.Uid,
		BlockId: sceneBlock.BlockId,
		Data:    data,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateSceneBlockGorm(sceneBlock *model.SceneBlock) error {
	data, err := msgpack.Marshal(sceneBlock)
	if err != nil {
		return err
	}
	err = d.gormDb.Updates(&SceneBlockGorm{
		Uid:     sceneBlock.Uid,
		BlockId: sceneBlock.BlockId,
		Data:    data,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QuerySceneBlockByUidAndBlockIdGorm(uid uint32, blockId uint32) (*model.SceneBlock, error) {
	sceneBlockGorm := new(SceneBlockGorm)
	err := d.gormDb.Where("uid = ? and block_id = ?", uid, blockId).First(sceneBlockGorm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	sceneBlock := new(model.SceneBlock)
	err = msgpack.Unmarshal(sceneBlockGorm.Data, sceneBlock)
	if err != nil {
		return nil, err
	}
	return sceneBlock, nil
}
