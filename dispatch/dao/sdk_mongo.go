package dao

import (
	"context"
	"errors"
	"hk4e/dispatch/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (d *Dao) InsertSdk(sdk *model.Sdk) error {
	if d.mongo == nil {
		return d.InsertSdkGorm(sdk)
	}
	db := d.mongoDb.Collection("sdk")
	_, err := db.InsertOne(context.TODO(), sdk)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateSdk(sdk *model.Sdk) error {
	if d.mongo == nil {
		return d.UpdateSdkGorm(sdk)
	}
	db := d.mongoDb.Collection("sdk")
	_, err := db.UpdateMany(
		context.TODO(),
		bson.D{},
		bson.D{{"$set", sdk}},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QuerySdk() (*model.Sdk, error) {
	if d.mongo == nil {
		return d.QuerySdkGorm()
	}
	db := d.mongoDb.Collection("sdk")
	result := db.FindOne(
		context.TODO(),
		bson.D{},
	)
	sdk := new(model.Sdk)
	err := result.Decode(sdk)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return sdk, nil
}
