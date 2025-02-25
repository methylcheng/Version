package routers

import (
	"StudentVersionManagement/controller"
	"github.com/gin-gonic/gin"
)

func SetUpVersionRouter(versionController *controller.VersionController) *gin.Engine {
	r := gin.Default()

	// 定义路由
	r.POST("/add_versions", versionController.AddVersion)
	r.DELETE("/delete_versions/:id", versionController.DeleteVersion)
	r.PUT("/update_versions/:id", versionController.UpdateVersion)
	r.GET("/get_versions/:id", versionController.GetVersionByID)

	r.GET("/join_raft_cluster", versionController.JoinRaftCluster)
	// 启动服务器
	return r
}
