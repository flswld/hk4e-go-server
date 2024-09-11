package dao

import (
	"errors"
	"hk4e/dispatch/model"

	"gorm.io/gorm"
)

func (d *Dao) InsertSdkGorm(sdk *model.Sdk) error {
	err := d.gormDb.Create(&model.SdkGorm{
		ID:               1,
		NextSdkAccountId: sdk.NextSdkAccountId,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateSdkGorm(sdk *model.Sdk) error {
	err := d.gormDb.Updates(&model.SdkGorm{
		ID:               1,
		NextSdkAccountId: sdk.NextSdkAccountId,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QuerySdkGorm() (*model.Sdk, error) {
	sdkGorm := new(model.SdkGorm)
	err := d.gormDb.First(sdkGorm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model.Sdk{
		NextSdkAccountId: sdkGorm.NextSdkAccountId,
	}, nil
}
