# iLock 智能门禁系统

## 项目概述

iLock 是一个基于 Go 语言开发的智能门禁管理系统，提供了强大的门禁控制、视频通话和紧急情况处理功能。系统采用 Docker 容器化部署，便于在各种环境中快速安装和更新。

## 系统架构

- **后端**: Go + Gin 框架 + GORM
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.4.1
- **部署**: Docker + Docker Compose
- **通讯**:
  - RESTful API: 基础业务操作
  - MQTT: 实时消息推送、视频通话信令
  - TRTC: 腾讯云实时音视频

## MQTT 通信架构

### 1. 主题设计

#### 视频通话相关主题

- **呼叫请求**: `calls/request/{device_device_id}`
- **来电通知**: `users/{user_id}/calls/incoming`
- **呼叫方控制**: `devices/{device_device_id}/calls/control`
- **接收方控制**: `users/{user_id}/calls/control`

### 2. 消息质量(QoS)

- 视频通话信令: QoS 1 (至少一次送达)
- 普通通知: QoS 0 (最多一次送达)
- 紧急通知: QoS 2 (确保一次送达)

### 3. 消息格式

所有消息采用 JSON 格式，包含以下基本字段：

- `message_id`: 消息唯一标识
- `timestamp`: 消息时间戳
- `type`: 消息类型
- `payload`: 消息内容

### 4. 实时通信流程

#### 视频通话流程

1. 访客通过门禁设备发起呼叫
2. 后端接收呼叫请求并创建 TRTC 房间
3. 向住户推送来电通知
4. 住户接听/拒绝通话
5. 后端处理响应并通知门禁设备
6. 建立/结束视频通话

#### 紧急情况处理

1. 系统检测到紧急情况
2. 通过紧急通知主题广播警报
3. 相关人员接收通知并处理
4. 系统记录响应情况

## 主要功能

- 用户管理（管理员、物业人员、居民）
- 设备管理（智能门锁监控和控制）
- 视频通话（访客与居民之间的实时沟通）
- 紧急情况处理（火灾、入侵、医疗等紧急事件）
- 完整的认证和权限管理

## 部署指南

### 前置要求

1. **服务器环境**:

   - Linux 服务器（推荐 Ubuntu 20.04 或 CentOS 8）
   - Docker 和 Docker Compose 已安装
   - 开放端口：8080(HTTP), 3310(MySQL), 6380(Redis)

2. **本地环境**（用于部署）:
   - Windows 操作系统
   - 已安装 PuTTY 工具集（包含 pscp.exe 和 plink.exe）

### 快速部署（Windows）

我们提供了一个一键部署脚本`deploy_ilock.bat`，可以自动完成打包、上传和部署过程：

1. **下载部署脚本**并保存到本地项目根目录。

2. **执行部署脚本**：

   - 双击运行`deploy_ilock.bat`
   - 按提示确认或修改服务器配置信息
   - 脚本会自动检查、打包、上传和部署项目

3. **验证部署**：
   - 脚本会自动验证服务是否成功启动
   - 访问`http://服务器IP:8080`检查服务运行状态
   - 访问`http://服务器IP:20033/swagger/index.html`查看 API 文档

### 手动部署

如果你需要手动部署，可以按照以下步骤操作：

1. **克隆代码到本地**：

   ```bash
   git clone <repository-url>
   cd ilock-http-service
   ```

2. **创建环境配置文件**：

   ```bash
   cp .env.example .env
   # 编辑.env文件，设置数据库和JWT等配置
   ```

3. **上传到服务器**：

   ```bash
   scp -r ./* user@server:/path/to/ilock-service/
   ```

4. **在服务器上启动服务**：
   ```bash
   cd /path/to/ilock-service/
   docker-compose up -d
   ```

### 服务器配置参考

默认配置如下，可以根据需要在部署时修改：

- **服务器**:

  - IP:
  - 部署目录: /root/ilock
  - SSH 端口: 22

- **数据库**:

  - 主机: localhost
  - 端口: 3309
  - 用户: root
  - 密码:
  - 数据库名: ilock_db

- **Docker 镜像**:
  - 镜像仓库: https://goproxy.cn,direct
  - 服务端口: 20033

## 更新与维护

### 使用部署脚本更新

当系统需要更新时，你可以使用同样的部署脚本：

1. 将需要更新的文件准备好（例如`controller/base/building_controller.go`等）
2. 运行部署脚本，它会自动创建备份并更新文件
3. 系统会自动验证更新是否成功，如果失败会自动回滚

### 常用维护命令

- **查看服务日志**:

  ```bash
  docker-compose logs -f backend
  ```

- **重启服务**:

  ```bash
  docker-compose restart backend
  ```

- **回滚到之前版本**:
  ```bash
  cd /root/ilock && \
  cp -r backup/controller/base/* controller/base/ && \
  cp -r backup/models/base/* models/base/ && \
  cp -r backup/services/base/* services/base/ && \
  docker-compose down && \
  docker-compose build backend && \
  docker-compose up -d
  ```

## API 文档

系统集成了 Swagger 文档，部署后可以通过以下地址访问：

http://服务器 IP:20033/swagger/index.html

主要 API 端点包括：

- **认证**: `/api/auth/login`
- **管理员**: `/api/admins/*`
- **物业人员**: `/api/staffs/*`
- **居民**: `/api/residents/*`
- **设备**: `/api/devices/*`
- **通话记录**: `/api/calls/*`
- **紧急情况**: `/api/emergency/*`

## MQTT 通讯协议接口文档

### 通讯架构

iLock 系统使用 MQTT 协议实现实时消息推送和视频通话信令传输。下面详细说明 MQTT 通讯协议的各个方面：

### 1. MQTT 主题设计

#### 视频通话相关主题

- **呼叫请求**
- **来电通知**
- **通话控制**
- **通话控制**
- **设备状态** :

- **系统消息**:

### 2. api 设计以及参数

所有消息载荷均使用 JSON 格式。

#### 呼叫请求 post(`mqtt_call/call`)

**方向**: device -> Backend

```json
{
  "device_device_id": "device_device_id",                  // 呼叫方设备ID
  "target_resident_id": "resident_resident_id",              // 目标用户ID
  "timestamp": 1678886400000                               // 发起呼叫的Unix毫秒时间戳
}
// 响应:
{
  "call_id": "unique_call_identifier_generated_by_device", // 本次呼叫的唯一id
  "device_device_id": "device_device_id",                  // 呼叫方设备ID
  "target_resident_id": "resident_resident_id",
  "timestamp": 1678886400000,
  "tencen_rtc": {                                          // 加入TRTC房间所需信息
    "room_id_type":"number",
    "room_id": "trtc_room_id_created_by_backend",          // TRTC房间号
    "sdk_app_id": 1400000000,                              // TRTC应用ID
    "user_id": "resident_user_id",                           // 被呼叫方在TRTC中使用的
    "user_sig": "generated_user_signature_for_resident"      // 被呼叫方的TRTC签名
  },
  "call_info":{
	"action": "ringing",                                     // 控制动作类型
	"call_id": "unique_call_identifier_from_request",        // 对应呼叫请求的ID
	"timestamp": 1678886500000,                              // 发送指令的Unix毫秒时间戳
	"reason": "Optional message for details"                 // 可选，提供额外信息
}

}
```

#### 来电通知 post(`mqtt_call/incoming`)

**方向**: Backend -> mqtt ->resident : 信息的转发

```json
{
  "call_id": "unique_call_identifier_generated_by_device", // 本次呼叫的唯一id
  "device_device_id": "device_device_id",                  // 呼叫方设备ID
  "target_resident_id": "resident_resident_id",
  "timestamp": 1678886400000
  "tencen_rtc": {                                          // 加入TRTC房间所需信息
    "room_id_type":"number",
    "room_id": "trtc_room_id_created_by_backend",          // TRTC房间号
    "sdk_app_id": 1400000000,                              // TRTC应用ID
    "user_id": "resident_user_id",                           // 被呼叫方在TRTC中使用的
    "user_sig": "generated_user_signature_for_resident"      // 被呼叫方的TRTC签名
  },
}
```

#### 通话控制(接听前) - 呼叫方挂断按钮 post(`mqtt_call/controller/device`)

**方向**:device -> Backend -> mqtt -> target_resident(已挂断)

```json
"call_info":{
	"action": "hangup", // 控制动作类型
	"call_id": "unique_call_identifier_from_request", // 对应呼叫请求的ID
	"timestamp": 1678886500000, // 发送指令的Unix毫秒时间戳
	"reason": "Optional message for details" // 可选，提供额外信息
}
```

**action 说明**:

- 'reveive':接听了
- `ringing`: resident 正在被呼叫（已发送 incoming 通知）
- `rejected`: resident 拒绝了通话
- `hangup`: resident 挂断了通话
- `timeout`: resident 无应答超时
- `error`: 处理过程中发生错误（如创建房间失败）

#### 通话控制(接听前) - 被呼叫方挂断按钮 post(`mqtt_call/controller/resident`)

**方向**:target_resident(已挂断) -> Backend -> mqtt -> device

```json
"call_info":{
	"action": "rejected", // 控制动作类型
	"call_id": "unique_call_identifier_from_request", // 对应呼叫请求的ID
	"timestamp": 1678886500000, // 发送指令的Unix毫秒时间戳
	"reason": "Optional message for details" // 可选，提供额外信息
}
```

#### 通话控制(接听前) - 被呼叫方接听按钮 post(`mqtt_call/controller/resident`)

**方向**:target_resident -> Backend -> mqtt -> device

```json
  "call_info":{
	"action": "rejected", // 控制动作类型
	"call_id": "unique_call_identifier_from_request", // 对应呼叫请求的ID
	"timestamp": 1678886500000, // 发送指令的Unix毫秒时间戳
	"reason": "Optional message for details" // 可选，提供额外信息
}
```

resident -> 开始进入视频通话
mqtt 进程再转发給 device -> 开始进入视频通话

#### 接听以后的通话控制 - 呼叫方挂断按钮 post(`mqtt_call/controller/device`)

**方向**: device -> Backend -> mqtt -> target_resident

```json
{
	"action": "hangup", // 控制动作类型
	"call_id": "unique_call_identifier_from_request", // 对应呼叫请求的ID
	"timestamp": 1678886550000, // 发送指令的Unix毫秒时间戳
	"reason": "Optional message for details" // 可选，提供额外信息
}
```

#### 接听以后的通话控制 - 被呼叫方挂断按钮 post(`mqtt_call/controller/resident`)

**方向**: target_resident -> Backend -> mqtt -> target_resident

```json
{
	"action": "hangup", // 控制动作类型
	"call_id": "unique_call_identifier_from_request", // 对应呼叫请求的ID
	"timestamp": 1678886550000, // 发送指令的Unix毫秒时间戳
	"reason": "Optional message for details" // 可选，提供额外信息
}
```

```json
{
	"type": "device_offline", // 消息类型
	"level": "warning", // 消息级别: info, warning, error
	"message": "门口设备离线", // 消息内容
	"timestamp": 1678886700000, // 发送时间戳
	"data": {
		// 额外数据（可选）
		"device_id": "device123",
		"last_seen": 1682570000000
	}
}
```

### 3. 上诉任何情况下都需要创建通话记录

**方向**: Backend

```json

```

### 4. 安全性考虑

- 所有 MQTT 通信使用 TLS 加密
- 客户端需要使用用户名/密码或证书进行身份验证
- 主题设计确保信息隔离，防止未授权访问

## 系统特性

- **自动备份与回滚**: 在更新前自动创建备份，更新失败时自动回滚
- **基于角色的访问控制**: 系统管理员、物业人员和居民具有不同的权限
- **安全通信**: 基于 JWT 的 API 认证
- **视频通话集成**: 集成阿里云 RTC 提供实时视频通话
- **容器化部署**: 使用 Docker 和 Docker Compose 简化部署和维护
- **紧急响应系统**: 快速处理火灾、入侵等紧急情况
- **健康检查**: 服务健康状态监控，确保系统稳定运行

## 故障排除

1. **服务无法启动**:

   - 检查 Docker 和 Docker Compose 是否正确安装
   - 检查端口是否被占用: `netstat -tunlp`
   - 查看容器日志: `docker-compose logs backend`

2. **数据库连接失败**:

   - 检查数据库配置是否正确
   - 确认数据库服务是否运行: `docker-compose ps mysql`
   - 尝试手动连接数据库验证凭据

3. **API 响应错误**:

   - 检查 JWT 密钥配置
   - 确认请求格式是否正确
   - 查看服务日志了解详细错误信息

4. **部署脚本执行失败**:
   - 确保 PuTTY 工具集已正确安装并在 PATH 中
   - 检查服务器连接信息是否正确
   - 验证本地文件存在且权限正确

## 许可证

版权所有 © 2024 iLock 开发团队
