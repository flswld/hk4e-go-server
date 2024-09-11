package dao

import (
	"errors"

	"gorm.io/gorm"
)

type AccountGorm struct {
	OpenId        string `gorm:"column:open_id;type:varchar(255);primaryKey"`
	Uid           uint32 `gorm:"column:uid;type:bigint(20)"`
	IsForbid      bool   `gorm:"column:is_forbid;type:tinyint(1)"`
	ForbidEndTime uint32 `gorm:"column:forbid_end_time;type:bigint(20)"`
}

func (a AccountGorm) TableName() string {
	return "account"
}

func (d *Dao) InsertAccountGorm(account *Account) error {
	err := d.gormDb.Create(&AccountGorm{
		OpenId:        account.OpenId,
		Uid:           account.Uid,
		IsForbid:      account.IsForbid,
		ForbidEndTime: account.ForbidEndTime,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QueryAccountByOpenIdGorm(openId string) (*Account, error) {
	accountGorm := new(AccountGorm)
	err := d.gormDb.Where("open_id = ?", openId).First(accountGorm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &Account{
		OpenId:        accountGorm.OpenId,
		Uid:           accountGorm.Uid,
		IsForbid:      accountGorm.IsForbid,
		ForbidEndTime: accountGorm.ForbidEndTime,
	}, nil
}
