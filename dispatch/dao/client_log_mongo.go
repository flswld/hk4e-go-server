package dao

import (
	"context"

	"hk4e/dispatch/model"

	"github.com/flswld/halo/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (d *Dao) InsertClientLog(clientLog *model.ClientLog) (primitive.ObjectID, error) {
	if d.mongo == nil {
		return primitive.ObjectID{}, nil
	}
	db := d.mongoDb.Collection("client_log")
	id, err := db.InsertOne(context.TODO(), clientLog)
	if err != nil {
		return primitive.ObjectID{}, err
	} else {
		_id, ok := id.InsertedID.(primitive.ObjectID)
		if !ok {
			logger.Error("get insert id error")
			return primitive.ObjectID{}, nil
		}
		return _id, nil
	}
}
