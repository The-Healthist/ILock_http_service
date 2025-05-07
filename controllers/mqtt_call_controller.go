package controllers

import (
	"ilock-http-service/services"
	"ilock-http-service/services/container"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// InterfaceMQTTCallController 定义MQTT通话控制器接口
type InterfaceMQTTCallController interface {
	InitiateCall()
	CallerAction()
	CalleeAction()
	GetCallSession()
	EndCallSession()
	PublishDeviceStatus()
	PublishSystemMessage()
}

// MQTTCallController MQTT通话控制器实现
type MQTTCallController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewMQTTCallController 创建一个新的MQTT通话控制器
func NewMQTTCallController(ctx *gin.Context, container *container.ServiceContainer) InterfaceMQTTCallController {
	return &MQTTCallController{
		Ctx:       ctx,
		Container: container,
	}
}

// 请求结构体定义
type (
	// InitiateCallRequest 发起通话请求
	InitiateCallRequest struct {
		DeviceID     string `json:"device_device_id" binding:"required"`   // 使用与MQTT通讯中相同的字段名
		TargetUserID string `json:"target_resident_id" binding:"required"` // 使用与MQTT通讯中相同的字段名
	}

	// CallActionRequest 通话控制请求
	CallActionRequest struct {
		CallID string `json:"call_id" binding:"required"`
		Action string `json:"action" binding:"required"`
		Reason string `json:"reason,omitempty"`
	}

	// GetCallSessionRequest 获取通话会话请求
	GetCallSessionRequest struct {
		CallID string `json:"call_id" binding:"required"`
	}

	// EndCallSessionRequest 结束通话会话请求
	EndCallSessionRequest struct {
		CallID string `json:"call_id" binding:"required"`
		Reason string `json:"reason,omitempty"`
	}

	// PublishDeviceStatusRequest 发布设备状态请求
	PublishDeviceStatusRequest struct {
		DeviceID   string                 `json:"device_id" binding:"required"`
		Online     bool                   `json:"online"`
		Battery    int                    `json:"battery"`
		Properties map[string]interface{} `json:"properties,omitempty"`
	}

	// PublishSystemMessageRequest 发布系统消息请求
	PublishSystemMessageRequest struct {
		Type    string                 `json:"type" binding:"required"`
		Level   string                 `json:"level" binding:"required"`
		Message string                 `json:"message" binding:"required"`
		Data    map[string]interface{} `json:"data,omitempty"`
	}

	// CallSessionResponse 通话会话响应
	CallSessionResponse struct {
		CallID       string    `json:"call_id"`
		DeviceID     string    `json:"device_id"`
		ResidentID   string    `json:"resident_id"`
		StartTime    time.Time `json:"start_time"`
		Status       string    `json:"status"`
		LastActivity time.Time `json:"last_activity"`
	}
)

// HandleMQTTCallFunc 返回一个处理MQTT通话请求的Gin处理函数
func HandleMQTTCallFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewMQTTCallController(ctx, container)

		switch method {
		case "initiateCall":
			controller.InitiateCall()
		case "callerAction":
			controller.CallerAction()
		case "calleeAction":
			controller.CalleeAction()
		case "getCallSession":
			controller.GetCallSession()
		case "endCallSession":
			controller.EndCallSession()
		case "publishDeviceStatus":
			controller.PublishDeviceStatus()
		case "publishSystemMessage":
			controller.PublishSystemMessage()
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的方法",
				"data":    nil,
			})
		}
	}
}

// 1. InitiateCall 发起通话
// @Summary      Initiate MQTT Call
// @Description  Start a new call via MQTT using improved session management
// @Tags         MQTT Call
// @Accept       json
// @Produce      json
// @Param        request body InitiateCallRequest true "Call initiation request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/mqtt/initiate [post]
func (c *MQTTCallController) InitiateCall() {
	var req InitiateCallRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.HandleError(http.StatusBadRequest, "无效的请求参数", err)
		return
	}

	mqttCallService := c.Container.GetService("mqtt_call").(services.InterfaceMQTTCallService)
	callID, err := mqttCallService.InitiateCall(req.DeviceID, req.TargetUserID)
	if err != nil {
		c.HandleError(http.StatusInternalServerError, "发起通话失败", err)
		return
	}

	c.HandleSuccess(gin.H{
		"call_id": callID,
	})
}

// 2. CallerAction 处理呼叫方动作
// @Summary      Handle MQTT Caller Action
// @Description  Process caller's actions using improved session management
// @Tags         MQTT Call
// @Accept       json
// @Produce      json
// @Param        request body CallActionRequest true "Call action request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/mqtt/caller-action [post]
func (c *MQTTCallController) CallerAction() {
	var req CallActionRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.HandleError(http.StatusBadRequest, "无效的请求参数", err)
		return
	}

	// 验证动作类型
	validActions := map[string]bool{"hangup": true, "cancelled": true}
	if !validActions[req.Action] {
		c.HandleError(http.StatusBadRequest, "不支持的动作类型", nil)
		return
	}

	mqttCallService := c.Container.GetService("mqtt_call").(services.InterfaceMQTTCallService)
	if err := mqttCallService.HandleCallerAction(req.CallID, req.Action, req.Reason); err != nil {
		c.HandleError(http.StatusInternalServerError, "处理呼叫方动作失败", err)
		return
	}

	c.HandleSuccess(nil)
}

// 3. CalleeAction 处理被呼叫方动作
// @Summary      Handle MQTT Callee Action
// @Description  Process callee's actions using improved session management
// @Tags         MQTT Call
// @Accept       json
// @Produce      json
// @Param        request body CallActionRequest true "Call action request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/mqtt/callee-action [post]
func (c *MQTTCallController) CalleeAction() {
	var req CallActionRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.HandleError(http.StatusBadRequest, "无效的请求参数", err)
		return
	}

	// 验证动作类型
	validActions := map[string]bool{"rejected": true, "answered": true, "hangup": true, "timeout": true}
	if !validActions[req.Action] {
		c.HandleError(http.StatusBadRequest, "不支持的动作类型", nil)
		return
	}

	mqttCallService := c.Container.GetService("mqtt_call").(services.InterfaceMQTTCallService)
	if err := mqttCallService.HandleCalleeAction(req.CallID, req.Action, req.Reason); err != nil {
		c.HandleError(http.StatusInternalServerError, "处理被呼叫方动作失败", err)
		return
	}

	c.HandleSuccess(nil)
}

// 4. GetCallSession 获取通话会话
// @Summary      Get MQTT Call Session
// @Description  Retrieve call session information
// @Tags         MQTT Call
// @Accept       json
// @Produce      json
// @Param        request body GetCallSessionRequest true "Get call session request"
// @Success      200  {object}  CallSessionResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/mqtt/session [get]
func (c *MQTTCallController) GetCallSession() {
	var req GetCallSessionRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.HandleError(http.StatusBadRequest, "无效的请求参数", err)
		return
	}

	mqttCallService := c.Container.GetService("mqtt_call").(services.InterfaceMQTTCallService)
	session, exists := mqttCallService.GetCallSession(req.CallID)
	if !exists {
		c.HandleError(http.StatusNotFound, "通话会话不存在", nil)
		return
	}

	// 创建响应对象
	response := CallSessionResponse{
		CallID:       session.CallID,
		DeviceID:     session.DeviceID,
		ResidentID:   session.ResidentID,
		StartTime:    session.StartTime,
		Status:       session.Status,
		LastActivity: session.LastActivity,
	}

	c.HandleSuccess(response)
}

// 5. EndCallSession 结束通话会话
// @Summary      End MQTT Call Session
// @Description  Forcefully end a call session
// @Tags         MQTT Call
// @Accept       json
// @Produce      json
// @Param        request body EndCallSessionRequest true "End call session request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/mqtt/end-session [post]
func (c *MQTTCallController) EndCallSession() {
	var req EndCallSessionRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.HandleError(http.StatusBadRequest, "无效的请求参数", err)
		return
	}

	mqttCallService := c.Container.GetService("mqtt_call").(services.InterfaceMQTTCallService)
	if err := mqttCallService.EndCallSession(req.CallID, req.Reason); err != nil {
		c.HandleError(http.StatusInternalServerError, "结束通话会话失败", err)
		return
	}

	c.HandleSuccess(nil)
}

// 6. PublishDeviceStatus 发布设备状态
// @Summary      Publish Device Status
// @Description  Publish device status information via MQTT using improved session management
// @Tags         MQTT
// @Accept       json
// @Produce      json
// @Param        request body PublishDeviceStatusRequest true "Device status information"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/mqtt/device/status [post]
func (c *MQTTCallController) PublishDeviceStatus() {
	var req PublishDeviceStatusRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.HandleError(http.StatusBadRequest, "无效的请求参数", err)
		return
	}

	mqttCallService := c.Container.GetService("mqtt_call").(services.InterfaceMQTTCallService)
	status := map[string]interface{}{
		"device_id":   req.DeviceID,
		"online":      req.Online,
		"battery":     req.Battery,
		"properties":  req.Properties,
		"last_update": time.Now().UnixMilli(),
	}

	if err := mqttCallService.PublishDeviceStatus(req.DeviceID, status); err != nil {
		c.HandleError(http.StatusInternalServerError, "发布设备状态失败", err)
		return
	}

	c.HandleSuccess(nil)
}

// 7. PublishSystemMessage 发布系统消息
// @Summary      Publish System Message
// @Description  Publish system message via MQTT using improved session management
// @Tags         MQTT
// @Accept       json
// @Produce      json
// @Param        request body PublishSystemMessageRequest true "System message information"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/mqtt/system/message [post]
func (c *MQTTCallController) PublishSystemMessage() {
	var req PublishSystemMessageRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.HandleError(http.StatusBadRequest, "无效的请求参数", err)
		return
	}

	// 验证消息级别
	validLevels := map[string]bool{"info": true, "warning": true, "error": true}
	if !validLevels[req.Level] {
		c.HandleError(http.StatusBadRequest, "无效的消息级别", nil)
		return
	}

	mqttCallService := c.Container.GetService("mqtt_call").(services.InterfaceMQTTCallService)
	message := map[string]interface{}{
		"type":      req.Type,
		"level":     req.Level,
		"message":   req.Message,
		"data":      req.Data,
		"timestamp": time.Now().UnixMilli(),
	}

	if err := mqttCallService.PublishSystemMessage(req.Type, message); err != nil {
		c.HandleError(http.StatusInternalServerError, "发布系统消息失败", err)
		return
	}

	c.HandleSuccess(nil)
}

// HandleSuccess 处理成功响应
func (c *MQTTCallController) HandleSuccess(data interface{}) {
	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    data,
	})
}

// HandleError 处理错误响应
func (c *MQTTCallController) HandleError(status int, message string, err error) {
	errMessage := message
	if err != nil {
		errMessage = message + ": " + err.Error()
	}

	c.Ctx.JSON(status, gin.H{
		"code":    status,
		"message": errMessage,
		"data":    nil,
	})
}
