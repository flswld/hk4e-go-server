package dao

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Region struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"`
	Ec2bData            []byte             `bson:"ec2b_data"`
	NextUid             uint32             `bson:"next_uid"`
	StopServer          bool               `bson:"stop_server"`
	StopServerStartTime uint32             `bson:"stop_server_start_time"`
	StopServerEndTime   uint32             `bson:"stop_server_end_time"`
	IpAddrWhiteList     []string           `bson:"ip_addr_white_list"`
}

func (d *Dao) InsertRegion(region *Region) error {
	if d.mongo == nil {
		return d.InsertRegionGorm(region)
	}
	db := d.mongoDb.Collection("region")
	_, err := db.InsertOne(context.TODO(), region)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateRegion(region *Region) error {
	if d.mongo == nil {
		return d.UpdateRegionGorm(region)
	}
	db := d.mongoDb.Collection("region")
	_, err := db.UpdateMany(
		context.TODO(),
		bson.D{},
		bson.D{{"$set", region}},
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QueryRegion() (*Region, error) {
	if d.mongo == nil {
		return d.QueryRegionGorm()
	}
	db := d.mongoDb.Collection("region")
	result := db.FindOne(
		context.TODO(),
		bson.D{},
	)
	region := new(Region)
	err := result.Decode(region)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return region, nil
}
