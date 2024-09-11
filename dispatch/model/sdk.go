package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sdk struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	NextSdkAccountId uint32             `bson:"next_sdk_account_id"` // 下一个自增账号id
}

type SdkGorm struct {
	ID               uint32 `gorm:"column:id;type:bigint(20);primaryKey"`
	NextSdkAccountId uint32 `gorm:"column:next_sdk_account_id;type:bigint(20)"` // 下一个自增账号id
}

func (s SdkGorm) TableName() string {
	return "sdk"
}
