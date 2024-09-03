package dao

import (
	"context"

	"hk4e/dispatch/model"
	"hk4e/pkg/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (d *Dao) InsertSdkAccount(account *model.SdkAccount) (primitive.ObjectID, error) {
	db := d.db.Collection("sdk_account")
	id, err := db.InsertOne(context.TODO(), account)
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

func (d *Dao) UpdateSdkAccountFieldByFieldName(fieldName string, fieldValue any, fieldUpdateName string, fieldUpdateValue any) (int64, error) {
	db := d.db.Collection("sdk_account")
	updateCount, err := db.UpdateMany(
		context.TODO(),
		bson.D{
			{fieldName, fieldValue},
		},
		bson.D{
			{"$set", bson.D{
				{fieldUpdateName, fieldUpdateValue},
			}},
		},
	)
	if err != nil {
		return 0, err
	} else {
		return updateCount.ModifiedCount, nil
	}
}

func (d *Dao) QuerySdkAccountByField(fieldName string, fieldValue any) (*model.SdkAccount, error) {
	db := d.db.Collection("sdk_account")
	find, err := db.Find(
		context.TODO(),
		bson.D{
			{fieldName, fieldValue},
		},
	)
	if err != nil {
		return nil, err
	}
	result := make([]*model.SdkAccount, 0)
	for find.Next(context.TODO()) {
		item := new(model.SdkAccount)
		err := find.Decode(item)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if len(result) == 0 {
		return nil, nil
	} else {
		return result[0], nil
	}
}
