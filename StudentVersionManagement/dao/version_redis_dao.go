package dao

import (
	"StudentVersionManagement/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

// VersionRedisDao 是一个用于操作版本信息的Redis数据访问对象（DAO）。
// 它提供了一个接口，用于与Redis数据库进行交互，执行如读取、写入版本信息等操作。
// 该DAO模式的使用，隔离了业务逻辑与数据存储之间的耦合，提高了代码的可维护性和可重用性。
type VersionRedisDao struct {
	// Client 是指向redis.Client的指针，用于执行Redis命令。
	// 通过这个客户端，可以连接到Redis服务器并进行数据操作。
	Client *redis.Client
}

// NewVersionRedisDao 初始化缓存层结构体实例
// NewVersionRedisDao 创建并返回一个新的 VersionRedisDao 实例。
// 该函数接收一个 *redis.Client 参数，用于 VersionRedisDao 实例与 Redis 进行交互。
func NewVersionRedisDao(client *redis.Client) *VersionRedisDao {
	return &VersionRedisDao{
		Client: client,
	}
}

// AddVersion 添加版本号信息，判断当前版本id是否存在，若存在则返回错误，不存在则添加版本号信息
func (d *VersionRedisDao) AddVersion(version *model.Version) error {
	ctx := context.Background()
	key := fmt.Sprintf("version:%s", version.ID)

	// 将版本信息存储为 JSON 字符串
	jsonData, err := json.Marshal(version)
	if err != nil {
		return fmt.Errorf("VersionRedisDao.AddVersion 序列化版本信息失败: %w", err)
	}

	// 设置键值对并设置过期时间，例如 1 小时
	expiration := time.Hour
	err = d.Client.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		return fmt.Errorf("VersionRedisDao.AddVersion 方法发生错误: %w", err)
	}

	return nil
}

// DeleteVersion 删除指定ID的版本信息。
// 该方法通过构造Redis键名并调用Del方法来删除缓存中的版本信息。
func (d *VersionRedisDao) DeleteVersion(id string) error {
	// 创建一个新的上下文对象，用于取消请求。
	ctx := context.Background()

	// 构造Redis键名，并尝试删除缓存中的版本信息。
	err := d.Client.Del(ctx, fmt.Sprintf("version:%s", id)).Err()

	// 如果删除操作失败，则返回自定义错误信息，包装原始错误。
	if err != nil {
		return fmt.Errorf("VersionRedisDao.DeleteVersion 方法发生错误: %w", err)
	}

	// 删除成功，返回nil表示操作成功。
	return nil
}

// UpdateVersion 更新版本信息到Redis中
// 该方法接收一个版本模型，将其序列化为JSON格式后存储到Redis中对应的版本ID键下
func (d *VersionRedisDao) UpdateVersion(version *model.Version) error {
	// 初始化上下文对象，用于取消请求或传递请求级值
	ctx := context.Background()

	// 将版本信息序列化为JSON格式
	data, err := json.Marshal(version)
	if err != nil {
		// 记录错误日志
		log.Printf("VersionRedisDao.UpdateVersion 方法发生错误: %v", err)
		// 返回错误，使用fmt.Errorf以包装原始错误，提供更丰富的错误上下文信息
		return fmt.Errorf("VersionRedisDao.UpdateVersion 方法发生错误: %w", err)
	}

	// 将序列化的版本信息存储到Redis中，键名格式为"version:版本ID"
	err = d.Client.Set(ctx, fmt.Sprintf("version:%s", version.ID), data, 0).Err()
	if err != nil {
		// 记录错误日志
		log.Printf("VersionRedisDao.UpdateVersion 方法发生错误: %v", err)
		// 返回错误，使用fmt.Errorf以包装原始错误，提供更丰富的错误上下文信息
		return fmt.Errorf("VersionRedisDao.UpdateVersion 方法发生错误:%w", err)
	}

	// 操作成功，返回nil表示没有发生错误
	return nil
}

// GetVersionByID 根据给定的ID从Redis中获取版本信息。

func (d *VersionRedisDao) GetVersionByID(id string) (*model.Version, error) {
	// 创建一个新的上下文对象，用于取消请求。
	ctx := context.Background()

	// 从Redis中获取指定ID的版本信息。
	data, err := d.Client.Get(ctx, fmt.Sprintf("version:%s", id)).Result()
	if err != nil {
		// 如果错误是由于键不存在引起的，则返回nil, nil，表示未找到版本信息。
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		// 如果是其他错误，封装错误信息并返回。
		return nil, fmt.Errorf("VersionRedisDao.GetVersionByID 方法发生错误: %w", err)
	}

	// 初始化一个版本信息对象。
	version := &model.Version{}

	// 将获取到的JSON数据解码为版本信息对象。
	err = json.Unmarshal([]byte(data), version)
	if err != nil {
		// 如果解码过程中发生错误，封装错误信息并返回。
		return nil, fmt.Errorf("VersionRedisDao.GetVersionByID 方法发生错误: %w", err)
	}

	// 返回解码后的版本信息对象。
	return version, nil
}
