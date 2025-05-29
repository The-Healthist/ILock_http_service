# 紧急情况接口

## 获取紧急情况日志

- **路径**: `/api/emergency`
- **方法**: GET
- **描述**: 获取所有紧急情况日志
- **响应**: 紧急情况日志列表

## 获取紧急情况详情

- **路径**: `/api/emergency/:id`
- **方法**: GET
- **描述**: 根据 ID 获取紧急情况详情
- **响应**: 紧急情况详情

## 更新紧急情况

- **路径**: `/api/emergency/:id`
- **方法**: PUT
- **描述**: 更新紧急情况状态
- **响应**: 更新结果

## 触发紧急情况

- **路径**: `/api/emergency/trigger`
- **方法**: POST
- **描述**: 触发紧急情况
- **响应**: 触发结果

## 触发警报

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

## 获取紧急联系人

- **路径**: `/api/emergency/contacts`
- **方法**: GET
- **描述**: 获取系统中所有紧急联系人
- **响应**: 紧急联系人列表

## 通知所有用户

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

## 紧急情况解锁所有门

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
