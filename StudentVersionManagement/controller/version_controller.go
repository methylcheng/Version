package controller

import (
	"StudentVersionManagement/model"
	"StudentVersionManagement/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// VersionController 版本号控制器
// 它包含一个版本服务的实例，用于执行与版本号相关的操作
type VersionController struct {
	// versionService 是 VersionController 的一个成员
	// 它是一个指向 service.VersionService 的指针，用于提供版本号相关的服务
	versionService *service.VersionService
}

// NewVersionController 创建版本号控制器实例
func NewVersionController(versionService *service.VersionService) *VersionController {
	return &VersionController{
		versionService: versionService,
	}
}

// AddVersion 添加版本号信息
// @Summary 添加版本号信息
// @Description 添加新的版本号信息
// @Tags 版本管理
// @Accept json
// @Produce json
// @Param version body model.Version true "版本号信息"
// @Success 200 "成功消息"
// @Failure 400 "请求错误"
// @Failure 500 "服务器错误"
// @Router /add_versions [post]

func (vc *VersionController) AddVersion(c *gin.Context) {
	// 初始化一个Version实例以存储要添加的版本信息
	var version model.Version

	// 尝试将请求体中的JSON数据绑定到version结构体中
	// 如果绑定失败，返回400状态码和错误信息
	if err := c.BindJSON(&version); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用versionService的AddVersion方法来添加版本信息
	// 如果添加失败，返回500状态码和错误信息
	err := vc.versionService.AddVersion(&version)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果成功，返回200状态码和成功消息
	c.JSON(http.StatusOK, gin.H{"message": "成功添加版本号"})
}

// DeleteVersion 删除版本号信息
// @Summary 删除版本号信息
// @Description 根据ID删除版本号信息
// @Tags 版本管理
// @Accept json
// @Produce json
// @Param id path string true "版本号ID"
// @Success 200 "成功消息"
// @Failure 500 "服务器错误"
// @Router /delete_versions/{id} [delete]
func (vc *VersionController) DeleteVersion(c *gin.Context) {
	// 获取URL参数中的ID
	id := c.Param("id")

	// 调用服务层的删除方法
	err := vc.versionService.DeleteVersion(id)
	if err != nil {
		// 如果删除失败，返回500错误和错误信息
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果删除成功，返回200和成功消息
	c.JSON(http.StatusOK, gin.H{"message": "版本号删除成功"})
}

// UpdateVersion 更新版本号信息
// @Summary 更新版本号信息
// @Description 根据ID更新版本号信息
// @Tags 版本管理
// @Accept json
// @Produce json
// @Param id path string true "版本号ID"
// @Param version body model.Version true "版本号信息"
// @Success 200 "成功消息"
// @Failure 400 "请求错误"
// @Failure 500 "服务器错误"
// @Router /update_versions/{id} [put]
func (vc *VersionController) UpdateVersion(c *gin.Context) {
	// 获取路径参数中的 ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version ID"})
		return
	}

	// 检查版本是否存在
	existingVersion, err := vc.versionService.GetVersionByID(id)
	if err != nil {
		// 记录错误日志
		log.Printf("根据ID寻找版本号时发生错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingVersion == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "版本号未找到"})
		return
	}

	var updatedVersion model.Version
	if err := c.BindJSON(&updatedVersion); err != nil {
		// 记录错误日志
		log.Printf("数据绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 确保版本 ID 一致
	updatedVersion.ID = id

	// 调用服务层的更新方法
	err = vc.versionService.UpdateVersion(&updatedVersion)
	if err != nil {
		// 记录错误日志
		log.Printf("版本号更新失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "版本号更新成功"})
}

// GetVersionByID 获取指定 ID 的版本号信息
// @Summary 根据ID获取版本号信息
// @Description 根据ID获取版本号信息
// @Tags 版本管理
// @Accept application/json
// @Produce application/json
// @Param id path string true "版本号ID"
// @Success 200 {object} model.Version
// @Failure 500  "服务器错误"
// @Router /get_versions/{id} [get]
func (vc *VersionController) GetVersionByID(c *gin.Context) {
	// 从路由参数中获取ID值
	id := c.Param("id")

	// 调用服务层方法，根据ID获取版本信息
	version, err := vc.versionService.GetVersionByID(id)
	if err != nil {
		// 如果发生错误，返回500错误，包含错误信息
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果未找到版本信息，返回一个空的版本对象
	if version == nil {
		version = &model.Version{}
	}

	// 返回200状态码，以及找到的版本信息
	c.JSON(http.StatusOK, version)
}

// JoinRaftCluster 向领导者节点发送请求 把自身加入到集群中
// @Summary 向领导者节点发送请求，把自身加入到集群中
// @Description 此接口用于将当前节点加入到Raft集群中。
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param ID query string true "节点ID"
// @Param Address query string true "节点地址"
// @Success 200 {object} map[string]string "成功消息"
// @Failure 400 {object} map[string]string "请求错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /join_raft_cluster/ [get]
func (vc *VersionController) JoinRaftCluster(c *gin.Context) {
	// 从请求路径参数中获取节点ID和地址
	ID := c.Param("ID")
	Address := c.Param("Address")

	// 调用服务层方法尝试加入Raft集群
	err := vc.versionService.JoinRaftCluster(ID, Address)
	if err != nil {
		// 如果发生错误，返回HTTP 500响应，包含错误信息
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果成功，返回HTTP 200响应，包含成功消息
	c.JSON(http.StatusOK, gin.H{"message": "节点成功加入集群"})
}
