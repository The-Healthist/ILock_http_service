package controllers

import (
	"ilock-http-service/services"
	"ilock-http-service/services/container"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InterfaceCallRecordController 定义通话记录控制器接口
type InterfaceCallRecordController interface {
	GetCallRecords()
	GetCallRecord()
	GetCallStatistics()
	GetDeviceCallRecords()
	GetResidentCallRecords()
	SubmitCallFeedback()
}

// CallRecordController 处理通话记录相关的请求
type CallRecordController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewCallRecordController 创建一个新的通话记录控制器
func NewCallRecordController(ctx *gin.Context, container *container.ServiceContainer) *CallRecordController {
	return &CallRecordController{
		Ctx:       ctx,
		Container: container,
	}
}

// CallFeedbackRequest 表示通话质量反馈请求
type CallFeedbackRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5" example:"4"` // 1-5 星评分
	Comment string `json:"comment" example:"通话质量良好，声音清晰"`                     // 可选评论
	Issues  string `json:"issues" example:"偶尔有一点延迟"`                          // 问题描述
}

// HandleCallRecordFunc 返回一个处理通话记录请求的Gin处理函数
func HandleCallRecordFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewCallRecordController(ctx, container)

		switch method {
		case "getCallRecords":
			controller.GetCallRecords()
		case "getCallRecord":
			controller.GetCallRecord()
		case "getCallStatistics":
			controller.GetCallStatistics()
		case "getDeviceCallRecords":
			controller.GetDeviceCallRecords()
		case "getResidentCallRecords":
			controller.GetResidentCallRecords()
		case "submitCallFeedback":
			controller.SubmitCallFeedback()
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的方法",
				"data":    nil,
			})
		}
	}
}

// 1. GetCallRecords 获取通话记录列表
// @Summary      Get Call Records
// @Description  Get a list of all call records in the system, with pagination
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number, default is 1" example:"1"
// @Param        page_size query int false "Items per page, default is 10" example:"10"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  ErrorResponse
// @Router       /calls [get]
func (c *CallRecordController) GetCallRecords() {
	// 获取分页参数
	page, _ := strconv.Atoi(c.Ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Ctx.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	callRecordService := c.Container.GetService("call_record").(services.InterfaceCallRecordService)

	calls, total, err := callRecordService.GetAllCallRecords(page, pageSize)
	if err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取通话记录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"records":     calls,
		},
	})
}

// 2. GetCallRecord 获取单个通话记录
// @Summary      Get Call Record By ID
// @Description  Get details of a specific call record by ID
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Call Record ID" example:"1"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/{id} [get]
func (c *CallRecordController) GetCallRecord() {
	id := c.Ctx.Param("id")
	recordID, err := strconv.Atoi(id)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的通话记录ID",
			"data":    nil,
		})
		return
	}

	callRecordService := c.Container.GetService("call_record").(services.InterfaceCallRecordService)

	record, err := callRecordService.GetCallRecordByID(uint(recordID))
	if err != nil {
		c.Ctx.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    record,
	})
}

// 3. GetCallStatistics 获取通话统计信息
// @Summary      Get Call Statistics
// @Description  Get call statistics including total, answered, missed, etc.
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/statistics [get]
func (c *CallRecordController) GetCallStatistics() {
	callRecordService := c.Container.GetService("call_record").(services.InterfaceCallRecordService)

	statistics, err := callRecordService.GetCallStatistics()
	if err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取通话统计信息失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    statistics,
	})
}

// 4. GetDeviceCallRecords 获取指定设备的通话记录
// @Summary      Get Device Call Records
// @Description  Get call records for a specific device, with pagination
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        deviceId path int true "Device ID" example:"1"
// @Param        page query int false "Page number, default is 1" example:"1"
// @Param        page_size query int false "Items per page, default is 10" example:"10"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/device/{deviceId} [get]
func (c *CallRecordController) GetDeviceCallRecords() {
	deviceID := c.Ctx.Param("deviceId")
	id, err := strconv.Atoi(deviceID)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
			"data":    nil,
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.Ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Ctx.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	callRecordService := c.Container.GetService("call_record").(services.InterfaceCallRecordService)

	calls, total, err := callRecordService.GetCallRecordsByDeviceID(uint(id), page, pageSize)
	if err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取设备通话记录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"records":     calls,
		},
	})
}

// 5. GetResidentCallRecords 获取指定居民的通话记录
// @Summary      Get Resident Call Records
// @Description  Get call records for a specific resident, with pagination
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        residentId path int true "Resident ID" example:"1"
// @Param        page query int false "Page number, default is 1" example:"1"
// @Param        page_size query int false "Items per page, default is 10" example:"10"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/resident/{residentId} [get]
func (c *CallRecordController) GetResidentCallRecords() {
	residentID := c.Ctx.Param("residentId")
	id, err := strconv.Atoi(residentID)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.Ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Ctx.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	callRecordService := c.Container.GetService("call_record").(services.InterfaceCallRecordService)

	calls, total, err := callRecordService.GetCallRecordsByResidentID(uint(id), page, pageSize)
	if err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取居民通话记录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"records":     calls,
		},
	})
}

// 6. SubmitCallFeedback 提交通话质量反馈
// @Summary      Submit Call Feedback
// @Description  Submit quality feedback for a specific call record
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Call Record ID" example:"1"
// @Param        request body CallFeedbackRequest true "Feedback information"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/{id}/feedback [post]
func (c *CallRecordController) SubmitCallFeedback() {
	callID := c.Ctx.Param("id")
	id, err := strconv.Atoi(callID)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的通话记录ID",
			"data":    nil,
		})
		return
	}

	var req CallFeedbackRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
			"data":    nil,
		})
		return
	}

	callRecordService := c.Container.GetService("call_record").(services.InterfaceCallRecordService)

	feedback := &services.CallFeedback{
		CallID:  uint(id),
		Rating:  req.Rating,
		Comment: req.Comment,
		Issues:  req.Issues,
	}

	if err := callRecordService.SubmitCallFeedback(feedback); err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交反馈失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功提交反馈",
		"data":    nil,
	})
}
