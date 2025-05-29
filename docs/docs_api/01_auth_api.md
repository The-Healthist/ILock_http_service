# 认证接口

## 用户登录

- **路径**: `/api/auth/login`
- **方法**: POST
- **描述**: 处理用户登录并根据用户角色返回不同权限的 JWT 令牌
- **参数**:
  ```json
  {
  	"username": "admin",
  	"password": "admin123"
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "Login successful",
  	"data": {
  		"user_id": 1,
  		"username": "admin",
  		"role": "admin",
  		"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  		"created_at": "2023-01-01T00:00:00Z"
  	}
  }
  ```
