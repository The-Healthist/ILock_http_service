package controllers

import (
	"fmt"
	"ilock-http-service/services/container"
	"ilock-http-service/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RTCController 处理RTC相关的请求
type RTCController struct {
	BaseControllerImpl
}

// NewRTCController 创建一个新的RTC控制器
func (f *ControllerFactory) NewRTCController(ctx *gin.Context) *RTCController {
	return &RTCController{
		BaseControllerImpl: BaseControllerImpl{
			Container: f.Container,
			Context:   ctx,
		},
	}
}

// TokenRequest 表示RTC令牌请求
type TokenRequest struct {
	ChannelID string `json:"channel_id" binding:"required" example:"room123"`
	UserID    string `json:"user_id" binding:"required" example:"user456"`
}

// GetToken 处理获取RTC令牌的请求
// @Summary      获取RTC令牌
// @Description  获取用于进行实时通信的RTC令牌
// @Tags         RTC
// @Accept       json
// @Produce      json
// @Param        request body TokenRequest true "令牌请求参数"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /rtc/token [post]
func (c *RTCController) GetToken() {
	var req TokenRequest
	if err := c.Context.ShouldBindJSON(&req); err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"data":    nil,
		})
		return
	}

	// 获取服务
	rtcService := c.Container.GetRTCService()

	// 生成新令牌
	tokenInfo, err := rtcService.GetToken(req.ChannelID, req.UserID)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成令牌失败",
			"data":    nil,
		})
		return
	}

	// 构建与示例格式完全匹配的响应
	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"token":       tokenInfo.Token,
			"channel_id":  tokenInfo.ChannelID,
			"user_id":     tokenInfo.UserID,
			"expire_time": tokenInfo.ExpireTime.Format(time.RFC3339),
			"rtc_app_id":  tokenInfo.AppID, // 添加RTC应用ID
		},
	})
}

// CallRequest 表示发起通话的请求
type CallRequest struct {
	DeviceID   string `json:"device_id" binding:"required"`
	ResidentID string `json:"resident_id" binding:"required"`
}

// StartCall 处理发起通话的请求
// @Summary      发起视频通话
// @Description  在设备和居民之间发起视频通话
// @Tags         RTC
// @Accept       json
// @Produce      json
// @Param        request body CallRequest true "通话请求参数"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /rtc/call [post]
func (c *RTCController) StartCall() {
	var req CallRequest
	if err := c.Context.ShouldBindJSON(&req); err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"data":    nil,
		})
		return
	}

	rtcService := c.Container.GetRTCService()

	// 创建通话通道
	channelID, err := rtcService.CreateVideoCall(req.DeviceID, req.ResidentID)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "发起通话失败",
			"data":    nil,
		})
		return
	}

	// 为双方生成令牌
	deviceToken, err := rtcService.GetToken(channelID, req.DeviceID)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成设备令牌失败",
			"data":    nil,
		})
		return
	}

	residentToken, err := rtcService.GetToken(channelID, req.ResidentID)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成居民令牌失败",
			"data":    nil,
		})
		return
	}

	// 生成设备的用户名（游客+6位随机数）
	deviceUsername := fmt.Sprintf("游客%06d", utils.RandomInt32()%1000000)

	// 这里应该从数据库查询居民的用户名
	// 暂时使用residentID作为用户名
	residentUsername := req.ResidentID

	// TODO: 从数据库查询真实的居民用户名
	// db := c.Container.GetDB()
	// var resident models.Resident
	// if err := db.Where("id = ?", req.ResidentID).First(&resident).Error; err == nil {
	//     residentUsername = resident.Username
	// }

	// 构建新的标准响应格式
	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"channel_id": channelID,
			"rtc_app_id": deviceToken.AppID, // 添加RTC应用ID
			"device": gin.H{
				"id":          req.DeviceID,
				"token":       deviceToken.Token,
				"expire_time": deviceToken.ExpireTime.Format(time.RFC3339),
				"username":    deviceUsername,
			},
			"resident": gin.H{
				"id":          req.ResidentID,
				"token":       residentToken.Token,
				"expire_time": residentToken.ExpireTime.Format(time.RFC3339),
				"username":    residentUsername,
			},
		},
	})
}

// HandleRTCFunc 返回一个处理RTC请求的Gin处理函数
func HandleRTCFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	factory := NewControllerFactory(container)

	return func(ctx *gin.Context) {
		controller := factory.NewRTCController(ctx)

		switch method {
		case "getToken":
			controller.GetToken()
		case "startCall":
			controller.StartCall()
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的方法",
				"data":    nil,
			})
		}
	}
}
