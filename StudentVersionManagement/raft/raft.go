package raft

import (
	"StudentVersionManagement/interfaces"
	"StudentVersionManagement/raft/fsm"
	"StudentVersionManagement/raft/node"
	"github.com/hashicorp/raft"
)

// RaftInitializerImpl 实现 RaftInitializer 接口
type RaftInitializerImpl struct{}

// InitRaft 方法用于初始化一个 Raft 节点。
// 它创建一个有限状态机（FSM）实例并与 Raft 节点关联，该节点能够与其它节点一起工作，形成一个 Raft 集群。
func (r *RaftInitializerImpl) InitRaft(ID string, address string, peers []string, service interfaces.VersionServiceInterface) (*raft.Raft, error) {
	// 创建一个有限状态机（FSM）实例，它将处理Raft节点的状态变化。
	fsmInstance := fsm.NewVersionFSM(service)

	// 使用提供的ID、地址、同伴节点列表和FSM实例来创建一个新的Raft节点。
	raftNode, err := node.NewRaftNode(ID, address, peers, fsmInstance)
	if err != nil {
		// 如果创建Raft节点时发生错误，返回nil和错误信息。
		return nil, err
	}

	// 返回成功创建的Raft节点指针和nil错误。
	return raftNode, nil
}
