package node

import (
	"fmt"
	"github.com/hashicorp/raft"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var snapshotDirMutex sync.Mutex

// NewRaftNode 创建并启动 Raft 节点
func NewRaftNode(ID string, address string, peers []string, fsm raft.FSM) (*raft.Raft, error) {
	// 打印创建Raft节点的日志信息。
	log.Printf("Creating Raft node with ID: %s, address: %s", ID, address)

	// 初始化Raft配置。
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(ID)
	config.SnapshotInterval = 200 * time.Second
	config.SnapshotThreshold = 1000

	// 初始化日志存储和状态存储。
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// 为每个节点创建独立的快照目录。
	snapshotDirMutex.Lock()
	snapshotDir := filepath.Join("snapshots", ID)
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		snapshotDirMutex.Unlock()
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}
	snapshotStore, err := raft.NewFileSnapshotStore(snapshotDir, 1, os.Stderr)
	snapshotDirMutex.Unlock()
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %w", err)
	}
	// 初始化传输层。
	transport, err := raft.NewTCPTransport("localhost:0", nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}
	if transport == nil {
		return nil, fmt.Errorf("failed to create transport")
	}
	// 创建Raft节点。
	raftNode, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft node: %w", err)
	}
	// 如果是第一个节点，初始化集群。
	if len(peers) == 0 {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		future := raftNode.BootstrapCluster(configuration)
		if err := future.Error(); err != nil {
			return nil, fmt.Errorf("failed to bootstrap cluster: %w", err)
		}
	} else {
		// 添加其他节点到集群。
		log.Printf("节点 %s 尝试加入现有集群，等待选举完成...", ID)
		time.Sleep(5 * time.Second)

		url := fmt.Sprintf("http://localhost:8080/JoinRaftCluster?ID=%s&Address=%s", ID, address)
		_, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to join cluster: %w", err)
		}
	}
	// 返回创建的Raft节点。
	return raftNode, nil
}
