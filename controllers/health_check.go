package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheckController 健康检查控制器
type HealthCheckController struct{}

// NewHealthCheckController 创建健康检查控制器实例
func NewHealthCheckController() *HealthCheckController {
	return &HealthCheckController{}
}

// Ping 健康检查端点
func (h *HealthCheckController) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"status":  "healthy",
	})
}
