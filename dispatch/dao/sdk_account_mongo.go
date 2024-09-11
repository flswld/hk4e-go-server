package dao

import (
	"context"

	"hk4e/dispatch/model"

	"go.mongodb.org/mongo-driver/bson"
)

func (d *Dao) InsertSdkAccount(account *model.SdkAccount) error {
	if d.mongo == nil {
		return d.InsertSdkAccountGorm(account)
	}
	db := d.mongoDb.Collection("sdk_account")
	_, err := db.InsertOne(context.TODO(), account)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateSdkAccountFieldByFieldName(fieldName string, fieldValue any, fieldUpdateName string, fieldUpdateValue any) error {
	if d.mongo == nil {
		return d.UpdateSdkAccountFieldByFieldNameGorm(fieldName, fieldValue, fieldUpdateName, fieldUpdateValue)
	}
	db := d.mongoDb.Collection("sdk_account")
	_, err := db.UpdateMany(
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
		return err
	}
	return nil
}

func (d *Dao) QuerySdkAccountByField(fieldName string, fieldValue any) (*model.SdkAccount, error) {
	if d.mongo == nil {
		return d.QuerySdkAccountByFieldGorm(fieldName, fieldValue)
	}
	db := d.mongoDb.Collection("sdk_account")
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
