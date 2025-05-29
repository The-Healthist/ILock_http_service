# iLock MQTT推送通知功能设计

## 一、视频通话流程图解析

根据流程图显示，视频通话完整流程如下：

1. **呼叫发起**
   - 呼叫端(设备)通过HTTP请求创建通话会话
   - 后端接收请求，生成通话ID

2. **通话房间创建**
   - 后端调用腾讯云TRTC API创建视频房间
   - 获取房间ID和UserSig等凭证信息

3. **来电通知推送**
   - 后端通过MQTT向被呼叫用户推送来电通知
   - 主题: `users/{user_id}/calls/incoming`
   - 通知内容包含通话信息和TRTC房间参数

4. **响铃状态发送**
   - 后端向呼叫方发送"正在响铃"状态
   - 主题: `devices/{caller_device_id}/calls/control`

5. **用户接听/拒绝**
   - 手机端接收到MQTT通知，显示来电界面
   - 用户做出选择后，通过HTTP API通知后端
   - 后端转发通知给呼叫方

6. **通话建立/结束**
   - 如果接听，双方使用TRTC参数建立连接
   - 任意方结束通话时，通过HTTP API通知后端
   - 后端通过MQTT通知对方

## 二、移动端MQTT设计

### 1. 移动端MQTT连接管理

为确保手机端能可靠接收来电通知，需要设计以下功能：

#### 1.1 MQTT连接与认证
- 用户登录后，自动连接MQTT服务器
- 使用JWT令牌作为MQTT认证凭证
- 实现自动重连和连接状态管理

```java
// Android示例：MQTT连接管理
private void connectToMQTT() {
    // 1. 创建MQTT客户端ID (使用用户ID+设备标识+随机字符串)
    String clientId = userId + "_android_" + UUID.randomUUID().toString();
    
    // 2. 初始化MQTT客户端
    mqttClient = new MqttAsyncClient("tcp://mqtt.ilock.com:1883", clientId, new MemoryPersistence());
    
    // 3. 设置连接选项
    MqttConnectOptions options = new MqttConnectOptions();
    options.setCleanSession(false);
    options.setAutomaticReconnect(true);
    options.setConnectionTimeout(10);
    options.setKeepAliveInterval(60);
    
    // 4. 使用JWT作为认证凭证
    options.setUserName(userId);
    options.setPassword(jwtToken.toCharArray());
    
    // 5. 连接MQTT服务器
    try {
        mqttClient.connect(options, null, new IMqttActionListener() {
            @Override
            public void onSuccess(IMqttToken asyncActionToken) {
                Log.d(TAG, "MQTT连接成功");
                subscribeToTopics();
            }
            
            @Override
            public void onFailure(IMqttToken asyncActionToken, Throwable exception) {
                Log.e(TAG, "MQTT连接失败", exception);
                // 可实现重试逻辑
            }
        });
    } catch (MqttException e) {
        Log.e(TAG, "MQTT连接异常", e);
    }
}
```

#### 1.2 主题订阅
- 连接成功后，订阅个人消息主题
- 为确保消息可靠送达，使用QoS 1

```java
// Android示例：订阅主题
private void subscribeToTopics() {
    try {
        // 1. 订阅来电通知主题
        mqttClient.subscribe("users/" + userId + "/calls/incoming", 1);
        
        // 2. 订阅通话控制主题
        mqttClient.subscribe("users/" + userId + "/calls/control", 1);
        
        Log.d(TAG, "成功订阅MQTT主题");
    } catch (MqttException e) {
        Log.e(TAG, "订阅MQTT主题失败", e);
    }
}
```

#### 1.3 消息回调处理
- 实现MQTT消息回调接口
- 根据主题类型，分发处理不同消息

```java
// Android示例：MQTT消息回调
mqttClient.setCallback(new MqttCallback() {
    @Override
    public void messageArrived(String topic, MqttMessage message) {
        Log.d(TAG, "收到MQTT消息: " + topic);
        
        try {
            String payload = new String(message.getPayload(), StandardCharsets.UTF_8);
            
            // 来电通知
            if (topic.endsWith("/calls/incoming")) {
                JSONObject callInfo = new JSONObject(payload);
                handleIncomingCall(callInfo);
            }
            // 通话控制消息
            else if (topic.endsWith("/calls/control")) {
                JSONObject controlInfo = new JSONObject(payload);
                handleCallControl(controlInfo);
            }
        } catch (Exception e) {
            Log.e(TAG, "处理MQTT消息失败", e);
        }
    }
    
    @Override
    public void connectionLost(Throwable cause) {
        Log.e(TAG, "MQTT连接断开", cause);
        // 实现重连逻辑
    }
    
    @Override
    public void deliveryComplete(IMqttDeliveryToken token) {
        // 消息发送完成的回调
    }
});
```

### 2. 来电处理流程

#### 2.1 接收来电通知
- 接收MQTT来电通知消息
- 解析房间信息与通话参数
- 显示来电界面

```java
// Android示例：处理来电通知
private void handleIncomingCall(JSONObject callInfo) {
    try {
        // 1. 提取关键信息
        String callId = callInfo.getString("call_id");
        String callerId = callInfo.getString("caller_id");
        JSONObject callerInfo = callInfo.getJSONObject("caller_info");
        JSONObject roomInfo = callInfo.getJSONObject("room_info");
        
        // 2. 保存TRTC房间信息(用于接听时加入房间)
        saveTRTCRoomInfo(roomInfo);
        
        // 3. 显示来电界面
        showIncomingCallUI(callId, callerId, callerInfo.getString("name"));
        
        // 4. 播放铃声
        playRingtone();
        
    } catch (JSONException e) {
        Log.e(TAG, "解析来电信息失败", e);
    }
}
```

#### 2.2 用户响应处理
- 用户接听/拒绝通话后通知后端
- 调用HTTP API发送响应

```java
// Android示例：用户接听通话
public void acceptCall(String callId) {
    // 停止铃声
    stopRingtone();
    
    // 通知后端用户已接听
    ApiService apiService = RetrofitClient.getApiService();
    apiService.calleeAction(
        new CalleeActionRequest(callId, "accepted", null)
    ).enqueue(new Callback<ApiResponse>() {
        @Override
        public void onResponse(Call<ApiResponse> call, Response<ApiResponse> response) {
            if (response.isSuccessful()) {
                // 加入TRTC房间开始通话
                joinTRTCRoom();
            } else {
                // 处理错误
                showError("接听失败: " + response.message());
            }
        }
        
        @Override
        public void onFailure(Call<ApiResponse> call, Throwable t) {
            showError("网络错误: " + t.getMessage());
        }
    });
}

// Android示例：用户拒绝通话
public void rejectCall(String callId) {
    // 停止铃声
    stopRingtone();
    
    // 通知后端用户拒绝接听
    ApiService apiService = RetrofitClient.getApiService();
    apiService.calleeAction(
        new CalleeActionRequest(callId, "rejected", "用户拒绝接听")
    ).enqueue(new Callback<ApiResponse>() {
        @Override
        public void onResponse(Call<ApiResponse> call, Response<ApiResponse> response) {
            // 关闭来电界面
            dismissIncomingCallUI();
        }
        
        @Override
        public void onFailure(Call<ApiResponse> call, Throwable t) {
            showError("网络错误: " + t.getMessage());
            dismissIncomingCallUI();
        }
    });
}
```

#### 2.3 处理通话控制消息
- 接收MQTT通话控制消息
- 根据控制类型更新通话状态

```java
// Android示例：处理通话控制消息
private void handleCallControl(JSONObject controlInfo) {
    try {
        String action = controlInfo.getString("action");
        String callId = controlInfo.getString("call_id");
        
        switch (action) {
            case "cancelled":
                // 对方取消了通话
                stopRingtone();
                showToast("对方取消了通话");
                dismissIncomingCallUI();
                break;
                
            case "hangup":
                // 对方挂断了通话
                showToast("对方挂断了通话");
                exitTRTCRoom();
                finishCallActivity();
                break;
                
            // 其他控制类型...
        }
    } catch (JSONException e) {
        Log.e(TAG, "解析通话控制消息失败", e);
    }
}
```

### 3. TRTC房间集成

#### 3.1 加入视频通话房间
- 使用MQTT消息中提供的TRTC参数
- 初始化TRTC SDK并加入房间

```java
// Android示例：加入TRTC房间
private void joinTRTCRoom() {
    try {
        // 1. 从保存的房间信息中获取参数
        String roomId = roomInfo.getString("room_id");
        int sdkAppId = roomInfo.getInt("sdk_app_id");
        String userId = roomInfo.getString("user_id");
        String userSig = roomInfo.getString("user_sig");
        
        // 2. 初始化TRTC SDK
        trtcCloud = TRTCCloud.sharedInstance(getApplicationContext());
        trtcCloud.setListener(new TRTCCloudListener());
        
        // 3. 设置TRTC参数
        TRTCParams trtcParams = new TRTCParams();
        trtcParams.sdkAppId = sdkAppId;
        trtcParams.userId = userId;
        trtcParams.userSig = userSig;
        trtcParams.roomId = Integer.parseInt(roomId);
        
        // 4. 设置视频参数
        TRTCVideoEncParam encParam = new TRTCVideoEncParam();
        encParam.videoResolution = TRTCCloudDef.TRTC_VIDEO_RESOLUTION_640_360;
        encParam.videoFps = 15;
        encParam.videoBitrate = 600;
        trtcCloud.setVideoEncoderParam(encParam);
        
        // 5. 设置本地视频渲染视图
        trtcCloud.startLocalPreview(true, localVideoView);
        
        // 6. 开启本地音频
        trtcCloud.startLocalAudio(TRTCCloudDef.TRTC_AUDIO_QUALITY_SPEECH);
        
        // 7. 加入房间
        trtcCloud.enterRoom(trtcParams, TRTCCloudDef.TRTC_APP_SCENE_VIDEOCALL);
        
    } catch (Exception e) {
        Log.e(TAG, "加入TRTC房间失败", e);
        showError("加入视频通话失败: " + e.getMessage());
    }
}
```

#### 3.2 远程用户视频处理
- 处理远程用户加入/离开事件
- 渲染远程用户视频画面

```java
// Android示例：TRTC监听器
class TRTCCloudListener extends TRTCCloudListenerImpl {
    @Override
    public void onUserVideoAvailable(String userId, boolean available) {
        if (available) {
            // 显示远程用户的视频画面
            trtcCloud.startRemoteView(userId, TRTCCloudDef.TRTC_VIDEO_STREAM_TYPE_SMALL, remoteVideoView);
        } else {
            // 停止显示远程用户的视频画面
            trtcCloud.stopRemoteView(userId, TRTCCloudDef.TRTC_VIDEO_STREAM_TYPE_SMALL);
        }
    }
    
    @Override
    public void onUserAudioAvailable(String userId, boolean available) {
        // 处理远程用户音频可用性变化
    }
    
    @Override
    public void onError(int errCode, String errMsg, Bundle extraInfo) {
        Log.e(TAG, "TRTC错误: " + errCode + ", " + errMsg);
        showError("视频通话错误: " + errMsg);
    }
}
```

#### 3.3 结束通话
- 用户挂断通话时通知后端
- 释放TRTC资源

```java
// Android示例：挂断通话
public void hangupCall(String callId) {
    // 1. 通知后端用户挂断通话
    ApiService apiService = RetrofitClient.getApiService();
    apiService.calleeAction(
        new CalleeActionRequest(callId, "hangup", null)
    ).enqueue(new Callback<ApiResponse>() {
        @Override
        public void onResponse(Call<ApiResponse> call, Response<ApiResponse> response) {
            // 处理响应
        }
        
        @Override
        public void onFailure(Call<ApiResponse> call, Throwable t) {
            Log.e(TAG, "通知挂断失败", t);
        }
    });
    
    // 2. 退出TRTC房间
    exitTRTCRoom();
    
    // 3. 关闭通话界面
    finishCallActivity();
}

// Android示例：退出TRTC房间并释放资源
private void exitTRTCRoom() {
    if (trtcCloud != null) {
        // 停止本地视频预览
        trtcCloud.stopLocalPreview();
        
        // 停止本地音频采集
        trtcCloud.stopLocalAudio();
        
        // 退出房间
        trtcCloud.exitRoom();
        
        // 销毁TRTC实例
        TRTCCloud.destroySharedInstance();
        trtcCloud = null;
    }
}
```

## 三、扩展功能设计

### 1. 后台通知功能

为确保App在后台仍能接收来电通知，需要实现以下功能：

#### 1.1 Android后台服务
- 实现前台Service保持MQTT连接
- 使用通知通道显示来电提醒

```java
// Android示例：前台服务保持MQTT连接
public class MQTTService extends Service {
    
    private MqttAsyncClient mqttClient;
    
    @Override
    public void onCreate() {
        super.onCreate();
        // 创建通知渠道（Android 8.0+）
        createNotificationChannel();
    }
    
    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        // 启动前台服务，提高进程优先级
        startForeground(NOTIFICATION_ID, createNotification("iLock服务运行中"));
        
        // 连接MQTT
        connectMQTT();
        
        // 确保服务不会被系统杀死
        return START_STICKY;
    }
    
    private void handleIncomingCall(JSONObject callInfo) {
        try {
            // 提取来电信息
            String callId = callInfo.getString("call_id");
            String callerId = callInfo.getString("caller_id");
            String callerName = callInfo.getJSONObject("caller_info").getString("name");
            
            // 创建来电通知
            Notification notification = createIncomingCallNotification(callId, callerId, callerName);
            
            // 显示通知
            NotificationManager notificationManager = getSystemService(NotificationManager.class);
            notificationManager.notify(INCOMING_CALL_NOTIFICATION_ID, notification);
            
            // 启动来电Activity
            Intent incomingCallIntent = new Intent(this, IncomingCallActivity.class);
            incomingCallIntent.putExtra("call_info", callInfo.toString());
            incomingCallIntent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
            startActivity(incomingCallIntent);
            
        } catch (JSONException e) {
            Log.e(TAG, "处理来电通知失败", e);
        }
    }
    
    // 其他服务方法...
}
```

#### 1.2 iOS后台模式
- 配置后台模式以支持MQTT长连接
- 使用推送通知唤醒App

```swift
// iOS示例：配置后台模式
// 在AppDelegate中:
func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
    // 请求推送通知权限
    UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge]) { granted, error in
        if granted {
            DispatchQueue.main.async {
                application.registerForRemoteNotifications()
            }
        }
    }
    
    // 配置后台模式
    setupBackgroundModes()
    
    return true
}

private func setupBackgroundModes() {
    // 设置后台任务标识符
    let backgroundTaskIdentifier = UIBackgroundTaskIdentifier.invalid
    
    NotificationCenter.default.addObserver(forName: UIApplication.didEnterBackgroundNotification, object: nil, queue: .main) { [weak self] _ in
        // 申请后台任务
        let taskID = UIApplication.shared.beginBackgroundTask {
            // 后台时间即将耗尽时执行
            UIApplication.shared.endBackgroundTask(backgroundTaskIdentifier)
        }
        
        // 确保MQTT连接保持活跃
        self?.mqttManager.ensureConnection()
    }
}

// 处理接收到的来电推送
func userNotificationCenter(_ center: UNUserNotificationCenter, didReceive response: UNNotificationResponse, withCompletionHandler completionHandler: @escaping () -> Void) {
    let userInfo = response.notification.request.content.userInfo
    
    if let callInfo = userInfo["call_info"] as? [String: Any] {
        // 处理来电信息
        handleIncomingCall(callInfo: callInfo)
    }
    
    completionHandler()
}
```

### 2. 多设备同步处理

当用户有多个设备同时在线时，需要确保通知同步和响应协调：

#### 2.1 设备优先级管理
- 设备根据活跃状态设置优先级
- 服务端向所有设备推送通知，但优先响应高优先级设备

```java
// Android示例：更新设备优先级
private void updateDevicePriority(boolean isActive) {
    ApiService apiService = RetrofitClient.getApiService();
    
    UpdateDevicePriorityRequest request = new UpdateDevicePriorityRequest();
    request.setClientId(clientId);
    request.setStatus(isActive ? "active" : "background");
    request.setPriority(isActive ? 10 : 5);
    request.setLastActive(System.currentTimeMillis());
    
    apiService.updateDevicePriority(request).enqueue(new Callback<ApiResponse>() {
        @Override
        public void onResponse(Call<ApiResponse> call, Response<ApiResponse> response) {
            Log.d(TAG, "设备优先级更新成功");
        }
        
        @Override
        public void onFailure(Call<ApiResponse> call, Throwable t) {
            Log.e(TAG, "设备优先级更新失败", t);
        }
    });
}
```

#### 2.2 设备响应同步
- 一个设备接听/拒绝后，其他设备收到同步消息
- 更新所有设备的界面状态

```java
// Android示例：处理设备同步消息
private void handleDeviceSyncMessage(JSONObject syncInfo) {
    try {
        String action = syncInfo.getString("action");
        String deviceId = syncInfo.getString("device_id");
        String callId = syncInfo.getString("call_id");
        
        // 如果是其他设备处理了来电
        if (!deviceId.equals(clientId)) {
            switch (action) {
                case "accepted":
                    // 其他设备已接听，关闭本设备的来电界面
                    showToast("通话已在其他设备接听");
                    dismissIncomingCallUI();
                    break;
                    
                case "rejected":
                    // 其他设备已拒绝，关闭本设备的来电界面
                    showToast("通话已在其他设备拒绝");
                    dismissIncomingCallUI();
                    break;
                    
                // 其他同步操作...
            }
        }
    } catch (JSONException e) {
        Log.e(TAG, "解析同步消息失败", e);
    }
}
```

## 四、MQTT接口规范

### 1. 主题定义

#### 1.1 基础主题
- **来电通知**: `users/{user_id}/calls/incoming`
- **通话控制**: `users/{user_id}/calls/control`
- **设备同步**: `users/{user_id}/devices/sync`

#### 1.2 QoS级别
- 来电通知: QoS 1 (至少一次送达)
- 通话控制: QoS 1 (至少一次送达)
- 设备同步: QoS 0 (最多一次送达)

### 2. 消息格式规范

#### 2.1 来电通知消息
```json
{
  "call_id": "call_device123_user456_1629123456789",
  "caller_id": "device123",
  "caller_info": {
    "name": "前门设备",
    "location": "小区大门"
  },
  "timestamp": 1679886450000,
  "room_info": {
    "room_id": "664321",
    "sdk_app_id": 1400000001,
    "user_id": "user456",
    "user_sig": "eJwtzM9qwkAUBeCvIrNtYZL5l5..."
  }
}
```

#### 2.2 通话控制消息
```json
{
  "action": "hangup",
  "call_id": "call_device123_user456_1629123456789",
  "timestamp": 1679886550000,
  "reason": "通话已结束"
}
```

#### 2.3 设备同步消息
```json
{
  "action": "accepted",
  "device_id": "user456_ios_abcd1234",
  "call_id": "call_device123_user456_1629123456789",
  "timestamp": 1679886500000
}
```

### 3. HTTP API接口

#### 3.1 设备注册接口
- **请求方式**: `POST /api/mqtt/clients/register`
- **功能**: 注册设备并获取MQTT连接参数
- **请求体**:
```json
{
  "user_id": "user456",
  "device_type": "mobile",
  "client_id": "user456_android_uuid1234",
  "push_token": "firebase_token123",
  "app_version": "1.2.3",
  "platform": "android"
}
```
- **响应**:
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "mqtt_credentials": {
      "broker_url": "tcp://broker.ilock.com:1883",
      "username": "user456",
      "password": "jwt_token_for_mqtt",
      "client_id": "user456_android_uuid1234"
    },
    "topics": {
      "incoming_calls": "users/user456/calls/incoming",
      "call_control": "users/user456/calls/control",
      "device_sync": "users/user456/devices/sync"
    }
  }
}
```

#### 3.2 设备状态更新接口
- **请求方式**: `PUT /api/mqtt/clients/status`
- **功能**: 更新设备在线状态和优先级
- **请求体**:
```json
{
  "client_id": "user456_android_uuid1234", 
  "status": "active",
  "priority": 10,
  "last_active": 1679886400000
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

## 五、安全考虑

### 1. 认证与授权
- 使用JWT令牌进行MQTT客户端认证
- 确保只有授权设备能订阅用户主题
- 设置主题ACL(访问控制列表)限制

### 2. 数据加密
- 使用TLS加密MQTT连接
- 敏感信息(如UserSig)传输时进行额外加密

### 3. 消息校验
- 在消息中加入时间戳防止重放攻击
- 实现消息有效期检查，过期消息自动丢弃

## 六、测试与验证

### 1. 测试方案
- 使用MQTT客户端工具模拟消息发送和接收
- 编写单元测试验证消息处理逻辑
- 进行多设备并发测试，验证同步机制

### 2. 性能指标
- 消息推送延迟: <500ms
- 连接恢复时间: <3s
- 设备注册响应时间: <1s