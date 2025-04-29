# iLock 智能门禁系统

## 项目概述

iLock是一个基于Go语言开发的智能门禁管理系统，提供了强大的门禁控制、视频通话和紧急情况处理功能。系统采用Docker容器化部署，便于在各种环境中快速安装和更新。

## 系统架构

- **后端**: Go + Gin框架 + GORM
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.4.1
- **部署**: Docker + Docker Compose
- **通讯**: 
  - RESTful API: 基础业务操作
  - MQTT: 实时消息推送、视频通话信令
  - TRTC: 腾讯云实时音视频

## MQTT通信架构

### 1. 主题设计

#### 视频通话相关主题
- **呼叫请求**: `calls/request/{caller_device_id}`
- **来电通知**: `users/{user_id}/calls/incoming`
- **呼叫方控制**: `devices/{caller_device_id}/calls/control`
- **接收方控制**: `users/{user_id}/calls/control`

### 2. 消息质量(QoS)
- 视频通话信令: QoS 1 (至少一次送达)
- 普通通知: QoS 0 (最多一次送达)
- 紧急通知: QoS 2 (确保一次送达)

### 3. 消息格式
所有消息采用JSON格式，包含以下基本字段：
- `message_id`: 消息唯一标识
- `timestamp`: 消息时间戳
- `type`: 消息类型
- `payload`: 消息内容

### 4. 实时通信流程

#### 视频通话流程
1. 访客通过门禁设备发起呼叫
2. 后端接收呼叫请求并创建TRTC房间
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
   - Linux服务器（推荐Ubuntu 20.04或CentOS 8）
   - Docker和Docker Compose已安装
   - 开放端口：8080(HTTP), 3310(MySQL), 6380(Redis)

2. **本地环境**（用于部署）:
   - Windows操作系统
   - 已安装PuTTY工具集（包含pscp.exe和plink.exe）

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
   - 访问`http://服务器IP:20033/swagger/index.html`查看API文档

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
  - SSH端口: 22

- **数据库**:
  - 主机: localhost
  - 端口: 3309
  - 用户: root
  - 密码: 
  - 数据库名: ilock_db

- **Docker镜像**:
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

## API文档

系统集成了Swagger文档，部署后可以通过以下地址访问：

http://服务器IP:20033/swagger/index.html

主要API端点包括：

- **认证**: `/api/auth/login`
- **管理员**: `/api/admins/*`
- **物业人员**: `/api/staffs/*`
- **居民**: `/api/residents/*`
- **设备**: `/api/devices/*`
- **通话记录**: `/api/calls/*`
- **紧急情况**: `/api/emergency/*`


## MQTT通讯协议接口文档

### 通讯架构

iLock系统使用MQTT协议实现实时消息推送和视频通话信令传输。下面详细说明MQTT通讯协议的各个方面：

### 1. MQTT主题设计

#### 视频通话相关主题
- **呼叫请求** (Caller -> Backend): `calls/request/{caller_device_id}`
  - `{caller_device_id}`: 呼叫方设备的唯一标识符
- **来电通知** (Backend -> Callee): `users/{user_id}/calls/incoming`
  - `{user_id}`: 被呼叫用户的唯一标识符，消息发送给该用户所有在线的MQTT客户端
- **通话控制** (Backend -> Caller): `devices/{caller_device_id}/calls/control`
  - `{caller_device_id}`: 接收控制指令的呼叫方设备
- **通话控制** (Backend -> Callee): `users/{user_id}/calls/control`
  - `{user_id}`: 接收控制指令的目标用户，消息发送给该用户所有在线的MQTT客户端
- **设备状态**: `devices/{device_id}/status`
  - `{device_id}`: 设备ID，用于设备状态的发布和订阅
- **系统消息**: `system/{message_type}`
  - `{message_type}`: 消息类型，如通知、警告、错误等

### 2. 消息载荷格式

所有消息载荷均使用JSON格式。

#### 呼叫请求 (`calls/request/{caller_device_id}`)
**方向**: Caller -> Backend
```json
{
  "call_id": "unique_call_identifier_generated_by_caller", // 本次呼叫的唯一ID
  "caller_id": "caller_device_id",                        // 呼叫方设备ID
  "target_user_id": "callee_user_id",                     // 目标用户ID
  "timestamp": 1678886400000                              // 发起呼叫的Unix毫秒时间戳
}
```

#### 来电通知 (`users/{user_id}/calls/incoming`)
**方向**: Backend -> Callee
```json
{
  "call_id": "unique_call_identifier_from_request",      // 对应呼叫请求的ID
  "caller_id": "caller_device_id",                       // 呼叫方设备ID
  "caller_info": {                                        // 呼叫方信息
    "name": "设备名称或用户昵称"
  },
  "timestamp": 1678886450000,                            // 发送通知的Unix毫秒时间戳
  "room_info": {                                          // 加入TRTC房间所需信息
    "room_id": "trtc_room_id_created_by_backend",        // TRTC房间号
    "sdk_app_id": 1400000000,                            // TRTC应用ID
    "user_id": "callee_user_id",                         // 被呼叫方在TRTC中使用的UserID
    "user_sig": "generated_user_signature_for_callee"    // 被呼叫方的TRTC签名
  }
}
```

#### 通话控制 - 发送给呼叫方 (`devices/{caller_device_id}/calls/control`)
**方向**: Backend -> Caller
```json
{
  "action": "ringing",                                   // 控制动作类型
  "call_id": "unique_call_identifier_from_request",      // 对应呼叫请求的ID
  "timestamp": 1678886500000,                            // 发送指令的Unix毫秒时间戳
  "reason": "Optional message for details"               // 可选，提供额外信息
}
```

**action说明**:
- `ringing`: Callee正在被呼叫（已发送incoming通知）
- `rejected`: Callee拒绝了通话
- `hangup`: Callee挂断了通话
- `timeout`: Callee无应答超时
- `error`: 处理过程中发生错误（如创建房间失败）

#### 通话控制 - 发送给被呼叫方 (`users/{user_id}/calls/control`)
**方向**: Backend -> Callee
```json
{
  "action": "cancelled",                                 // 控制动作类型
  "call_id": "unique_call_identifier_from_request",      // 对应呼叫请求的ID
  "timestamp": 1678886550000,                            // 发送指令的Unix毫秒时间戳
  "reason": "Optional message for details"               // 可选，提供额外信息
}
```

**action说明**:
- `cancelled`: Caller在Callee接听前取消了呼叫
- `hangup`: Caller挂断了通话

#### 设备状态 (`devices/{device_id}/status`)
```json
{
  "device_id": "device123",                              // 设备ID
  "online": true,                                        // 在线状态
  "battery": 85,                                         // 电池电量（百分比）
  "last_update": 1678886600000,                          // 最后更新时间戳
  "properties": {                                         // 其他属性（可选）
    "temperature": 25.5,
    "firmware_version": "1.2.3",
    "door_status": "closed"
  }
}
```

#### 系统消息 (`system/{message_type}`)
```json
{
  "type": "device_offline",                              // 消息类型
  "level": "warning",                                    // 消息级别: info, warning, error
  "message": "门口设备离线",                              // 消息内容
  "timestamp": 1678886700000,                            // 发送时间戳
  "data": {                                               // 额外数据（可选）
    "device_id": "device123",
    "last_seen": 1682570000000
  }
}
```
### 5. 安全性考虑

- 所有MQTT通信使用TLS加密
- 客户端需要使用用户名/密码或证书进行身份验证
- 主题设计确保信息隔离，防止未授权访问

### 6. 视频通话详细流程

#### 呼叫建立流程
1. **呼叫发起**：
   - 呼叫端(设备)通过MQTT向服务器发布呼叫请求消息，Topic为 `calls/request/{caller_device_id}`
   - 或者呼叫端通过HTTP API调用 `POST /api/mqtt/calls/initiate` 直接发起呼叫
   - 请求参数:
     ```json
     {
       "caller_id": "device123",  // 呼叫方设备ID
       "callee_id": "user456"     // 被呼叫方用户ID
     }
     ```
   - 响应:
     ```json
     {
       "code": 0,
       "message": "成功",
       "data": {
         "call_id": "call_device123_user456_1629123456789"
       }
     }
     ```

2. **后端处理呼叫请求**：
   - 后端服务接收到MQTT呼叫请求或HTTP API请求
   - 创建呼叫会话记录，生成唯一call_id
   - 通过调用腾讯云TRTC API创建视频通话房间
   - 内部实现:
     ```go
     // CallService.InitiateCall 方法实现
     func (s *CallService) InitiateCall(deviceID string, userID string) (string, error) {
       // 生成唯一呼叫ID
       callID := fmt.Sprintf("call_%s_%s_%d", deviceID, userID, time.Now().UnixMilli())
       
       // 创建通话会话记录
       session := s.GetOrCreateCallSession(callID, deviceID, userID)
       
       // 创建TRTC房间
       roomID, err := s.RTCService.CreateVideoCall(deviceID, userID)
       if err != nil {
         return "", fmt.Errorf("创建视频通话失败: %v", err)
       }
       
       // 保存房间ID
       session.RoomID = roomID
       
       // 继续处理...
     }
     ```

3. **获取TRTC凭证**：
   - 后端服务为被呼叫方生成TRTC所需的RoomID、UserSig等凭证信息
   - 将凭证和房间信息打包为来电通知
   - 内部实现:
     ```go
     // 为被呼叫方生成UserSig
     userSigInfo, err := s.RTCService.GetUserSig(userID)
     if err != nil {
       return "", fmt.Errorf("为被呼叫方生成UserSig失败: %v", err)
     }
     
     // 组装通知内容
     notification := &CallIncomingNotification{
       CallID:    callID,
       CallerID:  deviceID,
       Timestamp: time.Now().UnixMilli(),
       CallerInfo: CallerInfo{
         Name: "门口设备", // 实际项目中从数据库获取设备名称
       },
       RoomInfo: RoomInfo{
         RoomID:   roomID,
         SDKAppID: userSigInfo.SDKAppID,
         UserID:   userID,
         UserSig:  userSigInfo.UserSig,
       },
     }
     ```

4. **推送来电通知**：
   - 后端服务通过MQTT向被呼叫用户发布来电通知，Topic为 `users/{user_id}/calls/incoming`
   - 通知内容包含来电方信息和TRTC房间信息(RoomID, SDKAppID, UserID, UserSig)
   - MQTT消息格式:
     ```json
     {
       "call_id": "call_device123_user456_1629123456789",
       "caller_id": "device123",
       "caller_info": {
         "name": "门口设备"
       },
       "timestamp": 1678886450000,
       "room_info": {
         "room_id": "664321",
         "sdk_app_id": 1400000001,
         "user_id": "user456",
         "user_sig": "eJwtzM9qwkAUBeCvIrNtYZL5l5..."
       }
     }
     ```
   - 内部实现:
     ```go
     // 发送来电通知
     if err := s.MQTTService.SendIncomingCallNotification(userID, notification); err != nil {
       return "", fmt.Errorf("发送来电通知失败: %v", err)
     }
     ```

5. **发送响铃状态**：
   - 后端服务通过MQTT向呼叫方发送"正在响铃"状态，Topic为 `devices/{caller_device_id}/calls/control`
   - 呼叫方设备收到后显示"呼叫中"状态
   - MQTT消息格式:
     ```json
     {
       "action": "ringing",
       "call_id": "call_device123_user456_1629123456789",
       "timestamp": 1678886500000
     }
     ```
   - 内部实现:
     ```go
     // 发送响铃状态
     controlMsg := &CallControl{
       Action:    "ringing",
       CallID:    callID,
       Timestamp: time.Now().UnixMilli(),
     }
     
     if err := s.MQTTService.SendCallerControlMessage(deviceID, controlMsg); err != nil {
       return "", fmt.Errorf("发送响铃状态失败: %v", err)
     }
     ```

6. **用户接听/拒绝**：
   - 用户端(手机)从MQTT消息中解析房间信息
   - 用户决定接听或拒绝，通过HTTP API调用 `POST /api/mqtt/calls/callee-action` 通知后端
   - 请求参数:
     ```json
     {
       "call_id": "call_device123_user456_1629123456789",
       "action": "accepted",  // 可选值: accepted(接听), rejected(拒绝), timeout(超时)
       "reason": "用户手动接听"  // 可选，说明原因
     }
     ```
   - 响应:
     ```json
     {
       "code": 0,
       "message": "成功",
       "data": null
     }
     ```
   - 移动端实现示例(Android):
     ```java
     // 用户接听通话
     public void acceptCall(String callId) {
       // 停止铃声
       ringtonePlayer.stop();
       
       // 通知后端用户已接听
       apiService.calleeAction(
         new CalleeActionRequest(callId, "accepted", "用户接听")
       ).enqueue(new Callback<ApiResponse>() {
         @Override
         public void onResponse(Call<ApiResponse> call, Response<ApiResponse> response) {
           if (response.isSuccessful()) {
             // 使用TRTC SDK加入房间
             joinTRTCRoom(roomInfo);
           }
         }
       });
     }
     
     // 用户拒绝通话
     public void rejectCall(String callId) {
       // 停止铃声
       ringtonePlayer.stop();
       
       // 通知后端用户拒绝
       apiService.calleeAction(
         new CalleeActionRequest(callId, "rejected", "用户拒绝")
       ).enqueue(new Callback<ApiResponse> {
         @Override
         public void onResponse(Call<ApiResponse> call, Response<ApiResponse> response) {
           // 关闭来电界面
           finish();
         }
       });
     }
     ```

#### 通话连接流程
7. **建立视频连接**：
   - 如果用户接听，用户端使用收到的TRTC房间信息建立P2P视频连接
   - 后端通过MQTT向呼叫方发送用户已接听的通知 (`devices/{caller_device_id}/calls/control`)
   - MQTT消息格式:
     ```json
     {
       "action": "accepted",
       "call_id": "call_device123_user456_1629123456789",
       "timestamp": 1678886600000
     }
     ```
   - 内部实现:
     ```go
     // HandleCalleeAction 处理被呼叫方动作
     func (s *CallService) HandleCalleeAction(callID string, action string, reason string) error {
       session := s.GetCallSession(callID)
       if session == nil {
         return fmt.Errorf("通话不存在")
       }
       
       // 发送控制消息给呼叫方
       controlMsg := &CallControl{
         Action:    CallAction(action),
         CallID:    callID,
         Timestamp: time.Now().UnixMilli(),
         Reason:    reason,
       }
       
       if err := s.MQTTService.SendCallerControlMessage(session.CallerID, controlMsg); err != nil {
         return fmt.Errorf("发送控制消息失败: %v", err)
       }
       
       // 如果是拒绝或超时，结束通话
       if action == "rejected" || action == "timeout" {
         s.EndCallSession(callID)
       }
       
       return nil
     }
     ```
   - 移动端实现示例(Android):
     ```java
     // 加入TRTC房间
     private void joinTRTCRoom(JSONObject roomInfo) {
       try {
         // 提取房间信息
         String roomId = roomInfo.getString("room_id");
         int sdkAppId = roomInfo.getInt("sdk_app_id");
         String userId = roomInfo.getString("user_id");
         String userSig = roomInfo.getString("user_sig");
         
         // 初始化TRTC SDK
         trtcCloud = TRTCCloud.sharedInstance(getApplicationContext());
         
         // 设置参数
         TRTCParams trtcParams = new TRTCParams();
         trtcParams.sdkAppId = sdkAppId;
         trtcParams.userId = userId;
         trtcParams.userSig = userSig;
         trtcParams.roomId = Integer.parseInt(roomId);
         
         // 设置视频参数
         TRTCVideoEncParam encParam = new TRTCVideoEncParam();
         encParam.videoResolution = TRTCCloudDef.TRTC_VIDEO_RESOLUTION_640_360;
         encParam.videoFps = 15;
         encParam.videoBitrate = 600;
         trtcCloud.setVideoEncoderParam(encParam);
         
         // 开启本地预览
         trtcCloud.startLocalPreview(true, localVideoView);
         
         // 开启本地音频
         trtcCloud.startLocalAudio(TRTCCloudDef.TRTC_AUDIO_QUALITY_SPEECH);
         
         // 加入房间
         trtcCloud.enterRoom(trtcParams, TRTCCloudDef.TRTC_APP_SCENE_VIDEOCALL);
       } catch (Exception e) {
         Log.e(TAG, "加入TRTC房间失败", e);
       }
     }
     ```

8. **视频通话进行**：
   - 双方通过TRTC直接建立P2P连接，视频通话开始
   - 后端更新通话记录状态为"进行中"
   - 内部实现:
     ```go
     // 当用户接听后，更新通话状态
     if action == "accepted" {
       session.CallState = "connected"
     }
     ```

#### 通话结束流程
9. **结束通话**：
   - 任一方可通过HTTP API通知后端结束通话
   - 呼叫方通过 `POST /api/mqtt/calls/caller-action` 发送挂断请求
   - 被呼叫方通过 `POST /api/mqtt/calls/callee-action` 发送挂断请求
   - 请求参数(呼叫方):
     ```json
     {
       "call_id": "call_device123_user456_1629123456789",
       "action": "hangup",
       "reason": "用户主动挂断"
     }
     ```
   - 响应:
     ```json
     {
       "code": 0,
       "message": "成功",
       "data": null
     }
     ```
   - 移动端实现示例(Android):
     ```java
     // 挂断通话
     public void hangupCall(String callId) {
       // 通知后端挂断
       apiService.calleeAction(
         new CalleeActionRequest(callId, "hangup", "用户挂断")
       ).enqueue(new Callback<ApiResponse>() {
         @Override
         public void onResponse(Call<ApiResponse> call, Response<ApiResponse> response) {
           // 退出TRTC房间
           exitTRTCRoom();
           // 关闭通话界面
           finish();
         }
       });
     }
     
     // 退出TRTC房间
     private void exitTRTCRoom() {
       if (trtcCloud != null) {
         // 停止本地预览
         trtcCloud.stopLocalPreview();
         // 停止本地音频
         trtcCloud.stopLocalAudio();
         // 退出房间
         trtcCloud.exitRoom();
       }
     }
     ```

10. **发送结束通知**：
    - 后端服务接收到挂断请求后，通过MQTT向对方发送挂断通知
    - 呼叫方收到通知的Topic为 `devices/{caller_device_id}/calls/control`
    - 被呼叫方收到通知的Topic为 `users/{user_id}/calls/control`
    - MQTT消息格式(发送给被呼叫方):
      ```json
      {
        "action": "hangup",
        "call_id": "call_device123_user456_1629123456789",
        "timestamp": 1678886800000,
        "reason": "呼叫方挂断"
      }
      ```
    - 内部实现:
      ```go
      // HandleCallerAction 处理呼叫方动作
      func (s *CallService) HandleCallerAction(callID string, action string, reason string) error {
        session := s.GetCallSession(callID)
        if session == nil {
          return fmt.Errorf("通话不存在")
        }
        
        // 发送控制消息给被呼叫方
        controlMsg := &CallControl{
          Action:    CallAction(action),
          CallID:    callID,
          Timestamp: time.Now().UnixMilli(),
          Reason:    reason,
        }
        
        if err := s.MQTTService.SendCalleeControlMessage(session.CalleeID, controlMsg); err != nil {
          return fmt.Errorf("发送控制消息失败: %v", err)
        }
        
        // 结束通话
        s.EndCallSession(callID)
        
        return nil
      }
      ```

11. **释放资源**：
    - 后端关闭TRTC房间
    - 更新通话记录状态为"已结束"
    - 记录通话时长等信息
    - 内部实现:
      ```go
      // EndCallSession 结束通话会话
      func (s *CallService) EndCallSession(callID string) {
        s.callsMutex.Lock()
        defer s.callsMutex.Unlock()
        
        if session, exists := s.ActiveCalls[callID]; exists {
          session.CallState = "ended"
          session.EndTimestamp = time.Now().UnixMilli()
          
          // 计算通话时长
          duration := session.EndTimestamp - session.StartTimestamp
          
          // 保存通话记录到数据库
          callRecord := &models.CallRecord{
            CallID:     session.CallID,
            CallerID:   session.CallerID,
            CalleeID:   session.CalleeID,
            StartTime:  time.UnixMilli(session.StartTimestamp),
            EndTime:    time.UnixMilli(session.EndTimestamp),
            Duration:   duration / 1000, // 转换为秒
            CallResult: session.CallState,
          }
          
          s.DB.Create(callRecord)
          
          // 从活动通话中移除
          delete(s.ActiveCalls, callID)
        }
      }
      ```

#### 异常处理
12. **超时处理**：
    - 如果被呼叫方在规定时间内未响应，后端自动发送"超时"通知给呼叫方
    - 通过MQTT Topic `devices/{caller_device_id}/calls/control` 发送，action为"timeout"
    - MQTT消息格式:
      ```json
      {
        "action": "timeout",
        "call_id": "call_device123_user456_1629123456789",
        "timestamp": 1678886650000,
        "reason": "呼叫超时无人接听"
      }
      ```
    - 内部实现:
      ```go
      // 呼叫超时检查
      func (s *CallService) checkCallTimeout() {
        s.callsMutex.RLock()
        defer s.callsMutex.RUnlock()
        
        now := time.Now().UnixMilli()
        timeout := int64(30 * 1000) // 30秒超时
        
        for callID, session := range s.ActiveCalls {
          // 检查是否处于响铃状态且已超时
          if session.CallState == "ringing" && (now - session.StartTimestamp > timeout) {
            // 发送超时通知给呼叫方
            controlMsg := &CallControl{
              Action:    "timeout",
              CallID:    callID,
              Timestamp: now,
              Reason:    "呼叫超时无人接听",
            }
            
            s.MQTTService.SendCallerControlMessage(session.CallerID, controlMsg)
            
            // 结束通话
            go s.EndCallSession(callID)
          }
        }
      }
      ```

13. **连接错误处理**：
    - 如果TRTC房间创建失败或其他错误，后端发送错误通知给呼叫方
    - 通过MQTT Topic `devices/{caller_device_id}/calls/control` 发送，action为"error"
    - MQTT消息格式:
      ```json
      {
        "action": "error",
        "call_id": "call_device123_user456_1629123456789",
        "timestamp": 1678886550000,
        "reason": "创建TRTC房间失败"
      }
      ```
    - 内部实现:
      ```go
      // 处理创建房间失败
      if err != nil {
        // 发送错误通知给呼叫方
        errorMsg := &CallControl{
          Action:    "error",
          CallID:    callID,
          Timestamp: time.Now().UnixMilli(),
          Reason:    fmt.Sprintf("创建通话失败: %v", err),
        }
        
        s.MQTTService.SendCallerControlMessage(deviceID, errorMsg)
        
        // 结束通话会话
        s.EndCallSession(callID)
        
        return "", err
      }
      ```

14. **网络异常处理**：
    - 移动端在网络波动时自动重连MQTT服务
    - 视频通话中网络质量监控与切换
    - 移动端实现示例(Android):
      ```java
      // MQTT连接监听
      mqttClient.setCallback(new MqttCallback() {
        @Override
        public void connectionLost(Throwable cause) {
          Log.e(TAG, "MQTT连接断开", cause);
          
          // 网络连接丢失，尝试重连
          reconnectMQTT();
        }
        
        // 其他回调方法...
      });
      
      // TRTC网络质量监听
      @Override
      public void onNetworkQuality(TRTCCloudDef.TRTCQuality localQuality, 
                                   ArrayList<TRTCCloudDef.TRTCQuality> remoteQuality) {
        // 本地网络质量变化
        updateNetworkQualityUI(localQuality.quality);
        
        // 网络质量较差时降低视频分辨率
        if (localQuality.quality > TRTCCloudDef.TRTC_QUALITY_Poor) {
          TRTCVideoEncParam encParam = new TRTCVideoEncParam();
          encParam.videoResolution = TRTCCloudDef.TRTC_VIDEO_RESOLUTION_480_270;
          encParam.videoFps = 15;
          encParam.videoBitrate = 400;
          trtcCloud.setVideoEncoderParam(encParam);
        }
      }
      ```

#### 移动端MQTT接入流程

15. **移动端MQTT客户端初始化**：
    - 用户登录后，自动初始化MQTT客户端并连接服务器
    - 配置自动重连机制与会话持久化
    - 实现示例(Android):
      ```java
      // 在登录成功后初始化MQTT
      private void initMQTT(String userId, String jwtToken) {
        String clientId = userId + "_android_" + UUID.randomUUID().toString();
        
        // 保存客户端ID用于后续通信
        PreferenceManager.getDefaultSharedPreferences(this)
            .edit()
            .putString("mqtt_client_id", clientId)
            .apply();
        
        // 配置MQTT选项
        MqttConnectOptions options = new MqttConnectOptions();
        options.setCleanSession(false);  // 保持会话
        options.setAutomaticReconnect(true);  // 自动重连
        options.setKeepAliveInterval(60);  // 心跳间隔
        options.setConnectionTimeout(30);  // 连接超时
        
        // 使用JWT令牌作为密码
        options.setUserName(userId);
        options.setPassword(jwtToken.toCharArray());
        
        // 初始化MQTT客户端
        mqttClient = new MqttAndroidClient(
            getApplicationContext(),
            "tcp://mqtt.ilock.com:1883",
            clientId
        );
        
        // 设置回调
        mqttClient.setCallback(createMqttCallback());
        
        // 连接服务器
        try {
          mqttClient.connect(options, null, new IMqttActionListener() {
            @Override
            public void onSuccess(IMqttToken asyncActionToken) {
              Log.d(TAG, "MQTT连接成功");
              // 订阅相关主题
              subscribeToTopics();
              // 注册设备
              registerDevice(userId, clientId);
            }
            
            @Override
            public void onFailure(IMqttToken asyncActionToken, Throwable exception) {
              Log.e(TAG, "MQTT连接失败", exception);
            }
          });
        } catch (MqttException e) {
          Log.e(TAG, "MQTT连接异常", e);
        }
      }
      ```

16. **注册设备接收推送**：
    - 移动端启动时通过HTTP API注册设备
    - API地址: `POST /api/mqtt/clients/register`
    - 请求参数:
      ```json
      {
        "user_id": "user456",
        "device_type": "mobile",
        "client_id": "user456_android_uuid1234",
        "push_token": "firebase_token123",
        "platform": "android",
        "app_version": "1.2.0"
      }
      ```
    - 响应:
      ```json
      {
        "code": 0,
        "message": "成功",
        "data": {
          "mqtt_credentials": {
            "broker_url": "tcp://mqtt.ilock.com:1883",
            "username": "user456",
            "password": "jwt_token_for_mqtt",
            "client_id": "user456_android_uuid1234"
          },
          "topics": {
            "incoming_calls": "users/user456/calls/incoming",
            "call_control": "users/user456/calls/control"
          }
        }
      }
      ```
    - 实现示例(Android):
      ```java
      // 注册设备接收推送
      private void registerDevice(String userId, String clientId) {
        // 获取FCM推送令牌
        FirebaseMessaging.getInstance().getToken()
            .addOnCompleteListener(task -> {
              if (task.isSuccessful()) {
                String pushToken = task.getResult();
                
                // 构建请求数据
                JSONObject requestData = new JSONObject();
                try {
                  requestData.put("user_id", userId);
                  requestData.put("device_type", "mobile");
                  requestData.put("client_id", clientId);
                  requestData.put("push_token", pushToken);
                  requestData.put("platform", "android");
                  requestData.put("app_version", BuildConfig.VERSION_NAME);
                } catch (JSONException e) {
                  Log.e(TAG, "构建注册数据失败", e);
                  return;
                }
                
                // 发送注册请求
                apiService.registerDevice(requestData)
                    .enqueue(new Callback<ApiResponse>() {
                      @Override
                      public void onResponse(Call<ApiResponse> call, Response<ApiResponse> response) {
                        if (response.isSuccessful()) {
                          Log.d(TAG, "设备注册成功");
                        } else {
                          Log.e(TAG, "设备注册失败: " + response.message());
                        }
                      }
                      
                      @Override
                      public void onFailure(Call<ApiResponse> call, Throwable t) {
                        Log.e(TAG, "设备注册请求失败", t);
                      }
                    });
              }
            });
      }
      ```

## 系统特性

- **自动备份与回滚**: 在更新前自动创建备份，更新失败时自动回滚
- **基于角色的访问控制**: 系统管理员、物业人员和居民具有不同的权限
- **安全通信**: 基于JWT的API认证
- **视频通话集成**: 集成阿里云RTC提供实时视频通话
- **容器化部署**: 使用Docker和Docker Compose简化部署和维护
- **紧急响应系统**: 快速处理火灾、入侵等紧急情况
- **健康检查**: 服务健康状态监控，确保系统稳定运行

## 故障排除

1. **服务无法启动**:
   - 检查Docker和Docker Compose是否正确安装
   - 检查端口是否被占用: `netstat -tunlp`
   - 查看容器日志: `docker-compose logs backend`

2. **数据库连接失败**:
   - 检查数据库配置是否正确
   - 确认数据库服务是否运行: `docker-compose ps mysql`
   - 尝试手动连接数据库验证凭据

3. **API响应错误**:
   - 检查JWT密钥配置
   - 确认请求格式是否正确
   - 查看服务日志了解详细错误信息

4. **部署脚本执行失败**:
   - 确保PuTTY工具集已正确安装并在PATH中
   - 检查服务器连接信息是否正确
   - 验证本地文件存在且权限正确

## 许可证

版权所有 © 2024 iLock开发团队