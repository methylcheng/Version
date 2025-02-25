package service

import (
	"StudentVersionManagement/dao"
	"StudentVersionManagement/model"
	"fmt"
	"gorm.io/gorm"
	"log"
)

// VersionMysqlService 定义mysql数据库服务层结构体
type VersionMysqlService struct {
	mysqlDao *dao.VersionMysqlDao
}

// NewVersionMysqlService 创建一个新的 VersionMysqlService 实例
func NewVersionMysqlService(mysqlDao *dao.VersionMysqlDao) (*VersionMysqlService, error) {
	// 检查 mysqlDao 是否为 nil，如果为 nil，则返回错误。
	// 这是一个重要的检查，因为如果 mysqlDao 为 nil，那么在后续的数据库操作中将会引发 panic。
	if mysqlDao == nil {
		return nil, fmt.Errorf("mysqlDao 为空")
	}

	// 返回一个新的 VersionMysqlService 实例。
	// 这里将 mysqlDao 传递给 VersionMysqlService 的构造函数，以便在新实例中使用。
	return &VersionMysqlService{
		mysqlDao: mysqlDao,
	}, nil
}

// ConvertToVersion 将从数据库获取的版本信息转换为服务使用的版本格式。
// 这个函数接收一个指向model.Version的指针作为参数，该参数包含了从数据库中获取的版本信息。
// 如果传入的versionDB参数为空，函数将返回一个错误，指出versionDB为空的问题。
// 如果一切正常，函数将返回一个指向model.Version的指针，该指针包含了转换后的版本信息，以及nil错误。
func (sms *VersionMysqlService) ConvertToVersion(versionDB *model.Version) (*model.Version, error) {
	// 检查传入的versionDB参数是否为空，如果为空，则返回错误。
	if versionDB == nil {
		return nil, fmt.Errorf("VersionMysqlService.ConvertToVersion 方法发生错误：versionDB为空")
	}

	// 创建一个新的Version实例，并将其字段设置为与versionDB相同的值。
	// 这个新实例将用于服务内部的版本处理。
	version := &model.Version{
		ID:        versionDB.ID,
		Platform:  versionDB.Platform,
		VersionNo: versionDB.VersionNo,
	}

	// 返回转换后的version实例和nil错误，表示转换成功。
	return version, nil
}

// VersionExists 判断版本号是否存在
func (sms *VersionMysqlService) VersionExists(id string) error {
	// 调用数据层代码 判断版本号是否存在
	if _, err := sms.mysqlDao.GetVersionByID(id); err != nil {
		return fmt.Errorf("VersionMysqlService.VersionExists 版本号：%s不存在：%w", id, err)
	}
	return nil
}

// AddVersionToMysql 向数据库添加版本信息
func (sms *VersionMysqlService) AddVersionToMysql(version *model.Version) error {
	// 调用数据层代码 添加版本信息
	if err := sms.mysqlDao.AddVersion(version); err != nil {
		return fmt.Errorf("VersionMysqlService.AddVersionToMysql 添加版本：%s失败：%w", version.ID, err)
	}
	log.Printf("添加版本：%s", version.ID)
	return nil
}

// GetVersionFromMysql 从 MySQL 中获取指定 ID 的版本信息
func (sms *VersionMysqlService) GetVersionFromMysql(id string) (*model.Version, error) {
	// 调用数据层代码 获取版本信息
	versionDB, err := sms.mysqlDao.GetVersionByID(id)
	if err != nil {
		return nil, fmt.Errorf("VersionMysqlService.GetVersionFromMysql 获取版本：%s失败：%w", id, err)
	}
	// 调用数据层代码 把 MySQL 数据库中的版本号转化为 model 中的版本号
	version, err := sms.ConvertToVersion(versionDB)
	if err != nil {
		return nil, fmt.Errorf("VersionMysqlService.GetVersionFromMysql 把 MySQL 数据库中的版本号转化为 model 中的版本号失败：%w", err)
	}
	return version, nil
}

// UpdateVersion 更新 MySQL 中的版本信息
// 该方法接收一个事务指针和一个版本信息指针，先检查版本是否存在，然后调用数据层代码更新版本信息
func (sms *VersionMysqlService) UpdateVersion(tx *gorm.DB, version *model.Version) error {
	// 先判断是否存在
	err := sms.VersionExists(version.ID)
	if err != nil {
		// 回滚事务
		tx.Rollback()
		// 记录错误日志
		log.Printf("VersionMysqlService.UpdateVersion 更新学生：%s失败：%v", version.ID, err)
		return fmt.Errorf("VersionMysqlService.UpdateVersion 更新学生：%s失败：%w", version.ID, err)
	}

	// 调用数据层代码 更新学生信息
	if err = sms.mysqlDao.UpdateVersion(tx, version); err != nil {
		// 回滚事务
		tx.Rollback()
		// 记录错误日志
		log.Printf("VersionMysqlService.UpdateVersion 在数据库更新学生：%s失败：%v", version.ID, err)
		return fmt.Errorf("VersionMysqlService.UpdateVersion 在数据库更新学生：%s失败：%w", version.ID, err)
	}
	log.Printf("在数据库更新版本：%s", version.ID)
	//向成绩表插入数据 先判断是否存在 如果存在就更新 不存在就添加
	return nil
}

// DeleteVersion 删除版本号
func (sms *VersionMysqlService) DeleteVersion(id string) error {
	// 先判断是否存在
	if err := sms.VersionExists(id); err != nil {
		return fmt.Errorf("VersionMysqlService.DeleteVersion 删除版本：%s失败：%w", id, err)
	}

	// 调用数据层代码 删除版本信息
	if err := sms.mysqlDao.DeleteVersion(id); err != nil {
		return fmt.Errorf("VersionMysqlService.DeleteVersion 删除版本：%s失败：%w", id, err)
	}
	log.Printf("删除版本：%s", id)
	return nil
}

// GetAllVersions 获取所有版本号
func (sms *VersionMysqlService) GetAllVersions() ([]*model.Version, error) {
	if sms.mysqlDao == nil {
		return nil, fmt.Errorf("mysqlDao 为空")
	}
	// 调用数据层代码 获取所有版本信息
	versionsDB, err := sms.mysqlDao.GetAllVersions()
	if err != nil {
		return nil, fmt.Errorf("VersionMysqlService.GetAllVersions 获取所有版本失败：%w", err)
	}
	// 调用数据层代码 把MySQL数据库中的版本号转化为model中的版本号
	versions := make([]*model.Version, len(versionsDB))
	for i, versionDB := range versionsDB {
		version, err := sms.ConvertToVersion(versionDB)
		if err != nil {
			return nil, fmt.Errorf("VersionMysqlService.GetAllVersions 把MySQL数据库中的版本号转化为model中的版本号失败：%w", err)
		}
		versions[i] = version
	}
	return versions, nil
}
