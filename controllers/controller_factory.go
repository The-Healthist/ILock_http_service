package controllers

import (
	"ilock-http-service/services/container"

	"github.com/gin-gonic/gin"
)

// BaseController 是所有控制器的基础接口
type BaseController interface {
	// 获取服务容器
	GetContainer() *container.ServiceContainer
	// 获取Gin上下文
	GetContext() *gin.Context
}

// BaseControllerImpl 是控制器的基础实现
type BaseControllerImpl struct {
	Container *container.ServiceContainer
	Context   *gin.Context
}

// GetContainer 实现 BaseController 接口
func (c *BaseControllerImpl) GetContainer() *container.ServiceContainer {
	return c.Container
}

// GetContext 实现 BaseController 接口
func (c *BaseControllerImpl) GetContext() *gin.Context {
	return c.Context
}

// ControllerFactory 用于创建控制器的工厂
type ControllerFactory struct {
	Container *container.ServiceContainer
}

// NewControllerFactory 创建一个新的控制器工厂
func NewControllerFactory(container *container.ServiceContainer) *ControllerFactory {
	return &ControllerFactory{
		Container: container,
	}
}
