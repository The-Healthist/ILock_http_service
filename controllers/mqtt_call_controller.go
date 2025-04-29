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
	PublishDeviceStatus()
	PublishSystemMessage()
}

// MQTTCallController MQTT通话控制器
type MQTTCallController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewMQTTCallController 创建一个新的MQTT通话控制器
func NewMQTTCallController(ctx *gin.Context, container *container.ServiceContainer) *MQTTCallController {
	return &MQTTCallController{
		Ctx:       ctx,
		Container: container,
	}
}

// 请求结构体定义
type (
	// InitiateCallRequest 发起通话请求
	InitiateCallRequest struct {
		CallerID string `json:"caller_id" binding:"required"`
		CalleeID string `json:"callee_id" binding:"required"`
	}

	// CallActionRequest 通话控制请求
	CallActionRequest struct {
		CallID string `json:"call_id" binding:"required"`
		Action string `json:"action" binding:"required"`
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
// @Description  Start a new call via MQTT
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
	callID, err := mqttCallService.InitiateCall(req.CallerID, req.CalleeID)
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
// @Description  Process caller's cancel or hangup actions
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

	if req.Action != "cancelled" && req.Action != "hangup" {
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
// @Description  Process callee's reject, hangup or timeout actions
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

	if req.Action != "rejected" && req.Action != "hangup" && req.Action != "timeout" {
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

// 4. PublishDeviceStatus 发布设备状态
// @Summary      Publish Device Status
// @Description  Publish device status information via MQTT
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

// 5. PublishSystemMessage 发布系统消息
// @Summary      Publish System Message
// @Description  Publish system message via MQTT
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
		"timestamp": time.Now().UnixMilli(),
		"data":      req.Data,
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
		"code":    0,
		"message": "成功",
		"data":    data,
	})
}

// HandleError 处理错误响应
func (c *MQTTCallController) HandleError(status int, message string, err error) {
	errMsg := message
	if err != nil {
		errMsg = message + ": " + err.Error()
	}
	c.Ctx.JSON(status, gin.H{
		"code":    status,
		"message": errMsg,
		"data":    nil,
	})
}
