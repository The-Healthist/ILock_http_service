package services

import (
	"encoding/json"
	"fmt"
	"ilock-http-service/config"
	"log"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/gorm"
)

// InterfaceMQTTCallService defines the MQTT call service interface
type InterfaceMQTTCallService interface {
	Connect() error
	Disconnect()
	InitiateCall(deviceID string, userID string) (string, error)
	HandleCallerAction(callID string, action string, reason string) error
	HandleCalleeAction(callID string, action string, reason string) error
	GetOrCreateCallSession(callID string, callerID string, calleeID string) *CallSession
	GetCallSession(callID string) *CallSession
	EndCallSession(callID string)
	SendIncomingCallNotification(userID string, notification *CallIncomingNotification) error
	SendCallerControlMessage(deviceID string, control *CallControl) error
	SendCalleeControlMessage(userID string, control *CallControl) error
	PublishDeviceStatus(deviceID string, status map[string]interface{}) error
	PublishSystemMessage(messageType string, message map[string]interface{}) error
}

// MessageType 定义消息类型
type MessageType string

const (
	MessageTypeCallRequest   MessageType = "call_request"
	MessageTypeCallIncoming  MessageType = "call_incoming"
	MessageTypeCallControl   MessageType = "call_control"
	MessageTypeDeviceControl MessageType = "device_control"
	MessageTypeDeviceStatus  MessageType = "device_status"
	MessageTypeSystemMessage MessageType = "system_message"
)

// CallAction 定义通话控制动作
type CallAction string

const (
	// Caller 接收的动作
	ActionRinging  CallAction = "ringing"  // 正在呼叫
	ActionRejected CallAction = "rejected" // 被拒绝
	ActionHangup   CallAction = "hangup"   // 挂断
	ActionTimeout  CallAction = "timeout"  // 超时
	ActionError    CallAction = "error"    // 错误

	// Callee 接收的动作
	ActionCancelled CallAction = "cancelled" // 呼叫取消
)

// 主题常量
const (
	// 呼叫请求 (Caller -> Backend)
	TopicCallRequestFormat = "calls/request/%s" // %s: caller_device_id

	// 来电通知 (Backend -> Callee)
	TopicIncomingCallFormat = "users/%s/calls/incoming" // %s: user_id

	// 通话控制 (Backend -> Caller)
	TopicCallerControlFormat = "devices/%s/calls/control" // %s: caller_device_id

	// 通话控制 (Backend -> Callee)
	TopicCalleeControlFormat = "users/%s/calls/control" // %s: user_id

	// 设备状态 (Device -> Backend 或 Backend -> Device)
	TopicDeviceStatusFormat = "devices/%s/status" // %s: device_id

	// 系统消息 (Backend -> All)
	TopicSystemMessageFormat = "system/%s" // %s: message_type
)

// MQTTCallService 整合MQTT和通话服务
type MQTTCallService struct {
	DB           *gorm.DB
	Config       *config.Config
	RTCService   InterfaceTencentRTCService
	Client       mqtt.Client
	IsConnected  bool
	ActiveCalls  map[string]*CallSession
	CallHandlers map[string]func(payload []byte)
	mu           sync.RWMutex // 用于保护 CallHandlers 和 ActiveCalls
}

// CallSession 表示一个进行中的通话会话
type CallSession struct {
	CallID         string
	CallerID       string
	CalleeID       string
	RoomID         string
	CallState      string // "requesting", "ringing", "connected", "ended"
	StartTimestamp int64
	EndTimestamp   int64
}

// 消息结构体定义
type (
	// MQTTMessage MQTT消息基础结构
	MQTTMessage struct {
		Type      MessageType    `json:"type"`
		Timestamp int64          `json:"timestamp"`
		Payload   map[string]any `json:"payload"`
	}

	// CallRequest 呼叫请求结构
	CallRequest struct {
		CallID       string `json:"call_id"`        // 本次呼叫的唯一ID
		CallerID     string `json:"caller_id"`      // 呼叫方设备ID
		TargetUserID string `json:"target_user_id"` // 目标用户ID
		Timestamp    int64  `json:"timestamp"`      // 发起呼叫的Unix毫秒时间戳
	}

	// CallIncomingNotification 来电通知结构
	CallIncomingNotification struct {
		CallID     string     `json:"call_id"`     // 对应呼叫请求的ID
		CallerID   string     `json:"caller_id"`   // 呼叫方设备ID
		CallerInfo CallerInfo `json:"caller_info"` // 呼叫方信息
		Timestamp  int64      `json:"timestamp"`   // 发送通知的Unix毫秒时间戳
		RoomInfo   RoomInfo   `json:"room_info"`   // TRTC房间信息
	}

	// CallerInfo 呼叫方信息
	CallerInfo struct {
		Name string `json:"name"` // 设备名称或用户昵称
	}

	// RoomInfo TRTC房间信息
	RoomInfo struct {
		RoomID   string `json:"room_id"`    // TRTC房间号
		SDKAppID int    `json:"sdk_app_id"` // TRTC应用ID
		UserID   string `json:"user_id"`    // 用户ID
		UserSig  string `json:"user_sig"`   // TRTC签名
	}

	// CallControl 通话控制消息
	CallControl struct {
		Action    CallAction `json:"action"`           // 控制动作类型
		CallID    string     `json:"call_id"`          // 对应呼叫请求的ID
		Timestamp int64      `json:"timestamp"`        // 发送指令的Unix毫秒时间戳
		Reason    string     `json:"reason,omitempty"` // 可选，提供额外信息
	}

	// DeviceStatus 设备状态结构
	DeviceStatus struct {
		DeviceID   string                 `json:"device_id"`
		Online     bool                   `json:"online"`
		Battery    int                    `json:"battery"`
		LastUpdate int64                  `json:"last_update"`
		Properties map[string]interface{} `json:"properties"`
	}

	// SystemMessage 系统消息结构
	SystemMessage struct {
		Type      string                 `json:"type"`
		Level     string                 `json:"level"` // info, warning, error
		Message   string                 `json:"message"`
		Timestamp int64                  `json:"timestamp"`
		Data      map[string]interface{} `json:"data,omitempty"`
	}
)

// NewMQTTCallService 创建一个新的MQTT通话服务
func NewMQTTCallService(db *gorm.DB, cfg *config.Config, rtcService InterfaceTencentRTCService) InterfaceMQTTCallService {
	service := &MQTTCallService{
		DB:           db,
		Config:       cfg,
		RTCService:   rtcService,
		IsConnected:  false,
		ActiveCalls:  make(map[string]*CallSession),
		CallHandlers: make(map[string]func(payload []byte)),
	}

	// 初始化MQTT客户端
	service.setupMQTTClient()

	return service
}

// setupMQTTClient 设置MQTT客户端
func (s *MQTTCallService) setupMQTTClient() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(s.Config.MQTTBrokerURL)
	opts.SetClientID(s.Config.MQTTClientID)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Second * 30)
	opts.SetKeepAlive(time.Second * 60)
	opts.SetPingTimeout(time.Second * 10)
	opts.SetCleanSession(true)
	opts.SetOrderMatters(true)

	if s.Config.MQTTUsername != "" {
		opts.SetUsername(s.Config.MQTTUsername)
		opts.SetPassword(s.Config.MQTTPassword)
	}

	// 设置连接丢失回调
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("[MQTT] 连接丢失: %v", err)
		s.IsConnected = false

		// 尝试重新连接
		go func() {
			for !s.IsConnected {
				log.Println("[MQTT] 尝试重新连接...")
				if token := client.Connect(); token.Wait() && token.Error() != nil {
					log.Printf("[MQTT] 重新连接失败: %v", token.Error())
					time.Sleep(5 * time.Second)
				} else {
					log.Println("[MQTT] 成功重连")
					s.IsConnected = true
					// 重新订阅主题
					s.subscribeToTopics()
				}
			}
		}()
	})

	// 设置连接建立回调
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("[MQTT] 成功连接")
		s.IsConnected = true

		// 订阅主题
		s.subscribeToTopics()
	})

	// 创建客户端
	s.Client = mqtt.NewClient(opts)
}

// Connect 连接到MQTT服务器
func (s *MQTTCallService) Connect() error {
	if token := s.Client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("[MQTT] 连接失败: %v", token.Error())
	}
	return nil
}

// Disconnect 断开与MQTT服务器的连接
func (s *MQTTCallService) Disconnect() {
	if s.Client != nil && s.Client.IsConnected() {
		s.Client.Disconnect(250)
		s.IsConnected = false
		log.Println("[MQTT] 已断开连接")
	}
}

// 会话管理方法
func (s *MQTTCallService) GetOrCreateCallSession(callID string, callerID string, calleeID string) *CallSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.ActiveCalls[callID]; exists {
		return session
	}

	session := &CallSession{
		CallID:         callID,
		CallerID:       callerID,
		CalleeID:       calleeID,
		CallState:      "requesting",
		StartTimestamp: time.Now().UnixMilli(),
	}

	s.ActiveCalls[callID] = session
	return session
}

func (s *MQTTCallService) GetCallSession(callID string) *CallSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ActiveCalls[callID]
}

func (s *MQTTCallService) EndCallSession(callID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.ActiveCalls[callID]; exists {
		session.CallState = "ended"
		session.EndTimestamp = time.Now().UnixMilli()

		// 保存通话记录到数据库
		duration := session.EndTimestamp - session.StartTimestamp
		callRecord := struct {
			CallID     string
			CallerID   string
			CalleeID   string
			StartTime  time.Time
			EndTime    time.Time
			Duration   int64
			CallResult string
		}{
			CallID:     session.CallID,
			CallerID:   session.CallerID,
			CalleeID:   session.CalleeID,
			StartTime:  time.UnixMilli(session.StartTimestamp),
			EndTime:    time.UnixMilli(session.EndTimestamp),
			Duration:   duration / 1000, // 转换为秒
			CallResult: session.CallState,
		}

		s.DB.Create(&callRecord)

		// 从活动通话中移除
		delete(s.ActiveCalls, callID)
	}
}

// 通话控制方法
func (s *MQTTCallService) InitiateCall(deviceID string, userID string) (string, error) {
	// 生成唯一呼叫ID
	callID := fmt.Sprintf("call_%s_%s_%d", deviceID, userID, time.Now().UnixMilli())

	// 创建会话
	session := s.GetOrCreateCallSession(callID, deviceID, userID)

	// 创建房间ID
	roomID, err := s.RTCService.CreateVideoCall(deviceID, userID)
	if err != nil {
		return "", fmt.Errorf("创建视频通话失败: %v", err)
	}

	session.RoomID = roomID

	// 为被呼叫方生成UserSig
	userSigInfo, err := s.RTCService.GetUserSig(userID)
	if err != nil {
		return "", fmt.Errorf("为被呼叫方生成UserSig失败: %v", err)
	}

	// 发送来电通知
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

	if err := s.SendIncomingCallNotification(userID, notification); err != nil {
		return "", fmt.Errorf("发送来电通知失败: %v", err)
	}

	// 发送响铃状态给呼叫方
	controlMsg := &CallControl{
		Action:    ActionRinging,
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
	}

	if err := s.SendCallerControlMessage(deviceID, controlMsg); err != nil {
		return "", fmt.Errorf("发送响铃状态失败: %v", err)
	}

	return callID, nil
}

func (s *MQTTCallService) HandleCallerAction(callID string, action string, reason string) error {
	session := s.GetCallSession(callID)
	if session == nil {
		return fmt.Errorf("通话不存在")
	}

	var callAction CallAction
	switch action {
	case "cancelled":
		callAction = ActionCancelled
	case "hangup":
		callAction = ActionHangup
	default:
		return fmt.Errorf("不支持的动作: %s", action)
	}

	controlMsg := &CallControl{
		Action:    callAction,
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
		Reason:    reason,
	}

	if err := s.SendCalleeControlMessage(session.CalleeID, controlMsg); err != nil {
		return fmt.Errorf("发送控制消息失败: %v", err)
	}

	s.EndCallSession(callID)
	return nil
}

func (s *MQTTCallService) HandleCalleeAction(callID string, action string, reason string) error {
	session := s.GetCallSession(callID)
	if session == nil {
		return fmt.Errorf("通话不存在")
	}

	var callAction CallAction
	switch action {
	case "rejected":
		callAction = ActionRejected
	case "hangup":
		callAction = ActionHangup
	case "timeout":
		callAction = ActionTimeout
	default:
		return fmt.Errorf("不支持的动作: %s", action)
	}

	controlMsg := &CallControl{
		Action:    callAction,
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
		Reason:    reason,
	}

	if err := s.SendCallerControlMessage(session.CallerID, controlMsg); err != nil {
		return fmt.Errorf("发送控制消息失败: %v", err)
	}

	s.EndCallSession(callID)
	return nil
}

// MQTT消息处理方法
func (s *MQTTCallService) subscribeToTopics() {
	// 订阅呼叫请求主题
	token := s.Client.Subscribe("calls/request/+", byte(s.Config.MQTTQoS), s.handleCallRequest)
	if token.Wait() && token.Error() != nil {
		log.Printf("[MQTT] 订阅呼叫请求主题失败: %v", token.Error())
	}
}

func (s *MQTTCallService) handleCallRequest(client mqtt.Client, msg mqtt.Message) {
	log.Printf("[MQTT] 收到呼叫请求: Topic: %s", msg.Topic())

	var request CallRequest
	if err := json.Unmarshal(msg.Payload(), &request); err != nil {
		log.Printf("[MQTT] 解析呼叫请求失败: %v", err)
		return
	}

	s.mu.RLock()
	handler, exists := s.CallHandlers[request.CallID]
	s.mu.RUnlock()

	if exists {
		handler(msg.Payload())
	} else {
		s.processCallRequest(&request)
	}
}

func (s *MQTTCallService) processCallRequest(request *CallRequest) {
	log.Printf("[MQTT] 处理呼叫请求: CallID=%s, CallerID=%s, TargetUserID=%s",
		request.CallID, request.CallerID, request.TargetUserID)

	// 验证请求
	if request.CallID == "" || request.CallerID == "" || request.TargetUserID == "" {
		log.Printf("[MQTT] 无效的呼叫请求参数")
		s.sendErrorToDevice(request.CallerID, request.CallID, "无效的请求参数")
		return
	}

	// 创建会话并处理呼叫
	if _, err := s.InitiateCall(request.CallerID, request.TargetUserID); err != nil {
		log.Printf("[MQTT] 处理呼叫请求失败: %v", err)
		s.sendErrorToDevice(request.CallerID, request.CallID, err.Error())
	}
}

// MQTT消息发送方法
func (s *MQTTCallService) SendIncomingCallNotification(userID string, notification *CallIncomingNotification) error {
	topic := fmt.Sprintf(TopicIncomingCallFormat, userID)
	return s.publishMessage(topic, notification)
}

func (s *MQTTCallService) SendCallerControlMessage(deviceID string, control *CallControl) error {
	topic := fmt.Sprintf(TopicCallerControlFormat, deviceID)
	return s.publishMessage(topic, control)
}

func (s *MQTTCallService) SendCalleeControlMessage(userID string, control *CallControl) error {
	topic := fmt.Sprintf(TopicCalleeControlFormat, userID)
	return s.publishMessage(topic, control)
}

func (s *MQTTCallService) publishMessage(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	token := s.Client.Publish(topic, byte(s.Config.MQTTQoS), s.Config.MQTTRetained, data)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("发布消息失败: %v", token.Error())
	}

	return nil
}

func (s *MQTTCallService) sendErrorToDevice(deviceID string, callID string, reason string) {
	errorMsg := &CallControl{
		Action:    ActionError,
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
		Reason:    reason,
	}

	if err := s.SendCallerControlMessage(deviceID, errorMsg); err != nil {
		log.Printf("[MQTT] 发送错误消息失败: %v", err)
	}
}

// 设备状态和系统消息方法
func (s *MQTTCallService) PublishDeviceStatus(deviceID string, status map[string]interface{}) error {
	topic := fmt.Sprintf(TopicDeviceStatusFormat, deviceID)
	return s.publishMessage(topic, status)
}

func (s *MQTTCallService) PublishSystemMessage(messageType string, message map[string]interface{}) error {
	topic := fmt.Sprintf(TopicSystemMessageFormat, messageType)
	return s.publishMessage(topic, message)
}

func (s *MQTTCallService) handleDeviceStatus(client mqtt.Client, msg mqtt.Message) {
	log.Printf("[MQTT] 收到设备状态: Topic: %s", msg.Topic())

	deviceID := extractDeviceIDFromTopic(msg.Topic())
	if deviceID == "" {
		log.Printf("[MQTT] 无法从主题中提取设备ID: %s", msg.Topic())
		return
	}

	var status DeviceStatus
	if err := json.Unmarshal(msg.Payload(), &status); err != nil {
		log.Printf("[MQTT] 解析设备状态失败: %v", err)
		return
	}

	status.DeviceID = deviceID
	status.LastUpdate = time.Now().UnixMilli()

	log.Printf("[MQTT] 设备状态更新: DeviceID=%s, Online=%v, Battery=%d%%",
		status.DeviceID, status.Online, status.Battery)
}

func (s *MQTTCallService) handleSystemMessage(client mqtt.Client, msg mqtt.Message) {
	log.Printf("[MQTT] 收到系统消息: Topic: %s", msg.Topic())

	var sysMsg SystemMessage
	if err := json.Unmarshal(msg.Payload(), &sysMsg); err != nil {
		log.Printf("[MQTT] 解析系统消息失败: %v", err)
		return
	}

	switch sysMsg.Level {
	case "error":
		log.Printf("[MQTT] 系统错误: %s - %s", sysMsg.Type, sysMsg.Message)
	case "warning":
		log.Printf("[MQTT] 系统警告: %s - %s", sysMsg.Type, sysMsg.Message)
	default:
		log.Printf("[MQTT] 系统信息: %s - %s", sysMsg.Type, sysMsg.Message)
	}

	switch sysMsg.Type {
	case "device_offline":
		if deviceID, ok := sysMsg.Data["device_id"].(string); ok {
			log.Printf("[MQTT] 设备离线: %s", deviceID)
		}
	case "device_online":
		if deviceID, ok := sysMsg.Data["device_id"].(string); ok {
			log.Printf("[MQTT] 设备上线: %s", deviceID)
		}
	case "system_maintenance":
		log.Printf("[MQTT] 系统维护通知: %s", sysMsg.Message)
	}
}

// 工具方法
func extractDeviceIDFromTopic(topic string) string {
	parts := strings.Split(topic, "/")
	if len(parts) != 3 || parts[0] != "devices" || parts[2] != "status" {
		return ""
	}
	return parts[1]
}
