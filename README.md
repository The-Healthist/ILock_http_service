# ILock HTTP Service

A comprehensive door access management system with video calling capabilities and real-time weather information integration.

## Features

- **User Management**
  - Multi-role authentication (Admin, Property Staff, Resident)
  - Secure JWT-based authentication
  - Role-based access control

- **Door Access Control**
  - Device management and pairing with residents
  - Access logs for security tracking
  - Emergency response system
  
- **Real-time Communication**
  - Video calling between door devices and residents via Aliyun RTC
  - Token-based secure communication
  - In-memory token caching for improved performance
  
- **Weather Information**
  - Real-time weather data for device locations
  - Weather alerts and forecasts
  - Cached weather data for improved performance

- **System Features**
  - Comprehensive logging system
  - Monitoring and health checks
  - Docker containerization for easy deployment

## Technology Stack

- **Backend**: Go with Gin framework
- **Database**: MySQL with GORM ORM
- **Caching**: Redis for token and weather data
- **Communication**: Aliyun RTC for real-time video calls
- **Containerization**: Docker and Docker Compose
- **External APIs**: Weather API integration

## Project Structure

```
.
├── config/             # Configuration management and environment variables
├── controllers/        # HTTP request handlers (RTC, Weather, Auth)
│   ├── base_controller.go
│   ├── jwt_controller.go
│   ├── rtc_controller.go
│   └── weather_controller.go
├── middleware/         # Middleware components (JWT auth, logging)
├── models/             # Database models (GORM)
│   ├── admin.go
│   ├── device.go
│   ├── resident.go
│   └── ... 
├── routes/             # API route definitions and grouping
├── services/           # Business logic
│   ├── aliyun/         # Aliyun RTC integration
│   ├── redis/          # Redis caching service
│   ├── weather/        # Weather API integration
│   └── container.go    # Service container for dependency injection
├── utils/              # Utility functions
├── Dockerfile          # Docker configuration
├── docker-compose.yml  # Docker Compose configuration
├── main.go             # Application entry point
└── go.mod              # Go module definition
```

## Setup and Installation

### Prerequisites

- Go 1.20 or higher
- MySQL
- Redis
- Docker and Docker Compose (optional for containerized deployment)

### Local Development Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd ILock_http_service
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Configure environment variables or update them in `config/config.go`

4. Run the application:
   ```bash
   go run main.go
   ```

### Docker Deployment

1. Build and start the containers:
   ```bash
   docker-compose up -d
   ```

2. The service will be available at `http://localhost:8080`

## API 接口文档

系统提供多种API接口，按功能分类详述如下：

### 1. 认证与授权接口

#### 公共接口

- `POST /api/auth/login` - 用户登录
  - 支持多种角色登录（管理员、物业人员、居民）
  - 返回JWT令牌及用户信息

#### 认证中间件保护的路由

系统使用JWT认证保护各类接口，支持多级权限控制：
- 系统管理员 - 最高权限
- 物业人员 - 物业相关管理权限
- 居民 - 个人相关功能权限
- 设备 - 设备特定功能权限

### 2. 设备管理接口

#### 设备基础操作

- `GET /api/device` - 获取所有设备列表
- `GET /api/device/:id` - 获取单个设备详情
- `POST /api/device` - 添加新设备
- `PUT /api/device/:id` - 更新设备信息
- `DELETE /api/device/:id` - 删除设备

#### 设备状态与操作

- `GET /api/device/:id/status` - 获取设备状态

### 3. 用户管理接口

#### 管理员管理

- `GET /api/admins` - 获取管理员列表
- `GET /api/admins/:id` - 获取管理员详情
- `POST /api/admins` - 创建管理员账户
- `PUT /api/admins/:id` - 更新管理员信息
- `DELETE /api/admins/:id` - 删除管理员账户

#### 物业人员管理

- `GET /api/staff` - 获取物业人员列表
- `GET /api/staff/:id` - 获取物业人员详情
- `POST /api/staff` - 创建物业人员账户
- `PUT /api/staff/:id` - 更新物业人员信息
- `DELETE /api/staff/:id` - 删除物业人员账户

#### 居民管理

- `GET /api/residents` - 获取居民列表
- `GET /api/residents/:id` - 获取居民详情
- `POST /api/residents` - 创建居民账户
- `PUT /api/residents/:id` - 更新居民信息
- `DELETE /api/residents/:id` - 删除居民账户

### 4. 实时通信接口

#### RTC服务

- `POST /api/rtc/token` - 获取RTC通信令牌
- `POST /api/rtc/call` - 发起视频通话

#### 通话记录管理

- `GET /api/calls` - 获取通话记录列表
- `GET /api/calls/:id` - 获取单个通话记录详情
- `GET /api/calls/statistics` - 获取通话统计信息
- `GET /api/calls/device/:deviceId` - 获取指定设备的通话记录
- `GET /api/calls/resident/:residentId` - 获取指定居民的通话记录
- `POST /api/calls/:id/feedback` - 提交通话质量反馈

### 5. 天气服务接口

- `GET /api/weather` - 获取天气信息
- `GET /api/weather/device/:deviceId` - 获取指定设备位置的天气信息

### 6. 紧急情况处理接口

- `POST /api/emergency/alarm` - 触发紧急警报
  - 支持多种类型警报：火灾、入侵、医疗等
  - 记录警报详情并通知相关人员

- `GET /api/emergency/contacts` - 获取紧急联系人列表
  - 按优先级排序的联系人信息

- `POST /api/emergency/unlock-all` - 紧急情况下解锁所有门
  - 紧急疏散等情况使用
  - 记录所有解锁操作

- `POST /api/emergency/notify-all` - 向所有用户发送紧急通知
  - 支持多种重要程度：高、中、低
  - 支持目标群体筛选：全体、居民、物业人员

### 7. 系统监控接口

- `GET /api/ping` - 系统健康检查
  - 用于监控系统可用性

## 后续计划开发接口

以下是系统后续计划开发的API接口：

### 1. 设备扩展接口

- `PUT /api/device/{id}/configuration` - 更新设备配置
- `POST /api/device/{id}/reboot` - 重启设备
- `POST /api/device/{id}/unlock` - 远程开门

### 2. 访客管理

- `POST /api/visitors` - 创建访客邀请
- `GET /api/visitors` - 获取访客列表
- `GET /api/visitors/{id}` - 获取单个访客详情
- `PUT /api/visitors/{id}` - 更新访客信息
- `DELETE /api/visitors/{id}` - 删除访客邀请
- `POST /api/visitors/{id}/approve` - 批准访客邀请
- `POST /api/visitors/{id}/reject` - 拒绝访客邀请
- `GET /api/visitors/qr-code/{id}` - 获取访客二维码

### 3. 通知管理

- `POST /api/notifications` - 发送通知
- `GET /api/notifications` - 获取通知列表
- `GET /api/notifications/{id}` - 获取单个通知详情
- `PUT /api/notifications/{id}/read` - 标记通知为已读
- `DELETE /api/notifications/{id}` - 删除通知
- `POST /api/notifications/batch` - 批量发送通知

### 4. 系统监控与管理

- `GET /api/system/status` - 获取系统状态
- `GET /api/system/statistics` - 获取系统统计信息
- `GET /api/system/logs` - 获取系统日志
- `POST /api/system/backup` - 创建系统备份
- `GET /api/system/backups` - 获取备份列表
- `POST /api/system/restore` - 从备份恢复系统

## License

Copyright © 2024
智能门禁系统后续需要的接口
基于您现有的智能门禁系统，以下是我建议开发的后续接口，按功能模块分类：
1. 用户管理接口
居民(Resident)相关
GET /api/residents - 获取居民列表
GET /api/residents/{id} - 获取单个居民详情
POST /api/residents - 添加新居民
PUT /api/residents/{id} - 更新居民信息
DELETE /api/residents/{id} - 删除居民
GET /api/residents/{id}/access-logs - 获取居民的进出记录
物业人员相关
GET /api/staff - 获取物业人员列表
GET /api/staff/{id} - 获取单个物业人员详情
POST /api/staff - 添加新物业人员
PUT /api/staff/{id} - 更新物业人员信息
DELETE /api/staff/{id} - 删除物业人员
PUT /api/staff/{id}/permissions - 更新物业人员权限
系统管理员相关
GET /api/admins - 获取管理员列表
POST /api/admins - 添加新管理员
PUT /api/admins/{id} - 更新管理员信息
DELETE /api/admins/{id} - 删除管理员
2. 设备管理接口
设备相关
GET /api/devices - 获取所有设备列表
GET /api/devices/{id} - 获取单个设备详情
POST /api/devices - 添加新设备
PUT /api/devices/{id} - 更新设备信息
DELETE /api/devices/{id} - 删除设备
GET /api/devices/{id}/status - 获取设备状态
PUT /api/devices/{id}/configuration - 更新设备配置
POST /api/devices/{id}/reboot - 重启设备
POST /api/devices/{id}/unlock - 远程开门
设备分组管理
GET /api/device-groups - 获取设备分组列表
POST /api/device-groups - 创建设备分组
PUT /api/device-groups/{id} - 更新设备分组
DELETE /api/device-groups/{id} - 删除设备分组
POST /api/device-groups/{id}/devices - 向分组添加设备
DELETE /api/device-groups/{id}/devices/{deviceId} - 从分组移除设备
3. 通话记录管理
GET /api/calls - 获取通话记录列表
GET /api/calls/{id} - 获取单个通话记录详情
GET /api/calls/statistics - 获取通话统计信息
GET /api/calls/device/{deviceId} - 获取指定设备的通话记录
GET /api/calls/resident/{residentId} - 获取指定居民的通话记录
POST /api/calls/{id}/feedback - 提交通话质量反馈
4. 门禁记录管理
GET /api/access-logs - 获取门禁记录列表
GET /api/access-logs/{id} - 获取单个门禁记录详情
GET /api/access-logs/statistics - 获取门禁统计信息
GET /api/access-logs/device/{deviceId} - 获取指定设备的门禁记录
POST /api/access-logs - 手动添加门禁记录(适用于特殊情况)
5. 访客管理
POST /api/visitors - 创建访客邀请
GET /api/visitors - 获取访客列表
GET /api/visitors/{id} - 获取单个访客详情
PUT /api/visitors/{id} - 更新访客信息
DELETE /api/visitors/{id} - 删除访客邀请
POST /api/visitors/{id}/approve - 批准访客邀请
POST /api/visitors/{id}/reject - 拒绝访客邀请
GET /api/visitors/qr-code/{id} - 获取访客二维码
6. 通知管理
POST /api/notifications - 发送通知
GET /api/notifications - 获取通知列表
GET /api/notifications/{id} - 获取单个通知详情
PUT /api/notifications/{id}/read - 标记通知为已读
DELETE /api/notifications/{id} - 删除通知
POST /api/notifications/batch - 批量发送通知
7. 系统监控与管理
GET /api/system/status - 获取系统状态
GET /api/system/statistics - 获取系统统计信息
GET /api/system/logs - 获取系统日志
POST /api/system/backup - 创建系统备份
GET /api/system/backups - 获取备份列表
POST /api/system/restore - 从备份恢复系统
8. 紧急情况处理
POST /api/emergency/alarm - 触发紧急警报
GET /api/emergency/contacts - 获取紧急联系人列表
POST /api/emergency/unlock-all - 紧急情况下解锁所有门
POST /api/emergency/notify-all - 向所有用户发送紧急通知


## 已有接口