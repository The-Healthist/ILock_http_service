# 健康检查接口

## 健康状态检查

- **路径**: `/api/health/ping`
- **方法**: GET
- **描述**: 简单的健康检查端点，用于监控系统是否正常运行
- **响应**:
  ```json
  {
  	"message": "pong",
  	"status": "healthy"
  }
  ```
