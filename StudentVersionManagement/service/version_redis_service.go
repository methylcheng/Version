package service

import (
	"StudentVersionManagement/dao"
	"StudentVersionManagement/model"
	"fmt"
	"log"
	"strings"
)

// VersionRedisService 版本号 Redis 服务
type VersionRedisService struct {
	redisDao *dao.VersionRedisDao
}

// NewVersionRedisService 创建版本号 Redis 服务实例
func NewVersionRedisService(redisDao *dao.VersionRedisDao) *VersionRedisService {
	return &VersionRedisService{
		redisDao: redisDao,
	}
}

// VersionExists 判断版本是否存在
func (vrs *VersionRedisService) VersionExists(id string) error {
	_, err := vrs.redisDao.GetVersionByID(id)
	versionNotFoundErrMsg := fmt.Sprintf("缓存中不存在版本：%s", id)
	if err != nil {
		if strings.Contains(err.Error(), versionNotFoundErrMsg) {
			log.Printf("缓存中不存在版本：%s", id)
			return err
		}
		return err
	}
	return nil
}

// AddVersion 添加版本号信息到 Redis
func (vrs *VersionRedisService) AddVersion(version *model.Version) error {
	log.Println("添加版本：", version.ID)
	return vrs.redisDao.AddVersion(version)
}

// DeleteVersion 从 Redis 删除版本号信息
func (vrs *VersionRedisService) DeleteVersion(id string) error {
	if err := vrs.VersionExists(fmt.Sprintf("%d", id)); err != nil {
		return fmt.Errorf("版本号不存在")
	}
	log.Println("删除版本：", id)
	return vrs.redisDao.DeleteVersion(id)
}

// UpdateVersion 更新 Redis 中的版本信息
// 该方法接收一个版本信息指针，先检查版本是否存在，然后调用数据层代码更新版本信息
func (vrs *VersionRedisService) UpdateVersion(version *model.Version) error {
	if err := vrs.VersionExists(version.ID); err != nil {
		// 记录错误日志
		log.Printf("版本号不存在：%v", err)
		return fmt.Errorf("版本号不存在")
	}
	log.Println("更新版本：", version.ID)
	return vrs.redisDao.UpdateVersion(version)
}

// GetVersionByID 从 Redis 中获取指定 ID 的版本信息
func (vrs *VersionRedisService) GetVersionByID(id string) (*model.Version, error) {
	if err := vrs.VersionExists(id); err != nil {
		return nil, fmt.Errorf("版本号不存在")
	}
	log.Println("获取版本：", id)
	return vrs.redisDao.GetVersionByID(id)
}
