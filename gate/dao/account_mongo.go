package dao

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Account struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	OpenId        string             `bson:"open_id"`
	Uid           uint32             `bson:"uid"`
	IsForbid      bool               `bson:"is_forbid"`
	ForbidEndTime uint32             `bson:"forbid_end_time"`
}

func (d *Dao) InsertAccount(account *Account) error {
	if d.mongo == nil {
		return d.InsertAccountGorm(account)
	}
	db := d.mongoDb.Collection("account")
	_, err := db.InsertOne(context.TODO(), account)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QueryAccountByOpenId(openId string) (*Account, error) {
	if d.mongo == nil {
		return d.QueryAccountByOpenIdGorm(openId)
	}
	db := d.mongoDb.Collection("account")
	result := db.FindOne(
		context.TODO(),
		bson.D{{"open_id", openId}},
	)
	account := new(Account)
	err := result.Decode(account)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return account, nil
}
