package dao

import (
	"errors"

	"github.com/vmihailenco/msgpack/v5"
	"gorm.io/gorm"
)

type RegionGorm struct {
	ID   uint32 `gorm:"column:id;type:bigint(20);primaryKey"`
	Data []byte `gorm:"column:data;type:longblob"`
}

func (r RegionGorm) TableName() string {
	return "region"
}

func (d *Dao) InsertRegionGorm(region *Region) error {
	data, err := msgpack.Marshal(region)
	if err != nil {
		return err
	}
	err = d.gormDb.Create(&RegionGorm{
		ID:   1,
		Data: data,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdateRegionGorm(region *Region) error {
	data, err := msgpack.Marshal(region)
	if err != nil {
		return err
	}
	err = d.gormDb.Updates(&RegionGorm{
		ID:   1,
		Data: data,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) QueryRegionGorm() (*Region, error) {
	regionGorm := new(RegionGorm)
	err := d.gormDb.First(regionGorm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	region := new(Region)
	err = msgpack.Unmarshal(regionGorm.Data, region)
	if err != nil {
		return nil, err
	}
	return region, nil
}
