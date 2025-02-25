package config

import (
	"time"
)

// Config 定义配置结构体
type Config struct {
	MySQL struct {
		DSN string
	}
	Redis struct {
		Addr     string
		Password string
		DB       int
	}
	MemoryDB struct {
		Capacity   int
		EvictRatio float64
	}
	Server struct {
		ReloadInterval         time.Duration
		PeriodicDeleteInterval time.Duration
		ExamineSize            int
	}
	Raft struct {
		Nodes []Node
	}
}

type Node struct {
	ID      string
	Address string
	Port    string
}

// GetConfig 获取配置实例
func GetConfig() Config {
	return Config{
		//配置Mysql
		MySQL: struct {
			DSN string
		}{
			DSN: "root:wsm665881@tcp(127.0.0.1:3306)/student_version_db?charset=utf8mb4&parseTime=True&loc=Local",
		},
		//配置redis
		Redis: struct {
			Addr     string
			Password string
			DB       int
		}{
			Addr:     "127.0.0.1:6379",
			Password: "wsm665881",
			DB:       0,
		},
		//配置内存数据库
		Server: struct {
			ReloadInterval         time.Duration //重载缓存数据的时间间隔
			PeriodicDeleteInterval time.Duration //定期删除过期键的时间间隔
			ExamineSize            int           //定期删除过期键的检测数量
		}{
			ReloadInterval:         60 * time.Second,
			PeriodicDeleteInterval: time.Hour,
			ExamineSize:            10,
		},
		Raft: struct {
			Nodes []Node
		}{
			Nodes: []Node{
				{
					ID:      "node1",
					Address: "127.0.0.1",
					Port:    "8080",
				},
				{
					ID:      "node2",
					Address: "127.0.0.1",
					Port:    "8081",
				},
				{
					ID:      "node3",
					Address: "127.0.0.1",
					Port:    "8082",
				},
			},
		},
	}
}
