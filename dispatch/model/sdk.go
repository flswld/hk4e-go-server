package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sdk struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	NextSdkAccountId uint32             `bson:"next_sdk_account_id"` // 下一个自增账号id
}
