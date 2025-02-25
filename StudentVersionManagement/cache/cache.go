package cache

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

// RedisClient 是一个全局变量，用于存储指向 Redis 客户端的引用。
// 通过它可以在程序的任何地方执行 Redis 相关的操作，例如数据存储和检索。
var RedisClient *redis.Client

// InitRedis 初始化Redis客户端。
// 该函数需要三个参数：addr (Redis服务器地址)，password (访问密码)，db (数据库编号)。
// 它会尝试根据这些参数连接到Redis服务器，并验证连接是否成功。
func InitRedis(addr, password string, db int) {
	// 创建Redis客户端实例。
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 发送PING命令以验证与Redis服务器的连接。
	pong, err := RedisClient.Ping(RedisClient.Context()).Result()
	if err != nil {
		// 如果连接失败，输出错误信息。
		fmt.Printf("连接redis数据库失败: %v\n", err)
		return
	}

	// 如果连接成功，输出成功信息。
	fmt.Printf("连接redis数据库成功: %s\n", pong)
}

// AcquireLock 获取分布式锁
// 参数:
//
//	key: 锁的唯一标识符
//	expiration: 锁的过期时间
//
// 返回值:
//
//	bool: 是否成功获取锁
//	error: 错误信息，如果有的话
func AcquireLock(key string, expiration time.Duration) (bool, error) {
	// 使用SetNX命令尝试设置键为"locked"，并设置过期时间
	ok, err := RedisClient.SetNX(RedisClient.Context(), key, "locked", expiration).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

// ReleaseLock 释放分布式锁
// 参数:
//
//	key: 锁的唯一标识符
//
// 返回值:
//
//	bool: 锁是否被成功释放
//	error: 错误信息，如果有的话
func ReleaseLock(key string) (bool, error) {
	// 使用Del命令删除键，以释放锁
	deleted, err := RedisClient.Del(RedisClient.Context(), key).Result()
	if err != nil {
		return false, err
	}
	// 如果deleted大于0，表示锁被成功释放
	return deleted > 0, nil
}
