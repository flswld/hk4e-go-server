package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SdkAccount struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty"`
	AccountId            uint32             `bson:"account_id"`              // 账号id
	Username             string             `bson:"username"`                // 用户名
	Password             string             `bson:"password"`                // 密码
	Token                string             `bson:"token"`                   // 账号token
	TokenCreateTime      uint64             `bson:"token_create_time"`       // 毫秒时间戳
	ComboToken           string             `bson:"combo_token"`             // 游戏服务器token
	ComboTokenCreateTime uint64             `bson:"combo_token_create_time"` // 毫秒时间戳
}

type SdkAccountGorm struct {
	AccountId            uint32 `gorm:"column:account_id;type:bigint(20);primaryKey"`   // 账号id
	Username             string `gorm:"column:username;type:varchar(255)"`              // 用户名
	Password             string `gorm:"column:password;type:varchar(255)"`              // 密码
	Token                string `gorm:"column:token;type:varchar(255)"`                 // 账号token
	TokenCreateTime      uint64 `gorm:"column:token_create_time;type:bigint(20)"`       // 毫秒时间戳
	ComboToken           string `gorm:"column:combo_token;type:varchar(255)"`           // 游戏服务器token
	ComboTokenCreateTime uint64 `gorm:"column:combo_token_create_time;type:bigint(20)"` // 毫秒时间戳
}

func (s SdkAccountGorm) TableName() string {
	return "sdk_account"
}
