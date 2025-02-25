package fsm

import (
	"StudentVersionManagement/interfaces"
	"StudentVersionManagement/model"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"io"
)

// Command 定义 Raft 日志条目中的命令
// 它包含了操作类型、版本信息、唯一标识符和检查大小等信息。
type Command struct {
	// Operation 是命令的操作类型，例如创建、删除等。
	Operation string `json:"operation"`
	// Version 是执行命令时相关的版本信息，可能是nil，表示可选字段。
	Version *model.Version `json:"version,omitempty"`
	// Id 是命令的唯一标识符。
	Id string `json:"id"`
	// ExamineSize 是在执行命令前进行检查的大小限制。
	ExamineSize int `json:"examine_size"`
}

// VersionFSM 处理 Raft 日志条目的应用
// 它依赖于一个实现了 VersionServiceInterface 接口的服务来执行版本相关的操作。
type VersionFSM struct {
	// service 是 VersionFSM 的核心组件，提供了版本管理所需的服务接口。
	// 通过这个服务接口，VersionFSM 可以执行版本查询、更新等操作。
	service interfaces.VersionServiceInterface
}

// NewVersionFSM 创建一个新的 VersionFSM 实例。
func NewVersionFSM(service interfaces.VersionServiceInterface) *VersionFSM {
	return &VersionFSM{
		service: service,
	}
}

// Apply 应用 Raft 日志条目
func (f *VersionFSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("fsm.Apply 接触json绑定失败: %s", err)
	}
	switch cmd.Operation {
	case "add":
		return f.service.AddVersionInternal(cmd.Version)
	case "update":
		return f.service.UpdateVersionInternal(cmd.Version)
	case "delete":
		return f.service.DeleteVersionInternal(cmd.Id)
	case "reloadCacheData":
		if err := f.service.ReLoadCacheDataInternal(); err != nil {
			return fmt.Errorf("fsm.Apply reload cache data fail: %s", err)
		}
		return nil
	default:
		return fmt.Errorf("fsm.Apply unknown operation: %s", cmd.Operation)
	}
}

// Snapshot 创建快照
func (f *VersionFSM) Snapshot() (raft.FSMSnapshot, error) {
	// 实现快照逻辑
	return nil, nil
}

// Restore 恢复快照
func (fsm *VersionFSM) Restore(snapshot io.ReadCloser) error {
	defer snapshot.Close()
	fmt.Printf("Restoring snapshot data\n")
	return nil
}
