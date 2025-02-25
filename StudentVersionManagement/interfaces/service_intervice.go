package interfaces

import (
	"StudentVersionManagement/model"
)

// VersionServiceInterface 定义了版本管理服务的接口。
// 它提供了添加、更新、删除版本信息以及重新加载缓存数据的功能。
type VersionServiceInterface interface {
	// AddVersionInternal 用于内部添加新的版本信息。
	AddVersionInternal(student *model.Version) error

	// UpdateVersionInternal 用于内部更新版本信息。
	UpdateVersionInternal(student *model.Version) error

	// DeleteVersionInternal 用于内部删除版本信息。
	DeleteVersionInternal(id string) error

	// ReLoadCacheDataInternal 用于内部重新加载缓存数据。
	ReLoadCacheDataInternal() error
}
