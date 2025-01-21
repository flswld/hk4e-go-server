package dao

import (
	"errors"

	"hk4e/dispatch/model"

	"gorm.io/gorm"
)

func (d *Dao) InsertSdkAccountGorm(account *model.SdkAccount) error {
	err := d.gormDb.Create(&model.SdkAccountGorm{
		AccountId:            account.AccountId,
		Username:             account.Username,
		Password:             account.Password,
		Token:                account.Token,
		TokenCreateTime:      account.TokenCreateTime,
		ComboToken:           account.ComboToken,
		ComboTokenCreateTime: account.ComboTokenCreateTime,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateSdkAccountFieldByFieldNameGorm(fieldName string, fieldValue any, fieldUpdateName string, fieldUpdateValue any) error {
	err := d.gormDb.Model(&model.SdkAccountGorm{}).Where(fieldName+" = ?", fieldValue).Update(fieldUpdateName, fieldUpdateValue).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QuerySdkAccountByFieldGorm(fieldName string, fieldValue any) (*model.SdkAccount, error) {
	sdkAccountGorm := new(model.SdkAccountGorm)
	err := d.gormDb.Where(fieldName+" = ?", fieldValue).First(sdkAccountGorm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model.SdkAccount{
		AccountId:            sdkAccountGorm.AccountId,
		Username:             sdkAccountGorm.Username,
		Password:             sdkAccountGorm.Password,
		Token:                sdkAccountGorm.Token,
		TokenCreateTime:      sdkAccountGorm.TokenCreateTime,
		ComboToken:           sdkAccountGorm.ComboToken,
		ComboTokenCreateTime: sdkAccountGorm.ComboTokenCreateTime,
	}, nil
}
