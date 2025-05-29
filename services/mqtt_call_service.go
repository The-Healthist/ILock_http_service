package services

import (
	"encoding/json"
	"fmt"
	"ilock-http-service/config"
	"ilock-http-service/models"
	"log"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InterfaceMQTTCallService 定义MQTT通话服务接口
type InterfaceMQTTCallService interface {
	Connect() error
	Disconnect()
	InitiateCall(deviceID, residentID string) (string, error)
	InitiateCallToAll(deviceID string) (string, []string, error)
	InitiateCallToHousehold(deviceID string, householdID string) (string, []string, error)
	InitiateCallByPhone(deviceID string, phone string) (string, []string, error)
	HandleCallerAction(callID, action, reason string) error
	HandleCalleeAction(callID, action, reason string) error
	GetCallSession(callID string) (*models.CallSession, bool)
	EndCallSession(callID, reason string) error
	CleanupTimedOutSessions() int
	SubscribeToTopics() error
	PublishDeviceStatus(deviceID string, status map[string]interface{}) error
	PublishSystemMessage(messageType string, message map[string]interface{}) error
}

// MQTTCallService 整合MQTT和通话服务的实现
type MQTTCallService struct {
	DB              *gorm.DB
	Config          *config.Config
	RTCService      InterfaceTencentRTCService
	Client          mqtt.Client
	IsConnected     bool
	CallManager     *models.CallManager
	TopicHandlers   map[string]mqtt.MessageHandler
	CallRecordMutex sync.Mutex // 用于保护通话记录创建
}

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

// 消息结构体定义
type (
	// MQTTMessage MQTT消息基础结构
	MQTTMessage struct {
		Type      string         `json:"type"`
		Timestamp int64          `json:"timestamp"`
		Payload   map[string]any `json:"payload"`
	}

	// CallRequest 呼叫请求结构
	CallRequest struct {
		DeviceID     string `json:"device_id"`          // 呼叫方设备ID
		TargetUserID string `json:"target_resident_id"` // 目标用户ID
		Timestamp    int64  `json:"timestamp"`          // 发起呼叫的Unix毫秒时间戳
	}

	// CallResponse 呼叫响应结构
	CallResponse struct {
		CallID       string   `json:"call_id"`   // 本次呼叫的唯一ID
		DeviceID     string   `json:"device_id"` // 呼叫方设备ID
		TargetUserID string   `json:"target_resident_id"`
		Timestamp    int64    `json:"timestamp"`
		TRTCInfo     TRTCInfo `json:"tencen_rtc"` // 腾讯云TRTC信息
		CallInfo     CallInfo `json:"call_info"`  // 通话信息
	}

	// TRTCInfo 腾讯云RTC信息
	TRTCInfo struct {
		RoomIDType string `json:"room_id_type"`
		RoomID     string `json:"room_id"`
		SDKAppID   int    `json:"sdk_app_id"`
		UserID     string `json:"user_id"`
		UserSig    string `json:"user_sig"`
	}

	// CallInfo 通话控制信息
	CallInfo struct {
		Action    string `json:"action"`
		CallID    string `json:"call_id"`
		Timestamp int64  `json:"timestamp"`
		Reason    string `json:"reason,omitempty"`
	}

	// IncomingCallNotification 来电通知
	IncomingCallNotification struct {
		CallID       string   `json:"call_id"`
		DeviceID     string   `json:"device_id"` // 设备ID
		TargetUserID string   `json:"target_resident_id"`
		Timestamp    int64    `json:"timestamp"`
		TRTCInfo     TRTCInfo `json:"tencen_rtc"`
	}
)

// NewMQTTCallService 创建一个新的MQTT通话服务实现
func NewMQTTCallService(db *gorm.DB, cfg *config.Config, rtcService InterfaceTencentRTCService) InterfaceMQTTCallService {
	service := &MQTTCallService{
		DB:            db,
		Config:        cfg,
		RTCService:    rtcService,
		CallManager:   models.NewCallManager(),
		TopicHandlers: make(map[string]mqtt.MessageHandler),
		IsConnected:   false,
	}

	// 设置MQTT客户端
	service.setupMQTTClient()

	// 设置主题处理程序
	service.setupTopicHandlers()

	// 启动会话清理定时任务
	go service.startSessionCleanupTask()

	return service
}

// setupMQTTClient 设置MQTT客户端
func (s *MQTTCallService) setupMQTTClient() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(s.Config.MQTTBrokerURL)
	opts.SetClientID(fmt.Sprintf("%s-%s", s.Config.MQTTClientID, uuid.New().String()[:8]))
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
					s.SubscribeToTopics()
				}
			}
		}()
	})

	// 设置连接建立回调
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("[MQTT] 成功连接")
		s.IsConnected = true

		// 订阅主题
		s.SubscribeToTopics()
	})

	// 创建客户端
	s.Client = mqtt.NewClient(opts)
}

// setupTopicHandlers 设置主题处理程序
func (s *MQTTCallService) setupTopicHandlers() {
	// 呼叫请求处理程序
	s.TopicHandlers["calls/request/+"] = s.handleCallRequest

	// 设备状态处理程序
	s.TopicHandlers["devices/+/status"] = s.handleDeviceStatus
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
	}
}

// SubscribeToTopics 订阅相关主题
func (s *MQTTCallService) SubscribeToTopics() error {
	for topic, handler := range s.TopicHandlers {
		if token := s.Client.Subscribe(topic, 1, handler); token.Wait() && token.Error() != nil {
			return fmt.Errorf("订阅主题失败 [%s]: %v", topic, token.Error())
		}
		log.Printf("[MQTT] 已订阅主题: %s", topic)
	}
	return nil
}

// InitiateCall 发起通话
func (s *MQTTCallService) InitiateCall(deviceID, residentID string) (string, error) {
	// 生成唯一的通话ID
	callID := uuid.New().String()

	// 创建TRTC房间并生成签名
	rtcRoomID, err := s.RTCService.CreateVideoCall(deviceID, residentID)
	if err != nil {
		return "", fmt.Errorf("创建TRTC房间失败: %v", err)
	}

	// 为住户生成UserSig
	tokenInfo, err := s.RTCService.GetUserSig(residentID)
	if err != nil {
		return "", fmt.Errorf("生成UserSig失败: %v", err)
	}

	// 创建TRTC信息
	trtcInfo := models.TRTCInfo{
		RoomID:     rtcRoomID,
		RoomIDType: "string",
		SDKAppID:   tokenInfo.SDKAppID,
		UserID:     tokenInfo.UserID,
		UserSig:    tokenInfo.UserSig,
	}

	// 创建通话会话
	_, err = s.CallManager.CreateSession(callID, deviceID, residentID, trtcInfo)
	if err != nil {
		return "", fmt.Errorf("创建通话会话失败: %v", err)
	}

	// 发送呼入通知给住户
	incomingNotification := IncomingCallNotification{
		CallID:       callID,
		DeviceID:     deviceID,
		TargetUserID: residentID,
		Timestamp:    time.Now().UnixMilli(),
		TRTCInfo: TRTCInfo{
			RoomIDType: trtcInfo.RoomIDType,
			RoomID:     trtcInfo.RoomID,
			SDKAppID:   trtcInfo.SDKAppID,
			UserID:     trtcInfo.UserID,
			UserSig:    trtcInfo.UserSig,
		},
	}

	// 发布到住户的呼入通知主题
	incomingTopic := fmt.Sprintf(TopicIncomingCallFormat, residentID)
	if err := s.publishMessage(incomingTopic, incomingNotification); err != nil {
		// 发送失败，结束会话
		s.CallManager.EndSession(callID, "发送通知失败")
		return "", fmt.Errorf("发送呼入通知失败: %v", err)
	}

	// 更新会话状态为振铃中
	s.CallManager.UpdateSessionStatus(callID, "ringing")

	// 发送振铃控制消息给设备
	callerControl := CallInfo{
		Action:    "ringing",
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
	}
	callerTopic := fmt.Sprintf(TopicCallerControlFormat, deviceID)
	if err := s.publishMessage(callerTopic, callerControl); err != nil {
		log.Printf("[MQTT] 发送振铃控制消息失败: %v", err)
	}

	// 创建通话记录
	s.createCallRecord(callID, deviceID, residentID, "ringing")

	return callID, nil
}

// HandleCallerAction 处理呼叫方动作
func (s *MQTTCallService) HandleCallerAction(callID, action, reason string) error {
	// 获取会话
	session, exists := s.CallManager.GetSession(callID)
	if !exists {
		return fmt.Errorf("会话不存在: %s", callID)
	}

	// 更新会话状态
	var newStatus string
	switch action {
	case "hangup":
		newStatus = "ended"
	case "cancelled":
		newStatus = "cancelled"
	default:
		return fmt.Errorf("不支持的动作: %s", action)
	}

	// 更新会话状态
	if err := s.CallManager.UpdateSessionStatus(callID, newStatus); err != nil {
		return err
	}

	// 发送控制消息给被呼叫方
	calleeControl := CallInfo{
		Action:    action,
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
		Reason:    reason,
	}
	calleeTopic := fmt.Sprintf(TopicCalleeControlFormat, session.ResidentID)
	if err := s.publishMessage(calleeTopic, calleeControl); err != nil {
		log.Printf("[MQTT] 发送控制消息给被呼叫方失败: %v", err)
	}

	// 如果是结束通话的动作，结束会话
	if action == "hangup" || action == "cancelled" {
		if _, err := s.CallManager.EndSession(callID, reason); err != nil {
			return err
		}

		// 更新通话记录
		s.updateCallRecord(callID, "caller_"+action, reason)
	}

	return nil
}

// HandleCalleeAction 处理被呼叫方动作
func (s *MQTTCallService) HandleCalleeAction(callID, action, reason string) error {
	// 获取会话
	session, exists := s.CallManager.GetSession(callID)
	if !exists {
		return fmt.Errorf("会话不存在: %s", callID)
	}

	// 更新会话状态
	var newStatus string
	switch action {
	case "rejected":
		newStatus = "rejected"
	case "answered":
		newStatus = "connected"
	case "hangup":
		newStatus = "ended"
	case "timeout":
		newStatus = "timeout"
	default:
		return fmt.Errorf("不支持的动作: %s", action)
	}

	// 更新会话状态
	if err := s.CallManager.UpdateSessionStatus(callID, newStatus); err != nil {
		return err
	}

	// 发送控制消息给呼叫方
	callerControl := CallInfo{
		Action:    action,
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
		Reason:    reason,
	}
	callerTopic := fmt.Sprintf(TopicCallerControlFormat, session.DeviceID)
	if err := s.publishMessage(callerTopic, callerControl); err != nil {
		log.Printf("[MQTT] 发送控制消息给呼叫方失败: %v", err)
	}

	// 如果是结束通话的动作，结束会话
	if action == "rejected" || action == "hangup" || action == "timeout" {
		if _, err := s.CallManager.EndSession(callID, reason); err != nil {
			return err
		}

		// 更新通话记录
		s.updateCallRecord(callID, "callee_"+action, reason)
	}

	return nil
}

// GetCallSession 获取指定通话会话
func (s *MQTTCallService) GetCallSession(callID string) (*models.CallSession, bool) {
	return s.CallManager.GetSession(callID)
}

// EndCallSession 结束通话会话
func (s *MQTTCallService) EndCallSession(callID, reason string) error {
	session, err := s.CallManager.EndSession(callID, reason)
	if err != nil {
		return err
	}

	// 向双方发送通话结束通知
	endInfo := CallInfo{
		Action:    "ended",
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
		Reason:    reason,
	}

	// 通知呼叫方
	callerTopic := fmt.Sprintf(TopicCallerControlFormat, session.DeviceID)
	if err := s.publishMessage(callerTopic, endInfo); err != nil {
		log.Printf("[MQTT] 发送结束通知给呼叫方失败: %v", err)
	}

	// 通知被呼叫方
	calleeTopic := fmt.Sprintf(TopicCalleeControlFormat, session.ResidentID)
	if err := s.publishMessage(calleeTopic, endInfo); err != nil {
		log.Printf("[MQTT] 发送结束通知给被呼叫方失败: %v", err)
	}

	// 更新通话记录
	s.updateCallRecord(callID, "system_ended", reason)

	return nil
}

// CleanupTimedOutSessions 清理超时会话
func (s *MQTTCallService) CleanupTimedOutSessions() int {
	// 呼叫中状态超时时间: 30秒
	ringTimeout := 30 * time.Second
	// 通话中状态超时时间: 2小时
	callTimeout := 2 * time.Hour

	return s.CallManager.CleanupTimedOutSessions(callTimeout, ringTimeout)
}

// startSessionCleanupTask 启动会话清理定时任务
func (s *MQTTCallService) startSessionCleanupTask() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cleanedCount := s.CleanupTimedOutSessions()
			if cleanedCount > 0 {
				log.Printf("[MQTT] 清理超时会话: %d 个", cleanedCount)
			}
		}
	}
}

// handleCallRequest 处理呼叫请求
func (s *MQTTCallService) handleCallRequest(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()

	// 解析呼叫请求
	var request CallRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		log.Printf("[MQTT] 解析呼叫请求失败: %v", err)
		return
	}

	// 提取设备ID
	deviceID := extractDeviceIDFromTopic(topic)
	if deviceID == "" {
		log.Printf("[MQTT] 无法从主题提取设备ID: %s", topic)
		return
	}

	// 验证设备ID
	if deviceID != request.DeviceID {
		log.Printf("[MQTT] 设备ID不匹配: 主题=%s, 请求=%s", deviceID, request.DeviceID)
		return
	}

	// 发起通话
	callID, err := s.InitiateCall(request.DeviceID, request.TargetUserID)
	if err != nil {
		log.Printf("[MQTT] 发起通话失败: %v", err)
		// 发送错误响应给设备
		s.sendErrorToDevice(request.DeviceID, "", fmt.Sprintf("发起通话失败: %v", err))
		return
	}

	// 获取会话信息
	session, exists := s.CallManager.GetSession(callID)
	if !exists {
		log.Printf("[MQTT] 通话会话不存在: %s", callID)
		return
	}

	// 构建响应
	response := CallResponse{
		CallID:       callID,
		DeviceID:     request.DeviceID,
		TargetUserID: request.TargetUserID,
		Timestamp:    time.Now().UnixMilli(),
		TRTCInfo: TRTCInfo{
			RoomIDType: session.TRTCInfo.RoomIDType,
			RoomID:     session.TRTCInfo.RoomID,
			SDKAppID:   session.TRTCInfo.SDKAppID,
			UserID:     session.TRTCInfo.UserID,
			UserSig:    session.TRTCInfo.UserSig,
		},
		CallInfo: CallInfo{
			Action:    "ringing",
			CallID:    callID,
			Timestamp: time.Now().UnixMilli(),
		},
	}

	// 发送响应
	responseTopic := fmt.Sprintf(TopicCallerControlFormat, request.DeviceID)
	if err := s.publishMessage(responseTopic, response); err != nil {
		log.Printf("[MQTT] 发送呼叫响应失败: %v", err)
	}
}

// handleDeviceStatus 处理设备状态消息
func (s *MQTTCallService) handleDeviceStatus(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()

	// 提取设备ID
	deviceID := extractDeviceIDFromTopic(topic)
	if deviceID == "" {
		log.Printf("[MQTT] 无法从主题提取设备ID: %s", topic)
		return
	}

	// 解析设备状态
	var status map[string]interface{}
	if err := json.Unmarshal(payload, &status); err != nil {
		log.Printf("[MQTT] 解析设备状态失败: %v", err)
		return
	}

	// 更新设备状态（在实际应用中，你可能需要将状态存储到数据库）
	log.Printf("[MQTT] 设备状态更新: ID=%s, 状态=%v", deviceID, status)
}

// publishMessage 发布消息到指定主题
func (s *MQTTCallService) publishMessage(topic string, payload interface{}) error {
	// 序列化消息
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 发布消息
	if token := s.Client.Publish(topic, 1, false, jsonData); token.Wait() && token.Error() != nil {
		return fmt.Errorf("发布消息失败: %v", token.Error())
	}

	log.Printf("[MQTT] 已发布消息到主题: %s", topic)
	return nil
}

// sendErrorToDevice 发送错误消息给设备
func (s *MQTTCallService) sendErrorToDevice(deviceID, callID, reason string) {
	errorInfo := CallInfo{
		Action:    "error",
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
		Reason:    reason,
	}

	topic := fmt.Sprintf(TopicCallerControlFormat, deviceID)
	if err := s.publishMessage(topic, errorInfo); err != nil {
		log.Printf("[MQTT] 发送错误消息给设备失败: %v", err)
	}
}

// PublishDeviceStatus 发布设备状态
func (s *MQTTCallService) PublishDeviceStatus(deviceID string, status map[string]interface{}) error {
	topic := fmt.Sprintf(TopicDeviceStatusFormat, deviceID)
	return s.publishMessage(topic, status)
}

// PublishSystemMessage 发布系统消息
func (s *MQTTCallService) PublishSystemMessage(messageType string, message map[string]interface{}) error {
	topic := fmt.Sprintf(TopicSystemMessageFormat, messageType)
	return s.publishMessage(topic, message)
}

// createCallRecord 创建通话记录
func (s *MQTTCallService) createCallRecord(callID, deviceID, residentID, status string) {
	s.CallRecordMutex.Lock()
	defer s.CallRecordMutex.Unlock()

	// 在实际实现中，这里应该将通话记录保存到数据库
	// 这里只是示例，实际使用时需要替换成真实的数据库操作
	log.Printf("[MQTT] 创建通话记录: ID=%s, 设备=%s, 住户=%s, 状态=%s",
		callID, deviceID, residentID, status)
}

// updateCallRecord 更新通话记录
func (s *MQTTCallService) updateCallRecord(callID, status, reason string) {
	s.CallRecordMutex.Lock()
	defer s.CallRecordMutex.Unlock()

	// 在实际实现中，这里应该更新数据库中的通话记录
	// 这里只是示例，实际使用时需要替换成真实的数据库操作
	log.Printf("[MQTT] 更新通话记录: ID=%s, 状态=%s, 原因=%s", callID, status, reason)
}

// extractDeviceIDFromTopic 从主题中提取设备ID
func extractDeviceIDFromTopic(topic string) string {
	// 针对不同类型的主题进行解析
	if strings.HasPrefix(topic, "calls/request/") {
		parts := strings.Split(topic, "/")
		if len(parts) >= 3 {
			return parts[2]
		}
	} else if strings.HasPrefix(topic, "devices/") {
		parts := strings.Split(topic, "/")
		if len(parts) >= 3 {
			return parts[1]
		}
	}
	return ""
}

// InitiateCallToAll 向设备关联的户号下的所有居民发起通话
func (s *MQTTCallService) InitiateCallToAll(deviceID string) (string, []string, error) {
	// 查询设备信息及其关联的户号
	var device models.Device
	if err := s.DB.Preload("Household.Residents").First(&device, deviceID).Error; err != nil {
		return "", nil, fmt.Errorf("查询设备失败: %v", err)
	}

	// 如果设备没有关联户号，返回错误
	if device.Household == nil || device.HouseholdID == 0 {
		return "", nil, fmt.Errorf("设备未关联户号")
	}

	// 如果户号没有关联居民，返回错误
	if len(device.Household.Residents) == 0 {
		return "", nil, fmt.Errorf("户号未关联任何居民")
	}

	// 生成唯一的通话ID
	callID := uuid.New().String()

	// 收集所有居民ID
	residentIDs := make([]string, 0, len(device.Household.Residents))

	// 向每个居民发送呼叫通知
	for _, resident := range device.Household.Residents {
		residentID := fmt.Sprintf("%d", resident.ID)
		residentIDs = append(residentIDs, residentID)

		// 创建TRTC房间并生成签名
		rtcRoomID, err := s.RTCService.CreateVideoCall(deviceID, residentID)
		if err != nil {
			log.Printf("[MQTT] 为居民 %s 创建TRTC房间失败: %v", residentID, err)
			continue
		}

		// 为住户生成UserSig
		tokenInfo, err := s.RTCService.GetUserSig(residentID)
		if err != nil {
			log.Printf("[MQTT] 为居民 %s 生成UserSig失败: %v", residentID, err)
			continue
		}

		// 创建TRTC信息
		trtcInfo := models.TRTCInfo{
			RoomID:     rtcRoomID,
			RoomIDType: "string",
			SDKAppID:   tokenInfo.SDKAppID,
			UserID:     tokenInfo.UserID,
			UserSig:    tokenInfo.UserSig,
		}

		// 创建通话会话
		_, err = s.CallManager.CreateSession(callID, deviceID, residentID, trtcInfo)
		if err != nil {
			log.Printf("[MQTT] 为居民 %s 创建通话会话失败: %v", residentID, err)
			continue
		}

		// 发送呼入通知给住户
		incomingNotification := IncomingCallNotification{
			CallID:       callID,
			DeviceID:     deviceID,
			TargetUserID: residentID,
			Timestamp:    time.Now().UnixMilli(),
			TRTCInfo: TRTCInfo{
				RoomIDType: trtcInfo.RoomIDType,
				RoomID:     trtcInfo.RoomID,
				SDKAppID:   trtcInfo.SDKAppID,
				UserID:     trtcInfo.UserID,
				UserSig:    trtcInfo.UserSig,
			},
		}

		// 发布到住户的呼入通知主题
		incomingTopic := fmt.Sprintf(TopicIncomingCallFormat, residentID)
		if err := s.publishMessage(incomingTopic, incomingNotification); err != nil {
			log.Printf("[MQTT] 发送呼入通知给居民 %s 失败: %v", residentID, err)
			continue
		}

		// 创建通话记录
		s.createCallRecord(callID, deviceID, residentID, "ringing")
	}

	if len(residentIDs) == 0 {
		return "", nil, fmt.Errorf("没有成功向任何居民发起呼叫")
	}

	// 更新会话状态为振铃中
	s.CallManager.UpdateSessionStatus(callID, "ringing")

	// 发送振铃控制消息给设备
	callerControl := CallInfo{
		Action:    "ringing",
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
	}
	callerTopic := fmt.Sprintf(TopicCallerControlFormat, deviceID)
	if err := s.publishMessage(callerTopic, callerControl); err != nil {
		log.Printf("[MQTT] 发送振铃控制消息失败: %v", err)
	}

	return callID, residentIDs, nil
}

// InitiateCallToHousehold 向指定户号下的所有居民发起通话
func (s *MQTTCallService) InitiateCallToHousehold(deviceID string, householdID string) (string, []string, error) {
	// 查询户号关联的所有居民
	var household models.Household
	if err := s.DB.Preload("Residents").First(&household, householdID).Error; err != nil {
		return "", nil, fmt.Errorf("查询户号失败: %v", err)
	}

	// 如果没有关联的居民，返回错误
	if len(household.Residents) == 0 {
		return "", nil, fmt.Errorf("户号未关联任何居民")
	}

	// 生成唯一的通话ID
	callID := uuid.New().String()

	// 收集所有居民ID
	residentIDs := make([]string, 0, len(household.Residents))

	// 向每个居民发送呼叫通知
	for _, resident := range household.Residents {
		residentID := fmt.Sprintf("%d", resident.ID)
		residentIDs = append(residentIDs, residentID)

		// 创建TRTC房间并生成签名
		rtcRoomID, err := s.RTCService.CreateVideoCall(deviceID, residentID)
		if err != nil {
			log.Printf("[MQTT] 为居民 %s 创建TRTC房间失败: %v", residentID, err)
			continue
		}

		// 为住户生成UserSig
		tokenInfo, err := s.RTCService.GetUserSig(residentID)
		if err != nil {
			log.Printf("[MQTT] 为居民 %s 生成UserSig失败: %v", residentID, err)
			continue
		}

		// 创建TRTC信息
		trtcInfo := models.TRTCInfo{
			RoomID:     rtcRoomID,
			RoomIDType: "string",
			SDKAppID:   tokenInfo.SDKAppID,
			UserID:     tokenInfo.UserID,
			UserSig:    tokenInfo.UserSig,
		}

		// 创建通话会话
		_, err = s.CallManager.CreateSession(callID, deviceID, residentID, trtcInfo)
		if err != nil {
			log.Printf("[MQTT] 为居民 %s 创建通话会话失败: %v", residentID, err)
			continue
		}

		// 发送呼入通知给住户
		incomingNotification := IncomingCallNotification{
			CallID:       callID,
			DeviceID:     deviceID,
			TargetUserID: residentID,
			Timestamp:    time.Now().UnixMilli(),
			TRTCInfo: TRTCInfo{
				RoomIDType: trtcInfo.RoomIDType,
				RoomID:     trtcInfo.RoomID,
				SDKAppID:   trtcInfo.SDKAppID,
				UserID:     trtcInfo.UserID,
				UserSig:    trtcInfo.UserSig,
			},
		}

		// 发布到住户的呼入通知主题
		incomingTopic := fmt.Sprintf(TopicIncomingCallFormat, residentID)
		if err := s.publishMessage(incomingTopic, incomingNotification); err != nil {
			log.Printf("[MQTT] 发送呼入通知给居民 %s 失败: %v", residentID, err)
			continue
		}

		// 创建通话记录
		s.createCallRecord(callID, deviceID, residentID, "ringing")
	}

	if len(residentIDs) == 0 {
		return "", nil, fmt.Errorf("没有成功向任何居民发起呼叫")
	}

	// 更新会话状态为振铃中
	s.CallManager.UpdateSessionStatus(callID, "ringing")

	// 发送振铃控制消息给设备
	callerControl := CallInfo{
		Action:    "ringing",
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
	}
	callerTopic := fmt.Sprintf(TopicCallerControlFormat, deviceID)
	if err := s.publishMessage(callerTopic, callerControl); err != nil {
		log.Printf("[MQTT] 发送振铃控制消息失败: %v", err)
	}

	return callID, residentIDs, nil
}

// InitiateCallByPhone 通过住户电话发起通话
func (s *MQTTCallService) InitiateCallByPhone(deviceID string, phone string) (string, []string, error) {
	// 通过电话号码查询住户
	var resident models.Resident
	if err := s.DB.Where("phone = ?", phone).First(&resident).Error; err != nil {
		return "", nil, fmt.Errorf("未找到电话为 %s 的住户: %v", phone, err)
	}

	// 生成唯一的通话ID
	callID := uuid.New().String()

	// 获取住户ID
	residentID := fmt.Sprintf("%d", resident.ID)

	// 创建TRTC房间并生成签名
	rtcRoomID, err := s.RTCService.CreateVideoCall(deviceID, residentID)
	if err != nil {
		return "", nil, fmt.Errorf("创建TRTC房间失败: %v", err)
	}

	// 为住户生成UserSig
	tokenInfo, err := s.RTCService.GetUserSig(residentID)
	if err != nil {
		return "", nil, fmt.Errorf("生成UserSig失败: %v", err)
	}

	// 创建TRTC信息
	trtcInfo := models.TRTCInfo{
		RoomID:     rtcRoomID,
		RoomIDType: "string",
		SDKAppID:   tokenInfo.SDKAppID,
		UserID:     tokenInfo.UserID,
		UserSig:    tokenInfo.UserSig,
	}

	// 创建通话会话
	_, err = s.CallManager.CreateSession(callID, deviceID, residentID, trtcInfo)
	if err != nil {
		return "", nil, fmt.Errorf("创建通话会话失败: %v", err)
	}

	// 发送呼入通知给住户
	incomingNotification := IncomingCallNotification{
		CallID:       callID,
		DeviceID:     deviceID,
		TargetUserID: residentID,
		Timestamp:    time.Now().UnixMilli(),
		TRTCInfo: TRTCInfo{
			RoomIDType: trtcInfo.RoomIDType,
			RoomID:     trtcInfo.RoomID,
			SDKAppID:   trtcInfo.SDKAppID,
			UserID:     trtcInfo.UserID,
			UserSig:    trtcInfo.UserSig,
		},
	}

	// 发布到住户的呼入通知主题
	incomingTopic := fmt.Sprintf(TopicIncomingCallFormat, residentID)
	if err := s.publishMessage(incomingTopic, incomingNotification); err != nil {
		return "", nil, fmt.Errorf("发送呼入通知失败: %v", err)
	}

	// 更新会话状态为振铃中
	s.CallManager.UpdateSessionStatus(callID, "ringing")

	// 发送振铃控制消息给设备
	callerControl := CallInfo{
		Action:    "ringing",
		CallID:    callID,
		Timestamp: time.Now().UnixMilli(),
	}
	callerTopic := fmt.Sprintf(TopicCallerControlFormat, deviceID)
	if err := s.publishMessage(callerTopic, callerControl); err != nil {
		log.Printf("[MQTT] 发送振铃控制消息失败: %v", err)
	}

	// 创建通话记录
	s.createCallRecord(callID, deviceID, residentID, "ringing")

	return callID, []string{residentID}, nil
}
