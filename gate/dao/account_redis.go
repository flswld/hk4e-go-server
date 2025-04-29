package dao

import (
	"context"
	"time"

	"github.com/flswld/halo/logger"
)

// RedisAccountKeyPrefix key前缀
const RedisAccountKeyPrefix = "HK4E"

// GetRedisAccountLockKey 获取账号分布式锁key
func (d *Dao) GetRedisAccountLockKey(openId string) string {
	return RedisAccountKeyPrefix + ":ACCOUNT_LOCK:" + openId
}

// 基于redis的玩家离线数据分布式锁实现

const (
	MaxLockAliveTime = 10000 // 单个锁的最大存活时间 毫秒
)

// DistLock 加锁并返回是否成功
func (d *Dao) DistLock(openId string) bool {
	var result = false
	var err error = nil
	if d.redisCluster != nil {
		result, err = d.redisCluster.SetNX(context.TODO(),
			d.GetRedisAccountLockKey(openId),
			time.Now().UnixMilli(),
			time.Millisecond*time.Duration(MaxLockAliveTime)).Result()
	} else {
		result, err = d.redis.SetNX(context.TODO(),
			d.GetRedisAccountLockKey(openId),
			time.Now().UnixMilli(),
			time.Millisecond*time.Duration(MaxLockAliveTime)).Result()
	}
	if err != nil {
		logger.Error("redis lock setnx error: %v", err)
		return false
	}
	return result
}

// DistUnlock 解锁
func (d *Dao) DistUnlock(openId string) {
	var result int64 = 0
	var err error = nil
	if d.redisCluster != nil {
		result, err = d.redisCluster.Del(context.TODO(), d.GetRedisAccountLockKey(openId)).Result()
	} else {
		result, err = d.redis.Del(context.TODO(), d.GetRedisAccountLockKey(openId)).Result()
	}
	if err != nil {
		logger.Error("redis lock del error: %v", err)
		return
	}
	if result == 0 {
		logger.Error("redis lock del result is fail")
		return
	}
}
