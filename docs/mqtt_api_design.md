# MQTT通讯接口设计

## 一、主题结构

系统使用以下固定主题：

1. `mqtt_call/incoming`
   - 用途：发送来电通知
   - QoS：1
   - 消息格式：
     ```json
     {
         "call_id": "string",
         "device_device_id": "string",
         "target_resident_id": "string",
         "timestamp": 1746870072136,
         "tencen_rtc": {
             "room_id_type": "string",
             "room_id": "string",
             "sdk_app_id": 1600084384,
             "user_id": "string",
             "user_sig": "string"
         }
     }
     ```

2. `mqtt_call/controller/device`
   - 用途：设备端控制消息
   - QoS：1
   - 消息格式：
     ```json
     {
         "action": "string",  // ringing/answered/hangup
         "call_id": "string",
         "timestamp": 1746870072167,
         "reason": "string"
     }
     ```

3. `mqtt_call/controller/resident`
   - 用途：住户端控制消息
   - QoS：1
   - 消息格式：
     ```json
     {
         "action": "string",  // answered/rejected/hangup/timeout
         "call_id": "string",
         "timestamp": 1746870108846,
         "reason": "string"
     }
     ```

4. `mqtt_call/system`
   - 用途：系统消息广播
   - QoS：1
   - 消息格式：
     ```json
     {
         "type": "string",
         "level": "string",  // info/warning/error
         "message": "string",
         "data": {},
         "timestamp": 1746870072136
     }
     ```

## 二、通话流程

1. 发起呼叫
   - 设备通过HTTP API发起呼叫
   - 后端发送来电通知到`mqtt_call/incoming`
   - 后端发送振铃状态到`mqtt_call/controller/device`

2. 通知住户
   - 后端服务通过MQTT主题向住户App发送来电通知
   - 通知包含通话ID、设备信息和TRTC房间信息

3. 住户响应
   - 住户App通过HTTP API响应通话
   - 响应包含action指定操作（接听、拒绝等）
   - 后端服务通过MQTT主题向设备发送状态更新

4. 通话控制
   - 双方可通过HTTP API发送控制命令
   - 后端通过对应控制主题转发命令
   - 接听后为resident和device端同步发布相同信息

5. 结束通话
   - 通话结束后，后端服务更新通话记录，释放TRTC资源
   - 如果通话超时未接听，系统自动结束通话会话
   - 自动创建通话记录

## 三、MQTT通话完整流程详解

### 1. 通话架构概述

MQTT通话系统采用发布/订阅模式，由以下几个核心组件构成：

- **MQTT服务器**：作为消息代理，负责消息的路由和分发
- **后端服务**：处理业务逻辑，管理通话会话，发布和接收MQTT消息
- **设备端**：门禁设备，发起呼叫并参与通话
- **住户App**：接收来电通知，响应通话请求

### 2. 通话流程详细说明

#### 2.1 初始化连接

1. 服务启动时，后端服务连接MQTT服务器
2. 连接成功后，订阅所有相关主题
3. 设置相应的消息处理程序

#### 2.2 发起呼叫流程

1. 设备通过HTTP接口发起呼叫请求
2. 控制器处理呼叫请求，调用相应的服务方法
3. 服务层创建通话会话，生成通话ID
4. 创建TRTC房间并生成相关签名信息
5. 通过MQTT发送来电通知给住户
6. 发送振铃状态消息给设备
7. 创建通话记录

#### 2.3 接收和响应呼叫

1. 住户App接收到来电通知（通过订阅`mqtt_call/incoming`主题）
2. 住户响应通话（接听、拒绝等）
3. 控制器处理住户响应，验证动作类型
4. 服务层处理住户响应，更新会话状态
5. 发送相应的控制消息给设备
6. 根据动作类型处理结束通话逻辑

#### 2.4 通话控制和结束

1. 设备端发送控制命令（如挂断）
2. 或者通过统一的结束会话接口结束通话
3. 控制器处理结束通话请求
4. 服务层处理结束通话，更新会话状态
5. 向双方发送通话结束通知
6. 更新通话记录，释放资源

#### 2.5 通话超时处理

1. 系统定时清理超时会话
2. 对于振铃状态的会话，超时时间为30秒
3. 对于已接通的会话，超时时间为2小时
4. 超时后自动结束会话并通知双方

#### 2.6 系统消息广播

1. 通过API接口发送系统消息
2. 控制器处理系统消息请求，验证消息级别
3. 服务层将消息发布到系统消息主题
4. 所有订阅该主题的客户端都会收到消息

### 3. 完整通话示例

一次完整的通话流程示例：

1. **发起呼叫**:
   - 设备发送HTTP请求：`POST /api/mqtt/call`
   - 后端创建通话会话并返回call_id
   - 后端通过MQTT向住户发送来电通知
   - 住户App接收到通知并显示来电界面

2. **住户响应**:
   - 住户接听通话：`POST /api/mqtt/controller/resident` (action=answered)
   - 后端通过MQTT向设备发送状态更新
   - 设备接收到通知并建立视频连接

3. **通话中控制**:
   - 住户或设备可以发送控制命令
   - 如设备挂断：`POST /api/mqtt/controller/device` (action=hangup)

4. **结束通话**:
   - 后端向双方发送通话结束通知
   - 更新通话记录
   - 释放TRTC资源

### 客户端MQTT订阅说明

#### 设备端订阅
- **主题**: `mqtt_call/controller/device`
- **QoS**: 1
- **用途**: 接收通话控制指令，包括振铃、接听、挂断等
- **处理方式**: 
  - 收到`ringing`时，播放铃声，准备视频通话
  - 收到`answered`时，建立TRTC连接
  - 收到`hangup`或`ended`时，结束通话，释放资源

#### 住户端订阅
- **主题**: `mqtt_call/incoming`
- **QoS**: 1
- **用途**: 接收来电通知，包含通话ID和TRTC连接信息
- **处理方式**: 收到通知后显示来电界面，提示用户接听或拒绝

- **主题**: `mqtt_call/controller/resident`
- **QoS**: 1
- **用途**: 接收通话控制指令
- **处理方式**:
  - 收到`hangup`时，结束通话
  - 收到其他用户的`answered`时，更新UI状态（如显示"通话已在其他设备接听"）

#### 系统消息订阅
- **主题**: `mqtt_call/system`
- **QoS**: 1
- **用途**: 接收系统级通知，如维护信息、紧急通知等
- **处理方式**: 根据消息level显示不同级别的通知

### 消息示例流程

**住户端消息流**:
```
# 1. 接收来电通知
Topic: mqtt_call/incoming
QoS: 1
Payload: 
{
    "call_id": "d325ffd6-0d42-4516-8aa7-a45ff080d10b",
    "device_device_id": "1",
    "target_resident_id": "1",
    "timestamp": 1746870072136,
    "tencen_rtc": {
        "room_id_type": "string",
        "room_id": "room_1_1_1746870072",
        "sdk_app_id": 1600084384,
        "user_id": "1",
        "user_sig": "eAEAowBc-3siVExTLnZlciI6IjIuMCIsIlRMUy5pZGVudGlmaWVyIjoiMSIsIlRMUy5zZGthcHBpZCI6MTYwMDA4NDM4NCwiVExTLmV4cGlyZSI6ODY0MDAsIlRMUy50aW1lIjoxNzQ2ODcwMDcyLCJUTFMuc2lnIjoieDcwTHQ2ZmxxWkZSendoLzJDMlpwakt4UXQvTkRtcDV5eFlTRXZkYlBGRT0ifQoBAAD--892L3E_"
    }
}
时间: 2025-05-10 17:41:12:704

# 2. 接收住户响应确认（用户已接听）
Topic: mqtt_call/controller/resident
QoS: 1
Payload:
{
    "action": "answered",
    "call_id": "d325ffd6-0d42-4516-8aa7-a45ff080d10b",
    "timestamp": 1746870108846,
    "reason": "user_busy"
}
时间: 2025-05-10 17:41:49:408

# 3. 接收通话结束通知
Topic: mqtt_call/controller/resident
QoS: 1
Payload:
{
    "action": "hangup",
    "call_id": "d325ffd6-0d42-4516-8aa7-a45ff080d10b",
    "timestamp": 1746870238164,
    "reason": "call_completed"
}
时间: 2025-05-10 17:43:58:760
```

**设备端消息流**:
```
# 1. 接收振铃通知
Topic: mqtt_call/controller/device
QoS: 1
Payload:
{
    "action": "ringing",
    "call_id": "d325ffd6-0d42-4516-8aa7-a45ff080d10b",
    "timestamp": 1746870072167
}
时间: 2025-05-10 17:41:12:728

# 2. 接收住户接听通知
Topic: mqtt_call/controller/device
QoS: 1
Payload:
{
    "action": "answered",
    "call_id": "d325ffd6-0d42-4516-8aa7-a45ff080d10b",
    "timestamp": 1746870108815,
    "reason": "user_busy"
}
时间: 2025-05-10 17:41:49:377

# 3. 接收通话结束通知
Topic: mqtt_call/controller/device
QoS: 1
Payload:
{
    "action": "hangup",
    "call_id": "d325ffd6-0d42-4516-8aa7-a45ff080d10b",
    "timestamp": 1746870238164,
    "reason": "call_completed"
}
时间: 2025-05-10 17:43:58:720
```

**系统消息示例**:
```
# 系统维护通知
Topic: mqtt_call/system
QoS: 1
Payload:
{
    "type": "maintenance",
    "level": "info",
    "message": "系统将于今晚22:00进行升级维护，预计持续30分钟",
    "data": {
        "start_time": 1746870072136,
        "duration": 1800000
    },
    "timestamp": 1746870072136
}
```

## 四、HTTP API接口

### 1. 发起通话
- **路径**: `/api/mqtt/call`
- **方法**: POST
- **请求体**:
  ```json
  {
      "device_device_id": "string",
      "household_number": "string",
      "timestamp": 1746870072136
  }
  ```
- **响应**:
  ```json
  {
      "code": 200,
      "message": "成功",
      "data": {
          "call_id": "string",
          "device_device_id": "string",
          "target_resident_ids": ["string"],
          "timestamp": 1746870072136,
          "tencen_rtc": {
              "room_id_type": "string",
              "room_id": "string",
              "sdk_app_id": 1600084384,
              "user_id": "string",
              "user_sig": "string"
          },
          "call_info": {
              "action": "ringing",
              "call_id": "string",
              "timestamp": 1746870072136
          }
      }
  }
  ```

### 2. 设备端通话控制
- **路径**: `/api/mqtt/controller/device`
- **方法**: POST
- **请求体**:
  ```json
  {
      "action": "string",  // hangup/cancelled
      "call_id": "string",
      "timestamp": 1746870072136,
      "reason": "string"
  }
  ```

### 3. 住户端通话控制
- **路径**: `/api/mqtt/controller/resident`
- **方法**: POST
- **请求体**:
  ```json
  {
      "action": "string",  // answered/rejected/hangup/timeout
      "call_id": "string",
      "timestamp": 1746870072136,
      "reason": "string"
  }
  ```

### 4. 获取通话会话
- **路径**: `/api/mqtt/session`
- **方法**: GET
- **参数**: `call_id`
- **响应**:
  ```json
  {
      "code": 200,
      "message": "成功",
      "data": {
          "call_id": "string",
          "status": "string",
          "device_id": "string",
          "resident_id": "string",
          "start_time": 1746870072136,
          "end_time": 1746870238164
      }
  }
  ```

### 5. 结束通话会话
- **路径**: `/api/mqtt/end-session`
- **方法**: POST
- **请求体**:
  ```json
  {
      "call_id": "string",
      "reason": "string"
  }
  ```

### 6. 更新设备状态
- **路径**: `/api/mqtt/device/status`
- **方法**: POST
- **请求体**:
  ```json
  {
      "device_id": "string",
      "status": {
          "online": true,
          "timestamp": 1746870072136
      }
  }
  ```

### 7. 发送系统消息
- **路径**: `/api/mqtt/system/message`
- **方法**: POST
- **请求体**:
  ```json
  {
      "type": "string",
      "level": "string",
      "message": "string",
      "target": ["string"]
  }
  ```

## 五、消息动作说明

### 1. 设备端动作
- `ringing`: 呼叫振铃中
- `hangup`: 挂断通话
- `cancelled`: 取消呼叫

### 2. 住户端动作
- `answered`: 接听通话
- `rejected`: 拒绝通话
- `hangup`: 挂断通话
- `timeout`: 呼叫超时

## 六、错误处理

所有API响应均使用统一的错误响应格式：
```json
{
    "code": 400,
    "message": "错误描述",
    "data": null
}
```

常见错误码：
- 400: 请求参数错误
- 404: 资源不存在
- 500: 服务器内部错误