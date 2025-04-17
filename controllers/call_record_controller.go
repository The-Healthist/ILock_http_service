package controllers

import (
	"ilock-http-service/services"
	"ilock-http-service/services/container"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CallRecordController 处理通话记录相关的请求
type CallRecordController struct {
	BaseControllerImpl
}

// NewCallRecordController 创建一个新的通话记录控制器
func (f *ControllerFactory) NewCallRecordController(ctx *gin.Context) *CallRecordController {
	return &CallRecordController{
		BaseControllerImpl: BaseControllerImpl{
			Container: f.Container,
			Context:   ctx,
		},
	}
}

// CallFeedbackRequest 表示通话质量反馈请求
type CallFeedbackRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5" example:"4"` // 1-5 星评分
	Comment string `json:"comment" example:"通话质量良好，声音清晰"`                     // 可选评论
	Issues  string `json:"issues" example:"偶尔有一点延迟"`                          // 问题描述
}

// GetCallRecords 获取通话记录列表
// @Summary      Get Call Records
// @Description  Get a list of all call records in the system, with pagination
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number, default is 1" example:"1"
// @Param        page_size query int false "Items per page, default is 10" example:"10"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  ErrorResponse
// @Router       /calls [get]
func (c *CallRecordController) GetCallRecords() {
	// 获取分页参数
	page, _ := strconv.Atoi(c.Context.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Context.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	callRecordService := c.Container.GetCallRecordService()

	calls, total, err := callRecordService.GetAllCallRecords(page, pageSize)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取通话记录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
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

// GetCallRecord 获取单个通话记录
// @Summary      Get Call Record By ID
// @Description  Get details of a specific call record by ID
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Param        id path int true "Call Record ID" example:"1"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/{id} [get]
func (c *CallRecordController) GetCallRecord() {
	id := c.Context.Param("id")
	recordID, err := strconv.Atoi(id)
	if err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的通话记录ID",
			"data":    nil,
		})
		return
	}

	callRecordService := c.Container.GetCallRecordService()

	record, err := callRecordService.GetCallRecordByID(uint(recordID))
	if err != nil {
		c.Context.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    record,
	})
}

// GetCallStatistics 获取通话统计信息
// @Summary      Get Call Statistics
// @Description  Get call statistics including total, answered, missed, etc.
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/statistics [get]
func (c *CallRecordController) GetCallStatistics() {
	callRecordService := c.Container.GetCallRecordService()

	statistics, err := callRecordService.GetCallStatistics()
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取通话统计信息失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    statistics,
	})
}

// GetDeviceCallRecords 获取指定设备的通话记录
// @Summary      Get Device Call Records
// @Description  Get call records for a specific device, with pagination
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Param        deviceId path int true "Device ID" example:"1"
// @Param        page query int false "Page number, default is 1" example:"1"
// @Param        page_size query int false "Items per page, default is 10" example:"10"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/device/{deviceId} [get]
func (c *CallRecordController) GetDeviceCallRecords() {
	deviceID := c.Context.Param("deviceId")
	id, err := strconv.Atoi(deviceID)
	if err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
			"data":    nil,
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.Context.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Context.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	callRecordService := c.Container.GetCallRecordService()

	calls, total, err := callRecordService.GetCallRecordsByDeviceID(uint(id), page, pageSize)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取设备通话记录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
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

// GetResidentCallRecords 获取指定居民的通话记录
// @Summary      Get Resident Call Records
// @Description  Get call records for a specific resident, with pagination
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Param        residentId path int true "Resident ID" example:"1"
// @Param        page query int false "Page number, default is 1" example:"1"
// @Param        page_size query int false "Items per page, default is 10" example:"10"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/resident/{residentId} [get]
func (c *CallRecordController) GetResidentCallRecords() {
	residentID := c.Context.Param("residentId")
	id, err := strconv.Atoi(residentID)
	if err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.Context.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Context.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	callRecordService := c.Container.GetCallRecordService()

	calls, total, err := callRecordService.GetCallRecordsByResidentID(uint(id), page, pageSize)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取居民通话记录失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
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

// SubmitCallFeedback 提交通话质量反馈
// @Summary      Submit Call Feedback
// @Description  Submit quality feedback for a specific call record
// @Tags         CallRecord
// @Accept       json
// @Produce      json
// @Param        id path int true "Call Record ID" example:"1"
// @Param        request body CallFeedbackRequest true "Feedback information"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /calls/{id}/feedback [post]
func (c *CallRecordController) SubmitCallFeedback() {
	callID := c.Context.Param("id")
	id, err := strconv.Atoi(callID)
	if err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的通话记录ID",
			"data":    nil,
		})
		return
	}

	var req CallFeedbackRequest
	if err := c.Context.ShouldBindJSON(&req); err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
			"data":    nil,
		})
		return
	}

	callRecordService := c.Container.GetCallRecordService()

	feedback := &services.CallFeedback{
		CallID:  uint(id),
		Rating:  req.Rating,
		Comment: req.Comment,
		Issues:  req.Issues,
	}

	if err := callRecordService.SubmitCallFeedback(feedback); err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交反馈失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功提交反馈",
		"data":    nil,
	})
}

// HandleCallRecordFunc 返回一个处理通话记录请求的Gin处理函数
func HandleCallRecordFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	factory := NewControllerFactory(container)

	return func(ctx *gin.Context) {
		controller := factory.NewCallRecordController(ctx)

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
