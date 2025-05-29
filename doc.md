# ILock HTTP Service API 文档

## 目录

- [认证接口](#认证接口)
- [管理员接口](#管理员接口)
- [设备接口](#设备接口)
- [居民接口](#居民接口)
- [物业员工接口](#物业员工接口)
- [通话记录接口](#通话记录接口)
- [紧急情况接口](#紧急情况接口)
- [楼号接口](#楼号接口)
- [户号接口](#户号接口)
- [音视频通话接口](#音视频通话接口)

## 认证接口

### 用户登录

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

## 管理员接口

### 获取管理员列表

- **路径**: `/api/admin`
- **方法**: GET
- **描述**: 分页获取所有管理员用户列表
- **参数**:
  - `page`: 页码，默认 1
  - `page_size`: 每页条数，默认 10
  - `search`: 搜索关键词
- **响应**: 管理员列表

### 获取管理员详情

- **路径**: `/api/admin/:id`
- **方法**: GET
- **描述**: 根据 ID 获取特定管理员的详细信息
- **响应**: 管理员详情

### 创建管理员

- **路径**: `/api/admin`
- **方法**: POST
- **描述**: 创建一个新的管理员用户
- **参数**:
  ```json
  {
  	"username": "admin123",
  	"password": "Admin@123",
  	"phone": "13800138000",
  	"email": "admin@example.com"
  }
  ```
- **响应**: 创建的管理员信息

### 更新管理员

- **路径**: `/api/admin/:id`
- **方法**: PUT
- **描述**: 更新现有管理员用户的信息
- **参数**:
  ```json
  {
  	"phone": "13800138000",
  	"email": "admin@example.com",
  	"password": "NewPassword@123"
  }
  ```
- **响应**: 更新后的管理员信息

### 删除管理员

- **路径**: `/api/admin/:id`
- **方法**: DELETE
- **描述**: 删除指定 ID 的管理员用户
- **响应**: 操作结果

## 设备接口

### 获取设备列表

- **路径**: `/api/devices`
- **方法**: GET
- **描述**: 获取所有设备的列表，支持按类型和楼号筛选
- **参数**:
  - `device_type`: 设备类型（resident, house）
  - `building_id`: 楼号 ID
- **响应**: 设备列表

### 获取设备详情

- **路径**: `/api/devices/:id`
- **方法**: GET
- **描述**: 根据 ID 获取设备信息
- **响应**: 设备详情

### 创建设备

- **路径**: `/api/devices`
- **方法**: POST
- **描述**: 创建一个新的门禁设备，支持设备类型和关联
- **参数**:
  ```json
  {
  	"name": "门禁1号",
  	"serial_number": "SN12345678",
  	"status": "online",
  	"location": "小区北门入口",
  	"device_type": "resident",
  	"building_id": 1,
  	"household_ids": [1, 2],
  	"staff_ids": [1, 2, 3]
  }
  ```
- **响应**: 创建的设备信息

### 更新设备

- **路径**: `/api/devices/:id`
- **方法**: PUT
- **描述**: 根据 ID 更新设备信息，支持更新设备类型和关联
- **参数**: 同创建设备
- **响应**: 更新后的设备信息

### 删除设备

- **路径**: `/api/devices/:id`
- **方法**: DELETE
- **描述**: 根据 ID 删除设备
- **响应**: 操作结果

### 获取设备状态

- **路径**: `/api/devices/:id/status`
- **方法**: GET
- **描述**: 获取设备的当前状态信息，包括在线状态、最后更新时间等
- **响应**: 设备状态信息

### 设备健康检测

- **路径**: `/api/device/status`
- **方法**: POST
- **描述**: 设备用于报告在线状态的简单健康检测接口
- **参数**:
  ```json
  {
  	"device_id": "1"
  }
  ```
- **响应**: 设备状态更新结果

### 关联设备与楼号

- **路径**: `/api/devices/:id/building`
- **方法**: POST
- **描述**: 将指定设备关联到楼号
- **参数**:
  ```json
  {
  	"building_id": 1
  }
  ```
- **响应**: 关联结果

### 获取设备关联的户号

- **路径**: `/api/devices/:id/households`
- **方法**: GET
- **描述**: 获取指定设备关联的所有户号
- **响应**: 户号列表

### 关联设备与户号

- **路径**: `/api/devices/:id/households`
- **方法**: POST
- **描述**: 将指定设备关联到户号
- **参数**:
  ```json
  {
  	"household_id": 1
  }
  ```
- **响应**: 关联结果

### 解除设备与户号的关联

- **路径**: `/api/devices/:id/households/:household_id`
- **方法**: DELETE
- **描述**: 解除指定设备与户号的关联
- **响应**: 操作结果

## 居民接口

### 获取居民列表

- **路径**: `/api/residents`
- **方法**: GET
- **描述**: 获取系统中所有居民的列表
- **参数**:
  - `page`: 页码，默认 1
  - `page_size`: 每页条数，默认 10
- **响应**: 居民列表

### 获取居民详情

- **路径**: `/api/residents/:id`
- **方法**: GET
- **描述**: 根据 ID 获取特定居民的详细信息
- **响应**: 居民详情

### 创建居民

- **路径**: `/api/residents`
- **方法**: POST
- **描述**: 创建新的居民账户，需要关联到特定设备
- **参数**:
  ```json
  {
  	"name": "张三",
  	"phone": "13812345678",
  	"email": "zhangsan@resident.com",
  	"device_id": 101
  }
  ```
- **响应**: 创建的居民信息

### 更新居民

- **路径**: `/api/residents/:id`
- **方法**: PUT
- **描述**: 更新现有居民的信息
- **参数**:
  ```json
  {
  	"name": "李四",
  	"phone": "13987654321",
  	"email": "lisi@resident.com",
  	"device_id": 102
  }
  ```
- **响应**: 更新后的居民信息

### 删除居民

- **路径**: `/api/residents/:id`
- **方法**: DELETE
- **描述**: 删除指定 ID 的居民
- **响应**: 操作结果

## 物业员工接口

### 获取物业员工列表

- **路径**: `/api/staffs`
- **方法**: GET
- **描述**: 获取所有物业员工的列表，支持分页和搜索
- **参数**:
  - `page`: 页码，默认 1
  - `page_size`: 每页条数，默认 10
  - `search`: 搜索关键词
- **响应**: 物业员工列表

### 获取带设备信息的物业员工列表

- **路径**: `/api/staffs/with-devices`
- **方法**: GET
- **描述**: 获取所有物业员工的列表及其关联的设备信息
- **参数**: 同获取物业员工列表
- **响应**: 带设备信息的物业员工列表

### 获取物业员工详情

- **路径**: `/api/staffs/:id`
- **方法**: GET
- **描述**: 根据 ID 获取特定物业员工的详细信息
- **响应**: 物业员工详情

### 创建物业员工

- **路径**: `/api/staffs`
- **方法**: POST
- **描述**: 创建一个新的物业员工
- **参数**:
  ```json
  {
  	"name": "王物业",
  	"phone": "13700001234",
  	"property_name": "阳光花园小区",
  	"position": "物业经理",
  	"role": "manager",
  	"status": "active",
  	"remark": "负责A区日常管理工作",
  	"username": "wangwuye",
  	"password": "Property@123",
  	"device_ids": [1, 2, 3]
  }
  ```
- **响应**: 创建的物业员工信息

### 更新物业员工

- **路径**: `/api/staffs/:id`
- **方法**: PUT
- **描述**: 更新现有物业员工的信息
- **参数**:
  ```json
  {
  	"name": "李物业",
  	"phone": "13700005678",
  	"property_name": "幸福家园小区",
  	"position": "前台客服",
  	"role": "staff",
  	"status": "active",
  	"remark": "负责接待访客和处理居民投诉",
  	"username": "liwuye",
  	"password": "NewProperty@456",
  	"device_ids": [1, 3, 5]
  }
  ```
- **响应**: 更新后的物业员工信息

### 删除物业员工

- **路径**: `/api/staffs/:id`
- **方法**: DELETE
- **描述**: 删除指定 ID 的物业员工
- **响应**: 操作结果

## 通话记录接口

### 获取通话记录列表

- **路径**: `/api/call-records`
- **方法**: GET
- **描述**: 获取系统中所有通话记录，支持分页
- **参数**:
  - `page`: 页码，默认 1
  - `page_size`: 每页条数，默认 10
- **响应**: 通话记录列表

### 获取通话统计信息

- **路径**: `/api/call-records/statistics`
- **方法**: GET
- **描述**: 获取通话统计信息，包括总数、已接、未接等
- **响应**: 通话统计信息

### 获取设备通话记录

- **路径**: `/api/call-records/device/:deviceId`
- **方法**: GET
- **描述**: 获取特定设备的通话记录，支持分页
- **参数**:
  - `page`: 页码，默认 1
  - `page_size`: 每页条数，默认 10
- **响应**: 设备通话记录列表

### 获取居民通话记录

- **路径**: `/api/call-records/resident/:residentId`
- **方法**: GET
- **描述**: 获取特定居民的通话记录，支持分页
- **参数**:
  - `page`: 页码，默认 1
  - `page_size`: 每页条数，默认 10
- **响应**: 居民通话记录列表

### 通过 CallID 获取通话记录

- **路径**: `/api/call-records/session`
- **方法**: GET
- **描述**: 通过 CallID（MQTT 会话 ID）获取特定通话记录的详细信息
- **参数**:
  - `call_id`: 通话会话 ID
- **响应**: 通话记录详情

### 获取通话记录详情

- **路径**: `/api/call-records/:id`
- **方法**: GET
- **描述**: 根据 ID 获取特定通话记录的详细信息
- **响应**: 通话记录详情

### 提交通话反馈

- **路径**: `/api/call-records/:id/feedback`
- **方法**: POST
- **描述**: 为特定通话记录提交质量反馈
- **参数**:
  ```json
  {
  	"rating": 4,
  	"comment": "通话质量良好，声音清晰",
  	"issues": "偶尔有一点延迟"
  }
  ```
- **响应**: 反馈提交结果

## 紧急情况接口

### 获取紧急情况日志

- **路径**: `/api/emergency`
- **方法**: GET
- **描述**: 获取所有紧急情况日志
- **响应**: 紧急情况日志列表

### 获取紧急情况详情

- **路径**: `/api/emergency/:id`
- **方法**: GET
- **描述**: 根据 ID 获取紧急情况详情
- **响应**: 紧急情况详情

### 更新紧急情况

- **路径**: `/api/emergency/:id`
- **方法**: PUT
- **描述**: 更新紧急情况状态
- **响应**: 更新结果

### 触发紧急情况

- **路径**: `/api/emergency/trigger`
- **方法**: POST
- **描述**: 触发紧急情况
- **响应**: 触发结果

### 触发警报

- **路径**: `/api/emergency/alarm`
- **方法**: POST
- **描述**: 触发紧急警报并通知相关人员
- **参数**:
  ```json
  {
  	"type": "fire",
  	"location": "Building A, Floor 3",
  	"property_id": 1,
  	"reported_by": 1,
  	"description": "火灾警报被触发，疑似厨房起火"
  }
  ```
- **响应**: 触发结果

### 获取紧急联系人

- **路径**: `/api/emergency/contacts`
- **方法**: GET
- **描述**: 获取系统中所有紧急联系人
- **响应**: 紧急联系人列表

### 通知所有用户

- **路径**: `/api/emergency/notify-all`
- **方法**: POST
- **描述**: 在紧急情况下向所有用户发送通知
- **参数**:
  ```json
  {
  	"title": "紧急通知：小区火灾警报",
  	"content": "A栋3楼发生火灾，请所有居民立即疏散。",
  	"severity": "high",
  	"target_type": "all",
  	"property_id": 1,
  	"is_public": false,
  	"expires_at": "2023-07-01T15:00:00Z"
  }
  ```
- **响应**: 通知结果

### 紧急情况解锁所有门

- **路径**: `/api/emergency/unlock-all`
- **方法**: POST
- **描述**: 在紧急情况下解锁系统中的所有门
- **参数**:
  ```json
  {
  	"reason": "火灾疏散"
  }
  ```
- **响应**: 解锁结果

## 楼号接口

### 获取楼号列表

- **路径**: `/api/buildings`
- **方法**: GET
- **描述**: 获取系统中所有楼号的列表
- **参数**:
  - `page`: 页码，默认 1
  - `page_size`: 每页条数，默认 10
- **响应**: 楼号列表

### 获取楼号详情

- **路径**: `/api/buildings/:id`
- **方法**: GET
- **描述**: 根据 ID 获取楼号详细信息
- **响应**: 楼号详情

### 创建楼号

- **路径**: `/api/buildings`
- **方法**: POST
- **描述**: 创建一个新的楼号
- **参数**:
  ```json
  {
  	"building_name": "1号楼",
  	"building_code": "B001",
  	"address": "小区东南角",
  	"status": "active"
  }
  ```
- **响应**: 创建的楼号信息

### 更新楼号

- **路径**: `/api/buildings/:id`
- **方法**: PUT
- **描述**: 更新楼号信息
- **参数**: 同创建楼号
- **响应**: 更新后的楼号信息

### 删除楼号

- **路径**: `/api/buildings/:id`
- **方法**: DELETE
- **描述**: 删除指定的楼号
- **响应**: 操作结果

### 获取楼号关联的设备

- **路径**: `/api/buildings/:id/devices`
- **方法**: GET
- **描述**: 获取指定楼号关联的所有设备
- **响应**: 设备列表

### 获取楼号下的户号

- **路径**: `/api/buildings/:id/households`
- **方法**: GET
- **描述**: 获取指定楼号下的所有户号
- **响应**: 户号列表

## 户号接口

### 获取户号列表

- **路径**: `/api/households`
- **方法**: GET
- **描述**: 获取系统中所有户号的列表
- **参数**:
  - `page`: 页码，默认 1
  - `page_size`: 每页条数，默认 10
  - `building_id`: 楼号 ID，用于筛选特定楼号下的户号
- **响应**: 户号列表

### 获取户号详情

- **路径**: `/api/households/:id`
- **方法**: GET
- **描述**: 根据 ID 获取户号详细信息
- **响应**: 户号详情

### 创建户号

- **路径**: `/api/households`
- **方法**: POST
- **描述**: 创建一个新的户号，需要关联到楼号
- **参数**:
  ```json
  {
  	"household_number": "1-1-101",
  	"building_id": 1,
  	"status": "active"
  }
  ```
- **响应**: 创建的户号信息

### 更新户号

- **路径**: `/api/households/:id`
- **方法**: PUT
- **描述**: 更新户号信息
- **参数**: 同创建户号
- **响应**: 更新后的户号信息

### 删除户号

- **路径**: `/api/households/:id`
- **方法**: DELETE
- **描述**: 删除指定的户号
- **响应**: 操作结果

### 获取户号关联的设备

- **路径**: `/api/households/:id/devices`
- **方法**: GET
- **描述**: 获取指定户号关联的所有设备
- **响应**: 设备列表

### 获取户号下的居民

- **路径**: `/api/households/:id/residents`
- **方法**: GET
- **描述**: 获取指定户号下的所有居民
- **响应**: 居民列表

### 关联户号与设备

- **路径**: `/api/households/:id/devices`
- **方法**: POST
- **描述**: 将指定户号关联到设备
- **参数**:
  ```json
  {
  	"device_id": 1
  }
  ```
- **响应**: 关联结果

### 解除户号与设备的关联

- **路径**: `/api/households/:id/devices/:device_id`
- **方法**: DELETE
- **描述**: 解除指定户号与设备的关联
- **响应**: 操作结果

## 音视频通话接口

### 发起 MQTT 通话

- **路径**: `/api/mqtt/call`
- **方法**: POST
- **描述**: 通过 MQTT 向关联设备的所有居民发起视频通话请求
- **参数**:
  ```json
  {
  	"device_device_id": "1",
  	"target_resident_id": "2",
  	"timestamp": 1651234567890
  }
  ```
- **响应**: 通话会话信息

### 处理 MQTT 呼叫方动作

- **路径**: `/api/mqtt/controller/device`
- **方法**: POST
- **描述**: 处理设备端通话动作(挂断、取消等)
- **参数**:
  ```json
  {
  	"call_info": {
  		"call_id": "call-20250510-abcdef123456",
  		"action": "answered",
  		"reason": "user_busy",
  		"timestamp": 1651234567890
  	}
  }
  ```
- **响应**: 处理结果

### 处理 MQTT 被呼叫方动作

- **路径**: `/api/mqtt/controller/resident`
- **方法**: POST
- **描述**: 处理居民端通话动作(接听、拒绝、挂断、超时等)
- **参数**: 同处理 MQTT 呼叫方动作
- **响应**: 处理结果

### 获取 MQTT 通话会话

- **路径**: `/api/mqtt/session`
- **方法**: GET
- **描述**: 获取通话会话信息及 TRTC 房间详情
- **参数**:
  - `call_id`: 通话会话 ID
- **响应**: 通话会话详情

### 结束 MQTT 通话会话

- **路径**: `/api/mqtt/end-session`
- **方法**: POST
- **描述**: 强制结束通话会话并通知所有参与方
- **参数**:
  ```json
  {
  	"call_id": "call-20250510-abcdef123456",
  	"reason": "call_completed"
  }
  ```
- **响应**: 结束结果

### 更新设备状态

- **路径**: `/api/mqtt/device/status`
- **方法**: POST
- **描述**: 更新设备状态信息，包括在线状态、电池电量和其他自定义属性
- **参数**:
  ```json
  {
  	"device_id": "1",
  	"online": true,
  	"battery": 85,
  	"properties": {}
  }
  ```
- **响应**: 更新结果

### 发布系统消息

- **路径**: `/api/mqtt/system/message`
- **方法**: POST
- **描述**: 通过 MQTT 发布系统消息
- **参数**:
  ```json
  {
  	"type": "notification",
  	"level": "info",
  	"message": "系统将于今晚22:00进行升级维护",
  	"timestamp": 1651234567890,
  	"data": {}
  }
  ```
- **响应**: 发布结果

### 获取 RTC 令牌

- **路径**: `/api/rtc/token`
- **方法**: POST
- **描述**: 获取用于实时通信的 RTC 令牌
- **参数**:
  ```json
  {
  	"user_id": "user456",
  	"channel_id": "room123"
  }
  ```
- **响应**: RTC 令牌信息

### 开始视频通话

- **路径**: `/api/rtc/call`
- **方法**: POST
- **描述**: 在设备和居民之间发起视频通话
- **参数**:
  ```json
  {
  	"device_id": "1",
  	"resident_id": "2"
  }
  ```
- **响应**: 通话会话信息

### 获取腾讯云 UserSig

- **路径**: `/api/trtc/usersig`
- **方法**: POST
- **描述**: 获取腾讯云实时通信的 UserSig 凭证
- **参数**:
  ```json
  {
  	"user_id": "user123"
  }
  ```
- **响应**: UserSig 信息

### 开始腾讯视频通话

- **路径**: `/api/trtc/call`
- **方法**: POST
- **描述**: 在设备和居民之间发起腾讯云视频通话
- **参数**:
  ```json
  {
  	"device_id": "1",
  	"resident_id": "2"
  }
  ```
- **响应**: 通话会话信息
