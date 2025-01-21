package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hk4e/common/config"
	"hk4e/pkg/logger"

	"github.com/glebarez/sqlite"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Dao struct {
	mongo        *mongo.Client
	mongoDb      *mongo.Database
	gormDb       *gorm.DB
	redis        *redis.Client
	redisCluster *redis.ClusterClient
}

func NewDao() (*Dao, error) {
	r := new(Dao)

	if strings.Contains(config.GetConfig().Database.Url, "mongodb://") {
		clientOptions := options.Client().ApplyURI(config.GetConfig().Database.Url)
		clientOptions = clientOptions.SetMinPoolSize(10)
		clientOptions = clientOptions.SetMaxPoolSize(100)
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			logger.Error("mongo connect error: %v", err)
			return nil, err
		}
		err = client.Ping(context.TODO(), readpref.Primary())
		if err != nil {
			logger.Error("mongo ping error: %v", err)
			return nil, err
		}
		r.mongo = client
		r.mongoDb = client.Database("gs_hk4e")
	} else {
		if strings.Contains(config.GetConfig().Database.Url, "mysql://") {
			dsn := strings.ReplaceAll(config.GetConfig().Database.Url, "mysql://", "")
			db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
				Logger: gormlogger.Default.LogMode(gormlogger.Info),
			})
			if err != nil {
				logger.Error("gorm open error: %v", err)
				return nil, err
			}
			r.gormDb = db
			sqlDb, err := db.DB()
			if err != nil {
				logger.Error("sql db open error: %v", err)
				return nil, err
			}
			sqlDb.SetMaxIdleConns(10)
			sqlDb.SetMaxOpenConns(100)
			sqlDb.SetConnMaxLifetime(time.Hour)
		} else if strings.Contains(config.GetConfig().Database.Url, "sqlite://") {
			dsn := strings.ReplaceAll(config.GetConfig().Database.Url, "sqlite://", "")
			db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
				Logger: gormlogger.Default.LogMode(gormlogger.Info),
			})
			if err != nil {
				logger.Error("gorm open error: %v", err)
				return nil, err
			}
			r.gormDb = db
		} else {
			err := errors.New(fmt.Sprintf("not support db type, url: %v", config.GetConfig().Database.Url))
			logger.Error("%v", err)
			return nil, err
		}
		tableList := []any{new(PlayerGorm), new(ChatMsgGorm), new(SceneBlockGorm)}
		for _, table := range tableList {
			err := r.gormDb.AutoMigrate(table)
			if err != nil {
				logger.Error("auto migrate error: %v", err)
				return nil, err
			}
		}
	}

	if !config.GetConfig().Hk4e.StandaloneModeEnable {
		r.redis = nil
		r.redisCluster = nil
		redisAddr := strings.ReplaceAll(config.GetConfig().Redis.Addr, "redis://", "")
		if strings.Contains(redisAddr, ",") {
			redisAddrList := strings.Split(redisAddr, ",")
			r.redisCluster = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:        redisAddrList,
				Password:     config.GetConfig().Redis.Password,
				PoolSize:     10,
				MinIdleConns: 1,
			})
		} else {
			r.redis = redis.NewClient(&redis.Options{
				Addr:         redisAddr,
				Password:     config.GetConfig().Redis.Password,
				DB:           0,
				PoolSize:     10,
				MinIdleConns: 1,
			})
		}
		var err error = nil
		if r.redisCluster != nil {
			err = r.redisCluster.Ping(context.TODO()).Err()
		} else {
			err = r.redis.Ping(context.TODO()).Err()
		}
		if err != nil {
			logger.Error("redis ping error: %v", err)
			return nil, err
		}
	}

	return r, nil
}

func (d *Dao) CloseDao() {
	if d.mongo != nil {
		err := d.mongo.Disconnect(context.TODO())
		if err != nil {
			logger.Error("mongo close error: %v", err)
		}
	}

	if !config.GetConfig().Hk4e.StandaloneModeEnable {
		var err error = nil
		if d.redisCluster != nil {
			err = d.redisCluster.Close()
		} else {
			err = d.redis.Close()
		}
		if err != nil {
			logger.Error("redis close error: %v", err)
		}
	}
}
