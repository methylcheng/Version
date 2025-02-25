package main

import (
	"StudentVersionManagement/cache"
	"StudentVersionManagement/config"
	"StudentVersionManagement/controller"
	"StudentVersionManagement/dao"
	"StudentVersionManagement/mysql"
	"StudentVersionManagement/routers"
	"StudentVersionManagement/service"
	"log"
	"sync"
	"time"
)

func startNode(node config.Node, cfg config.Config, nodeIndex int) {
	// 初始化数据库和缓存
	if err := mysql.InitDB(cfg.MySQL.DSN); err != nil {
		log.Fatalf("节点 %s 初始化数据库失败: %v", node.ID, err)
	}
	cache.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	// 初始化 DAO
	versionMysqlDao, err := dao.NewVersionMysqlDao(mysql.DB)
	if err != nil {
		log.Fatalf("节点 %s 初始化MySQL DAO失败: %v", node.ID, err)
	}
	versionRedisDao := dao.NewVersionRedisDao(cache.RedisClient)

	// 初始化服务
	versionMysqlService, err := service.NewVersionMysqlService(versionMysqlDao)
	if err != nil {
		log.Fatalf("节点 %s 初始化服务失败: %v", node.ID, err)
	}
	versionRedisService := service.NewVersionRedisService(versionRedisDao)
	versionService, err := service.NewVersionService(versionMysqlService, versionRedisService, node, cfg, nodeIndex)
	if err != nil {
		log.Fatalf("节点 %s 初始化服务失败: %v", node.ID, err)
	}

	// 初始化控制器
	versionController := controller.NewVersionController(versionService)

	//初始化路由
	versionRouter := routers.SetUpVersionRouter(versionController)
	ServerAddress := node.Address + ":" + node.Port
	if err = versionRouter.Run(ServerAddress); err != nil {
		log.Fatalf("节点 %s 初始化服务失败: %v", node.ID, err)
	}
}
func main() {
	cfg := config.GetConfig()
	// 初始化 Redis
	cache.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	if err := mysql.InitDB(cfg.MySQL.DSN); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	versionMysqlDao, err := dao.NewVersionMysqlDao(mysql.DB)
	if err != nil {
		log.Fatalf("初始化MySQL DAO失败: %v", err)
	}
	versionMysqlService, err := service.NewVersionMysqlService(versionMysqlDao)
	if err != nil {
		log.Fatalf("初始化VersionMysqlService失败: %v", err)
	}

	versionRedisDao := dao.NewVersionRedisDao(cache.RedisClient)
	versionRedisService := service.NewVersionRedisService(versionRedisDao)

	versionService, err := service.NewVersionService(versionMysqlService, versionRedisService, cfg.Raft.Nodes[0], cfg, 0)
	if err != nil {
		log.Fatalf("初始化VersionService失败: %v", err)
	}

	// 启动定期同步任务
	go startPeriodicSync(versionService, cfg.Server.ReloadInterval)
	var wg sync.WaitGroup
	for nodeIndex, node := range cfg.Raft.Nodes {
		wg.Add(1)
		go func(node config.Node, nodeIndex int) {
			defer wg.Done()
			startNode(node, cfg, nodeIndex)
		}(node, nodeIndex)
	}
	wg.Wait()
}

// startPeriodicSync 启动定期同步任务
func startPeriodicSync(versionService *service.VersionService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("开始定期同步 MySQL 数据到 Redis")
		if err := versionService.ReLoadCacheDataInternal(); err != nil {
			log.Printf("定期同步 MySQL 数据到 Redis 失败: %v", err)
		} else {
			log.Println("定期同步 MySQL 数据到 Redis 完成")
		}
	}
}
