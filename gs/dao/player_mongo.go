package dao

import (
	"context"
	"errors"

	"hk4e/gs/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MaxQueryChatMsgLen = 1000 // 最大可查询聊天记录条数
)

func (d *Dao) InsertPlayer(player *model.Player) error {
	if d.mongo == nil {
		return d.InsertPlayerGorm(player)
	}
	db := d.mongoDb.Collection("player")
	_, err := db.InsertOne(context.TODO(), player)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) InsertPlayerList(playerList []*model.Player) error {
	if d.mongo == nil {
		return d.InsertPlayerListGorm(playerList)
	}
	if len(playerList) == 0 {
		return nil
	}
	db := d.mongoDb.Collection("player")
	modelOperateList := make([]mongo.WriteModel, 0)
	for _, player := range playerList {
		modelOperate := mongo.NewInsertOneModel().SetDocument(player)
		modelOperateList = append(modelOperateList, modelOperate)
	}
	_, err := db.BulkWrite(context.TODO(), modelOperateList)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) DeletePlayer(playerId uint32) error {
	if d.mongo == nil {
		return d.DeletePlayerGorm(playerId)
	}
	db := d.mongoDb.Collection("player")
	_, err := db.DeleteOne(context.TODO(), bson.D{{"player_id", playerId}})
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) DeletePlayerList(playerIdList []uint32) error {
	if d.mongo == nil {
		return d.DeletePlayerListGorm(playerIdList)
	}
	if len(playerIdList) == 0 {
		return nil
	}
	db := d.mongoDb.Collection("player")
	modelOperateList := make([]mongo.WriteModel, 0)
	for _, playerId := range playerIdList {
		modelOperate := mongo.NewDeleteOneModel().SetFilter(bson.D{{"player_id", playerId}})
		modelOperateList = append(modelOperateList, modelOperate)
	}
	_, err := db.BulkWrite(context.TODO(), modelOperateList)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdatePlayer(player *model.Player) error {
	if d.mongo == nil {
		return d.UpdatePlayerGorm(player)
	}
	db := d.mongoDb.Collection("player")
	_, err := db.UpdateMany(
		context.TODO(),
		bson.D{{"player_id", player.PlayerId}},
		bson.D{{"$set", player}},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdatePlayerList(playerList []*model.Player) error {
	if d.mongo == nil {
		return d.UpdatePlayerListGorm(playerList)
	}
	if len(playerList) == 0 {
		return nil
	}
	db := d.mongoDb.Collection("player")
	modelOperateList := make([]mongo.WriteModel, 0)
	for _, player := range playerList {
		modelOperate := mongo.NewUpdateManyModel().SetFilter(bson.D{{"player_id", player.PlayerId}}).SetUpdate(bson.D{{"$set", player}})
		modelOperateList = append(modelOperateList, modelOperate)
	}
	_, err := db.BulkWrite(context.TODO(), modelOperateList)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QueryPlayerById(playerId uint32) (*model.Player, error) {
	if d.mongo == nil {
		return d.QueryPlayerByIdGorm(playerId)
	}
	db := d.mongoDb.Collection("player")
	result := db.FindOne(
		context.TODO(),
		bson.D{{"player_id", playerId}},
	)
	player := new(model.Player)
	err := result.Decode(player)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return player, nil
}

func (d *Dao) QueryPlayerList() ([]*model.Player, error) {
	if d.mongo == nil {
		return d.QueryPlayerListGorm()
	}
	db := d.mongoDb.Collection("player")
	find, err := db.Find(
		context.TODO(),
		bson.D{},
	)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Player, 0)
	for find.Next(context.TODO()) {
		item := new(model.Player)
		err = find.Decode(item)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (d *Dao) InsertChatMsg(chatMsg *model.ChatMsg) error {
	if d.mongo == nil {
		return d.InsertChatMsgGorm(chatMsg)
	}
	db := d.mongoDb.Collection("chat_msg")
	_, err := db.InsertOne(context.TODO(), chatMsg)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) DeleteUpdateChatMsgByUid(uid uint32) error {
	if d.mongo == nil {
		return d.DeleteUpdateChatMsgByUidGorm(uid)
	}
	db := d.mongoDb.Collection("chat_msg")
	_, err := db.UpdateMany(
		context.TODO(),
		bson.D{{"$or", []bson.D{{{"to_uid", uid}}, {{"uid", uid}}}}},
		bson.D{{"$set", bson.D{{"is_delete", true}}}},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateChatMsgByUidAndToUidActionRead(uid uint32, toUid uint32) error {
	if d.mongo == nil {
		return d.UpdateChatMsgByUidAndToUidActionReadGorm(uid, toUid)
	}
	db := d.mongoDb.Collection("chat_msg")
	_, err := db.UpdateMany(
		context.TODO(),
		bson.D{{"to_uid", uid}, {"uid", toUid}},
		bson.D{{"$set", bson.D{{"is_read", true}}}},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QueryChatMsgListByUid(uid uint32) ([]*model.ChatMsg, error) {
	if d.mongo == nil {
		return d.QueryChatMsgListByUidGorm(uid)
	}
	db := d.mongoDb.Collection("chat_msg")
	result := make([]*model.ChatMsg, 0)
	find, err := db.Find(
		context.TODO(),
		bson.D{
			{"$and", []bson.D{
				{{"$or", []bson.D{{{"to_uid", uid}}, {{"uid", uid}}}}},
				{{"is_delete", false}},
			}},
		},
		options.Find().SetSort(bson.M{"time": -1}),
		options.Find().SetLimit(MaxQueryChatMsgLen),
	)
	if err != nil {
		return nil, err
	}
	for find.Next(context.TODO()) {
		item := new(model.ChatMsg)
		err = find.Decode(item)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (d *Dao) InsertSceneBlock(sceneBlock *model.SceneBlock) error {
	if d.mongo == nil {
		return d.InsertSceneBlockGorm(sceneBlock)
	}
	db := d.mongoDb.Collection("scene_block")
	_, err := db.InsertOne(context.TODO(), sceneBlock)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateSceneBlock(sceneBlock *model.SceneBlock) error {
	if d.mongo == nil {
		return d.UpdateSceneBlockGorm(sceneBlock)
	}
	db := d.mongoDb.Collection("scene_block")
	_, err := db.UpdateMany(
		context.TODO(),
		bson.D{{"_id", sceneBlock.ID}},
		bson.D{{"$set", sceneBlock}},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QuerySceneBlockByUidAndBlockId(uid uint32, blockId uint32) (*model.SceneBlock, error) {
	if d.mongo == nil {
		return d.QuerySceneBlockByUidAndBlockIdGorm(uid, blockId)
	}
	db := d.mongoDb.Collection("scene_block")
	result := db.FindOne(
		context.TODO(),
		bson.D{{"$and", []bson.D{{{"uid", uid}}, {{"block_id", blockId}}}}},
	)
	sceneBlock := new(model.SceneBlock)
	err := result.Decode(sceneBlock)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return sceneBlock, nil
}
