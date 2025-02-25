package model

// Version 结构体定义了版本信息的数据库模型。
// 它包括唯一的ID、版本号和平台信息。
type Version struct {
	// ID 是版本的唯一标识符，用作主键。
	ID string `gorm:"primaryKey"`

	// VersionNo 表示版本号，不能为空。
	VersionNo string `gorm:"not null"`

	// Platform 表示版本适用的平台，不能为空。
	Platform string `gorm:"not null"`
}

// VersionDB 关联 MySQL 的版本号表，用于存储不同平台的版本信息。

type VersionDB struct {
	// ID 是版本记录的唯一标识符，是主键。该字段在 JSON 中表示为 "id"，并且是必填项。
	ID string `json:"id" validate:"required" gorm:"primaryKey"`

	// VersionNo 是版本号，用于标识不同的版本。该字段在 JSON 中表示为 "versionNo"，并且是必填项。
	VersionNo string `json:"versionNo" validate:"required"`

	// Platform 是平台名称，用于区分不同平台的版本。该字段在 JSON 中表示为 "platform"，并且是必填项。
	Platform string `json:"platform" validate:"required"`
}
