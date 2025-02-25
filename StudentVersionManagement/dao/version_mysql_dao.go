package dao

import (
	"StudentVersionManagement/model"
	"fmt"
	"gorm.io/gorm"
	"log"
)

// VersionMysqlDao 是版本号的数据访问对象（DAO），用于执行与版本信息相关的 MySQL 数据库操作。
// 该结构体封装了数据库连接和相关的方法，以便进行版本信息的增删改查等操作。
type VersionMysqlDao struct {
	// DB 是 GORM 数据库连接实例，用于执行所有数据库操作。
	// 通过该字段可以与 MySQL 数据库进行交互，执行查询、插入、更新和删除等操作。
	DB *gorm.DB
}

// NewVersionMysqlDao 创建并初始化一个新的 VersionMysqlDao 实例。
func NewVersionMysqlDao(db *gorm.DB) (*VersionMysqlDao, error) {
	// 检查数据库连接对象是否为 nil，如果为 nil，则返回错误。
	if db == nil {
		return nil, fmt.Errorf("数据库不存在，无法连接")
	}
	// 返回一个新的 VersionMysqlDao 实例，包含有效的数据库连接对象。
	return &VersionMysqlDao{
		DB: db,
	}, nil
}

// AddVersion 添加版本号信息
func (d *VersionMysqlDao) AddVersion(version *model.Version) error {
	// 使用DB连接创建新的版本记录
	err := d.DB.Create(version).Error
	// 如果创建过程中出现错误，封装错误信息并返回
	if err != nil {
		return fmt.Errorf("VersionMysqlDao.AddVersion 方法发生错误:%w", err)
	}
	// 如果创建成功，返回nil表示操作成功
	return nil
}

// DeleteVersion 通过ID删除版本信息。

func (d *VersionMysqlDao) DeleteVersion(id string) error {
	// 检查传入的ID是否为空，如果为空则返回错误。
	if id == "" {
		return fmt.Errorf("传入了空的版本ID")
	}

	// 执行数据库删除操作。
	// 如果发生错误，则返回包装过的错误信息。
	err := d.DB.Delete(&model.Version{ID: id}).Error
	if err != nil {
		return fmt.Errorf("VersionMysqlDao.DeleteVersion 方法发生错误:%w", err)
	}

	// 操作成功，没有错误返回。
	return nil
}

// UpdateVersion 更新 MySQL 中的版本信息
// 该方法接收一个事务指针和一个版本信息指针，使用 SQL 语句更新版本信息
func (d *VersionMysqlDao) UpdateVersion(tx *gorm.DB, version *model.Version) error {
	// 定义SQL语句，更新versions表中的version_no和platform字段
	// 只有当新值不为空时才进行更新，否则保持原值不变
	sqlStmt := `
        UPDATE versions
        SET
            version_no = IF(COALESCE(?, '') != '', ?, version_no),
            platform = IF(COALESCE(?, '') != '', ?, platform)
        WHERE id = ?
    `

	// 执行SQL语句，使用Exec方法可以避免SQL注入
	// 如果执行过程中发生错误，则记录错误日志并返回错误信息
	err := tx.Exec(sqlStmt,
		version.VersionNo, version.VersionNo,
		version.Platform, version.Platform,
		version.ID).Error
	if err != nil {
		// 记录错误日志
		log.Printf("StudentMysqlDao.UpdateStudent 方法发生错误: %v", err)
		// 返回错误信息，使用fmt.Errorf包装原始错误，以便于后续可能的错误处理
		return fmt.Errorf("StudentMysqlDao.UpdateStudent 方法发生错误:%w", err)
	}
	// 如果执行成功，返回nil表示操作成功
	return nil
}

// GetVersionByID 从 MySQL 中的 versions 表中获取指定 ID 的版本信息
func (d *VersionMysqlDao) GetVersionByID(id string) (*model.Version, error) {
	// 初始化版本信息结构体
	var version model.Version
	// 使用提供的ID查询数据库中的版本信息
	// 如果查询过程中遇到错误，返回错误信息
	err := d.DB.Where("id = ?", id).First(&version).Error
	if err != nil {
		return nil, fmt.Errorf("VersionMysqlDao.GetVersionByID 方法发生错误:%w", err)
	}
	// 查询成功，返回版本信息
	return &version, nil
}

// GetAllVersions 获取所有版本信息
// 该方法从数据库中查询并返回所有版本记录
// 如果数据库连接为空或查询过程中出现错误，将返回相应的错误
func (d *VersionMysqlDao) GetAllVersions() ([]*model.Version, error) {
	// 检查数据库连接是否为空
	if d.DB == nil {
		return nil, fmt.Errorf("数据库连接为空")
	}
	var versions []*model.Version
	// 执行数据库查询操作
	err := d.DB.Find(&versions).Error
	if err != nil {
		// 返回查询过程中出现的错误
		return nil, fmt.Errorf("VersionMysqlDao.GetAllVersions 方法发生错误:%w", err)
	}
	// 返回查询到的版本信息
	return versions, nil
}
