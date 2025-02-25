package service

import (
	"StudentVersionManagement/cache"
	"StudentVersionManagement/config"
	"StudentVersionManagement/interfaces"
	"StudentVersionManagement/model"
	"StudentVersionManagement/raft"
	"StudentVersionManagement/raft/fsm"
	"encoding/json"
	"fmt"
	raftfpk "github.com/hashicorp/raft"
	"log"
	"strings"
	"time"
)

// VersionService 定义版本服务层结构体
type VersionService struct {
	MysqlService *VersionMysqlService
	RedisService *VersionRedisService
	RaftNodes    map[string]*raftfpk.Raft
}

// NewVersionService 创建一个新的 VersionService 实例
func NewVersionService(mysqlService *VersionMysqlService, redisService *VersionRedisService, node config.Node, cfg config.Config, nodeIndex int) (*VersionService, error) {
	vs := &VersionService{
		MysqlService: mysqlService,
		RedisService: redisService,
		RaftNodes:    make(map[string]*raftfpk.Raft),
	}

	initializer := &raft.RaftInitializerImpl{}

	// 构建 peers 列表
	peers := make([]string, 0)
	for Index, peer := range cfg.Raft.Nodes {
		if Index < nodeIndex {
			peers = append(peers, peer.Address)
		}
	}

	raftNode, err := initializer.InitRaft(node.ID, node.Address, peers, vs)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Raft: %w", err)
	}
	vs.RaftNodes[node.ID] = raftNode
	return vs, nil

}

// 确保实现 VersionServiceInterface 接口
var _ interfaces.VersionServiceInterface = (*VersionService)(nil)

// VersionNotFoundErr 判断错误是不是没有找到版本号之类的错误 如果是那就继续去下一个数据源找 不返回
func (vs *VersionService) VersionNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("不存在版本"))
}

// JoinRaftCluster 将节点加入 Raft 集群
func (vs *VersionService) JoinRaftCluster(ID string, Address string) error {
	for _, node := range vs.RaftNodes {
		if node.State() == raftfpk.Leader {
			// 向领导者节点发送请求
			future := node.AddVoter(raftfpk.ServerID(ID), raftfpk.ServerAddress(Address), 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("failed to add voter to Raft: %w", err)
			}
		}
	}
	return fmt.Errorf("没有领导者")
}

// applyRaftCommand 将命令提交给领导者
func (vs *VersionService) applyRaftCommand(operation string, version *model.Version, id string, examineSize int) error {
	// 创建 Raft 命令
	cmd := fsm.Command{
		Operation:   operation,
		Version:     version,
		Id:          id,
		ExamineSize: examineSize,
	}
	// 序列化命令
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal Raft command: %w", err)
	}
	//找到领导者节点
	for ID, raftNode := range vs.RaftNodes {
		if raftNode.State() == raftfpk.Leader {
			// 向领导者节点发送请求
			future := raftNode.Apply(data, 200)
			if err := future.Error(); err != nil {
				return fmt.Errorf("ApplyRaftCommandToLeader failed to apply command to leader Node %s: %w", ID, err)
			}
			// 处理响应
			if future.Error() != nil {
				return fmt.Errorf("failed to apply command to Raft: %w", future.Error())
			}
			return nil
		}
	}
	return fmt.Errorf("未找到领导者")
}

// RestoreCacheData 恢复缓存机制 mysql有事务可以很方便地回滚 此函数专门用于恢复缓存的数据
func (vs *VersionService) RestoreCacheData(id string) error {
	//如果要恢复数据 mysql的事务会回滚 所以这个时候找到的版本还是一开始的版本
	versionBack, err := vs.MysqlService.GetVersionFromMysql(id)
	if err != nil {
		return fmt.Errorf("failed to get version from mysql: %w", err)
	}
	if err := vs.RedisService.AddVersion(versionBack); err != nil {
		return fmt.Errorf("failed to add version to redis: %w", err)
	}
	return nil
}

// ReLoadCacheDataInternal 重新加载缓存数据（内部方法）
func (vs *VersionService) ReLoadCacheDataInternal() error {
	// 获取分布式锁
	lockKey := "reload_cache_lock"
	expiration := 10 * time.Second
	acquired, err := cache.AcquireLock(lockKey, expiration)
	if err != nil {
		log.Printf("获取分布式锁失败: %v", err)
		return err
	}
	if !acquired {
		log.Printf("无法获取锁，另一个进程正在持有它。")
		return fmt.Errorf("无法获取锁，另一个进程正在持有它。")
	}
	defer func() {
		// 释放分布式锁
		if _, err := cache.ReleaseLock(lockKey); err != nil {
			log.Printf("无法释放锁: %v", err)
		}
	}()

	// 获取所有版本
	versions, err := vs.MysqlService.GetAllVersions()
	if err != nil {
		log.Printf("无法从mysql中获取所有版本信息: %v", err)
		return err
	}
	// 添加版本到 Redis
	for _, version := range versions {
		if err := vs.RedisService.AddVersion(version); err != nil {
			log.Printf("无法向redis中添加版本信息: %v", err)
			return err
		}
	}
	return nil
}

// AddVersionInternal 添加版本信息
func (vs *VersionService) AddVersionInternal(version *model.Version) error {
	// 获取分布式锁
	lockKey := "add_version_lock"
	expiration := 10 * time.Minute
	acquired, err := cache.AcquireLock(lockKey, expiration)
	if err != nil {
		return fmt.Errorf("获取分布式锁失败: %w", err)
	}
	if !acquired {
		return fmt.Errorf("无法获取锁，另一个进程正在持有它。")
	}
	defer func() {
		// 释放分布式锁
		if _, err := cache.ReleaseLock(lockKey); err != nil {
			log.Printf("无法释放锁: %v", err)
		}
	}()

	// 开始 MySQL 事务
	tx := vs.MysqlService.mysqlDao.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("未能开始 MySQL 事务: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("发生错误: %v", r)
		}
	}()
	// 在 MySQL 数据库事务中添加版本信息
	if err := vs.MysqlService.AddVersionToMysql(version); err != nil {
		tx.Rollback()
		return fmt.Errorf("无法向mysql中添加版本信息: %w", err)
	}
	// MySQL 数据库事务提交成功后，尝试添加到redis
	if err := vs.RedisService.AddVersion(version); err != nil {
		tx.Rollback()
		return fmt.Errorf("无法向redis中添加版本信息: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交MySQL事务失败: %w", err)
	}
	return nil
}

// GetVersionByID 获取指定 ID 的版本信息
func (vs *VersionService) GetVersionByID(id string) (*model.Version, error) {
	// 尝试从缓存获取版本
	version, err := vs.RedisService.GetVersionByID(id)
	if err == nil {
		return version, nil
	}

	// 如果缓存中没有找到版本，则从数据库获取。如果确定缓存中没有该版本信息，则向缓存中添加该版本信息
	if vs.VersionNotFoundErr(err) {
		version, err = vs.MysqlService.GetVersionFromMysql(id)
		if err != nil {
			return nil, fmt.Errorf("从mysql获取版本失败: %w", err)
		}
		if err := vs.RedisService.AddVersion(version); err != nil {
			return nil, fmt.Errorf("向redis添加版本失败: %w", err)
		}
		return version, nil
	}
	return nil, err
}

// UpdateVersionInternal 内部更新版本信息的方法
// 该方法接收一个版本信息指针，使用 MySQL 事务进行更新操作，并更新 Redis 缓存
func (vs *VersionService) UpdateVersionInternal(version *model.Version) error {
	// 开始 MySQL 事务
	tx := vs.MysqlService.mysqlDao.DB.Begin()
	if tx.Error != nil {
		// 记录错误日志
		log.Printf("启动MySQL事务失败: %v", tx.Error)
		return fmt.Errorf("启动MySQL事务失败: %w", tx.Error)
	}
	// 确保在发生 panic 时回滚事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("发生错误: %v", r)
		}
	}()

	// 先检查 MySQL 中版本是否存在
	if err := vs.MysqlService.VersionExists(version.ID); err != nil {
		// 回滚事务
		tx.Rollback()
		// 记录错误日志
		log.Printf("VersionMysqlService.UpdateVersion 更新版本：%s失败：%v", version.ID, err)
		return fmt.Errorf("VersionMysqlService.UpdateVersion 更新版本：%s失败：%w", version.ID, err)
	}

	// 再检查 Redis 中版本是否存在
	if err := vs.RedisService.VersionExists(version.ID); err != nil {
		// 回滚事务
		tx.Rollback()
		// 记录错误日志
		log.Printf("Redis 中版本号：%s 不存在：%v", version.ID, err)
		return fmt.Errorf("Redis 中版本号：%s 不存在：%w", version.ID, err)
	}

	// 调用数据层代码更新 MySQL 中的版本信息
	if err := vs.MysqlService.UpdateVersion(tx, version); err != nil {
		// 回滚事务
		tx.Rollback()
		// 记录错误日志
		log.Printf("更新版本号失败：%v", err)
		return fmt.Errorf("更新版本号失败：%w", err)
	}

	// MySQL 数据库事务提交成功后，尝试更新缓存，还要确保数据一致性
	if err := vs.RedisService.UpdateVersion(version); err != nil {
		// 回滚事务
		tx.Rollback()
		// 记录错误日志
		log.Printf("更新版本到redis失败: %v", err)
		return fmt.Errorf("更新版本到redis失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		// 记录错误日志
		log.Printf("提交MySQL事务失败: %v", err)
		return fmt.Errorf("提交MySQL事务失败: %w", err)
	}

	return nil
}

// DeleteVersionInternal 删除版本信息
func (vs *VersionService) DeleteVersionInternal(id string) error {
	// 获取分布式锁
	lockKey := "delete_version_lock"
	expiration := 10 * time.Second
	acquired, err := cache.AcquireLock(lockKey, expiration)
	if err != nil {
		return fmt.Errorf("获取锁失败: %w", err)
	}
	if !acquired {
		return fmt.Errorf("获取锁失败，另一个进程持有该锁")
	}
	defer func() {
		// 释放分布式锁
		if _, err := cache.ReleaseLock(lockKey); err != nil {
			log.Printf("释放锁失败: %v", err)
		}
	}()

	// 开始 MySQL 事务
	tx := vs.MysqlService.mysqlDao.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("启动MySQL事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("发生错误: %v", r)
		}
	}()
	if err := vs.MysqlService.DeleteVersion(id); err != nil {
		if !vs.VersionNotFoundErr(err) {
			tx.Rollback()
			return fmt.Errorf("从mysql中删除版本失败: %w", err)
		}
	}
	if err := vs.RedisService.DeleteVersion(id); err != nil {
		if !vs.VersionNotFoundErr(err) {
			tx.Rollback()
			return fmt.Errorf("从redis删除版本失败: %w", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交MySQL事务失败: %w", err)
	}
	return nil
}

// ReLoadCacheData 重新加载缓存数据
func (vs *VersionService) ReLoadCacheData() error {
	// 获取分布式锁
	lockKey := "reload_cache_lock"
	expiration := 10 * time.Second
	acquired, err := cache.AcquireLock(lockKey, expiration)
	if err != nil {
		return fmt.Errorf("获取锁失败: %w", err)
	}
	if !acquired {
		return fmt.Errorf("获取锁失败，另一个进程持有该锁")
	}
	defer func() {
		// 释放分布式锁
		if _, err := cache.ReleaseLock(lockKey); err != nil {
			log.Printf("释放锁失败: %v", err)
		}
	}()

	// 获取所有版本
	versions, err := vs.MysqlService.GetAllVersions()
	if err != nil {
		return fmt.Errorf("无法从mysql获取所有版本: %w", err)
	}
	// 添加版本到 Redis
	for _, version := range versions {
		if err := vs.RedisService.AddVersion(version); err != nil {
			return fmt.Errorf("日志含义向redis添加版本失败: %w", err)
		}
	}
	return vs.applyRaftCommand("reload", nil, "", len(versions))
}

// AddVersion 接收添加版本命令 提交给Raft节点
func (vs *VersionService) AddVersion(version *model.Version) error {
	// 提交给Raft节点
	return vs.applyRaftCommand("add", version, "", 0)
}

// UpdateVersion 提交更新版本信息的请求给 Raft 节点
// 该方法接收一个版本信息指针，将更新操作提交给 Raft 节点
func (vs *VersionService) UpdateVersion(version *model.Version) error {
	// 提交给 Raft 节点
	return vs.applyRaftCommand("update", version, "", 0)
}

// DeleteVersion 接收删除版本命令 提交给Raft节点
func (vs *VersionService) DeleteVersion(id string) error {
	// 提交给Raft节点
	return vs.applyRaftCommand("delete", nil, id, 0)
}
