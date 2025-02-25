package routers

import (
	"StudentVersionManagement/controller"
	"github.com/gin-gonic/gin"
)

func SetUpRaftRouter(versionController *controller.VersionController) *gin.Engine {
	r := gin.Default()

	r.GET("", versionController.JoinRaftCluster)

	return r
}
