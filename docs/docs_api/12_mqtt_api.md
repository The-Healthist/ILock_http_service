# MQTT 通讯接口

MQTT 通讯接口用于支持设备与应用程序之间的实时通信，主要用于视频通话、设备状态更新和系统消息推送等功能。

## 主题结构

系统使用以下 MQTT 主题结构：

- 呼叫请求: `calls/request/{device_id}`
- 来电通知: `users/{user_id}/calls/incoming`
- 通话控制(呼叫方): `devices/{device_id}/calls/control`
- 通话控制(被呼叫方): `users/{user_id}/calls/control`
- 设备状态: `devices/{device_id}/status`
- 系统消息: `system/{message_type}`

## 发起通话

- **路径**: `/api/mqtt/call`
- **方法**: POST
- **描述**: 通过 MQTT 向设备关联的住户发起视频通话请求。如果提供了 household_number 参数，则呼叫该户号下的所有住户；如果未提供，则呼叫该设备绑定的户号下的所有住户。
- **注意**: 该接口已简化，移除了通过住户电话呼叫的功能，现在只支持通过设备 ID 和可选的户号参数发起呼叫。
- **参数**:
  ```json
  {
  	"device_id": "1", // 必填，设备ID
  	"household_number": "101", // 可选，指定户号
  	"timestamp": 1651234567890 // 可选，时间戳
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": {
  		"call_id": "call-20250510-abcdef123456",
  		"device_id": "1",
  		"target_resident_ids": ["2", "3"],
  		"call_info": {
  			"call_id": "call-20250510-abcdef123456",
  			"action": "initiated",
  			"timestamp": 1651234567890
  		},
  		"tencen_rtc": {
  			"sdk_app_id": 1400000001,
  			"user_id": "device_1",
  			"user_sig": "eJwtzM1Og0AUhmG...",
  			"room_id": "call-20250510-abcdef123456",
  			"room_id_type": "string"
  		},
  		"timestamp": 1651234567890
  	}
  }
  ```

## 处理呼叫方动作

- **路径**: `/api/mqtt/controller/device`
- **方法**: POST
- **描述**: 处理设备端通话动作，支持的动作类型包括：hangup(挂断)、cancelled(取消呼叫)
- **参数**:
  ```json
  {
  	"call_info": {
  		"call_id": "call-20250510-abcdef123456",
  		"action": "hangup", // 支持：hangup, cancelled
  		"reason": "user_cancelled", // 可选，原因
  		"timestamp": 1651234567890 // 可选，时间戳
  	}
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": null
  }
  ```

## 处理被呼叫方动作

- **路径**: `/api/mqtt/controller/resident`
- **方法**: POST
- **描述**: 处理居民端通话动作，支持的动作类型包括：rejected(拒绝)、answered(接听)、hangup(挂断)、timeout(超时)
- **参数**:
  ```json
  {
  	"call_info": {
  		"call_id": "call-20250510-abcdef123456",
  		"action": "answered", // 支持：rejected, answered, hangup, timeout
  		"reason": "user_busy", // 可选，原因
  		"timestamp": 1651234567890 // 可选，时间戳
  	}
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": null
  }
  ```

## 获取通话会话

- **路径**: `/api/mqtt/session`
- **方法**: GET
- **描述**: 获取通话会话信息及 TRTC 房间详情，包括设备 ID、住户 ID、通话状态、开始时间等
- **参数**:
  - `call_id`: 通话会话 ID (查询参数)
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": {
  		"call_id": "call-20250510-abcdef123456",
  		"device_id": "1",
  		"resident_id": "2",
  		"start_time": "2025-05-10T15:04:05Z",
  		"status": "connected",
  		"last_activity": "2025-05-10T15:09:10Z",
  		"tencen_rtc": {
  			"sdk_app_id": 1400000001,
  			"user_id": "device_1",
  			"user_sig": "eJwtzM1Og0AUhmG...",
  			"room_id": "call-20250510-abcdef123456",
  			"room_id_type": "string"
  		}
  	}
  }
  ```

## 结束通话会话

- **路径**: `/api/mqtt/end-session`
- **方法**: POST
- **描述**: 强制结束通话会话并通知所有参与方，适用于系统管理或异常情况下的通话强制终止
- **参数**:
  ```json
  {
  	"call_id": "call-20250510-abcdef123456",
  	"reason": "call_completed" // 可选，结束原因
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": null
  }
  ```

## 更新设备状态

- **路径**: `/api/mqtt/device/status`
- **方法**: POST
- **描述**: 更新设备状态信息，包括在线状态、电池电量和其他自定义属性，无需 MQTT 连接，系统会通过 MQTT 推送给相关订阅方
- **参数**:
  ```json
  {
  	"device_id": "1",
  	"online": true,
  	"battery": 85,
  	"properties": {
  		// 可选，自定义属性
  		"firmware_version": "1.2.3",
  		"wifi_signal": 78,
  		"temperature": 36.5
  	}
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": null
  }
  ```

## 发布系统消息

- **路径**: `/api/mqtt/system/message`
- **方法**: POST
- **描述**: 通过 MQTT 发布系统消息，支持 info、warning、error 三种级别，消息会推送给所有订阅相关主题的客户端
- **参数**:
  ```json
  {
  	"type": "notification", // 消息类型
  	"level": "info", // 级别：info, warning, error
  	"message": "系统将于今晚22:00进行升级维护", // 消息内容
  	"data": {
  		// 可选，附加数据
  		"maintenance_start": "2025-05-10T22:00:00Z",
  		"maintenance_end": "2025-05-11T02:00:00Z",
  		"affected_services": ["video_call", "door_access"]
  	},
  	"timestamp": 1651234567890 // 可选，时间戳
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": null
  }
  ```

## 错误响应

当 API 调用失败时，将返回以下格式的错误响应：

```json
{
	"code": 400, // HTTP状态码
	"message": "无效的请求参数: 缺少必要字段", // 错误消息
	"data": null
}
```

常见错误码：

- 400: 请求参数错误
- 404: 资源不存在(如通话会话未找到)
- 500: 服务器内部错误
