package controllers

import (
	"github.com/gin-gonic/gin"

	"ilock-http-service/internal/error/response"
)

// HealthCheckController 健康检查控制器
type HealthCheckController struct{}

// NewHealthCheckController 创建健康检查控制器实例
func NewHealthCheckController() *HealthCheckController {
	return &HealthCheckController{}
}

// Ping 健康检查端点
func (h *HealthCheckController) Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"status":  "healthy",
		"message": "pong",
	})
}
