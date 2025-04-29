package controllers

import (
	"ilock-http-service/models"
	"ilock-http-service/services"
	"ilock-http-service/services/container"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InterfaceDeviceController 定义设备控制器接口
type InterfaceDeviceController interface {
	GetDevices()
	GetDevice()
	CreateDevice()
	UpdateDevice()
	DeleteDevice()
	GetDeviceStatus()
}

// DeviceController 处理设备相关的请求
type DeviceController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewDeviceController 创建一个新的设备控制器
func NewDeviceController(ctx *gin.Context, container *container.ServiceContainer) *DeviceController {
	return &DeviceController{
		Ctx:       ctx,
		Container: container,
	}
}

// DeviceRequest 表示设备请求
type DeviceRequest struct {
	Name         string `json:"name" binding:"required" example:"门禁1号"`
	SerialNumber string `json:"serial_number" binding:"required" example:"SN2024050001"`
	Status       string `json:"status" example:"online"` // online, offline, fault
	Location     string `json:"location" example:"小区北门入口"`
	StaffIDs     []uint `json:"staff_ids" example:"[1,2]"` // 关联的物业员工ID列表
}

// HandleDeviceFunc 返回一个处理设备请求的Gin处理函数
func HandleDeviceFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewDeviceController(ctx, container)

		switch method {
		case "getDevices":
			controller.GetDevices()
		case "getDevice":
			controller.GetDevice()
		case "createDevice":
			controller.CreateDevice()
		case "updateDevice":
			controller.UpdateDevice()
		case "deleteDevice":
			controller.DeleteDevice()
		case "getDeviceStatus":
			controller.GetDeviceStatus()
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的方法",
				"data":    nil,
			})
		}
	}
}

// 1. GetDevices 获取所有设备列表
// @Summary 获取所有设备
// @Description 获取所有设备的列表
// @Tags device
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Device
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /devices [get]
func (c *DeviceController) GetDevices() {
	deviceService := c.Container.GetService("device").(services.InterfaceDeviceService)

	devices, err := deviceService.GetAllDevices()
	if err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取设备列表失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    devices,
	})
}

// 2. GetDevice 获取单个设备详情
// @Summary 获取单个设备
// @Description 根据ID获取设备信息
// @Tags device
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "设备ID"
// @Success 200 {object} models.Device
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /devices/{id} [get]
func (c *DeviceController) GetDevice() {
	id := c.Ctx.Param("id")
	deviceID, err := strconv.Atoi(id)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
			"data":    nil,
		})
		return
	}

	deviceService := c.Container.GetService("device").(services.InterfaceDeviceService)

	device, err := deviceService.GetDeviceByID(uint(deviceID))
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
		"data":    device,
	})
}

// 3. CreateDevice 创建新设备
// @Summary 创建新设备
// @Description 创建一个新的门禁设备
// @Tags device
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param device body DeviceRequest true "设备信息"
// @Success 201 {object} models.Device
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /devices [post]
func (c *DeviceController) CreateDevice() {
	var req DeviceRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
			"data":    nil,
		})
		return
	}

	device := &models.Device{
		Name:         req.Name,
		SerialNumber: req.SerialNumber,
		Location:     req.Location,
	}

	// 如果提供了状态，则设置状态
	if req.Status != "" {
		switch req.Status {
		case "online":
			device.Status = models.DeviceStatusOnline
		case "offline":
			device.Status = models.DeviceStatusOffline
		case "fault":
			device.Status = models.DeviceStatusFault
		default:
			device.Status = models.DeviceStatusOffline
		}
	} else {
		device.Status = models.DeviceStatusOffline
	}

	deviceService := c.Container.GetService("device").(services.InterfaceDeviceService)

	if err := deviceService.CreateDevice(device); err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建设备失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "成功",
		"data":    device,
	})
}

// 4. UpdateDevice 更新设备信息
// @Summary 更新设备信息
// @Description 根据ID更新设备信息
// @Tags device
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "设备ID"
// @Param device body DeviceRequest true "设备信息"
// @Success 200 {object} models.Device
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /devices/{id} [put]
func (c *DeviceController) UpdateDevice() {
	id := c.Ctx.Param("id")
	deviceID, err := strconv.Atoi(id)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
			"data":    nil,
		})
		return
	}

	var req DeviceRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 创建更新映射
	updates := make(map[string]interface{})
	updates["name"] = req.Name
	updates["serial_number"] = req.SerialNumber
	updates["location"] = req.Location

	// 处理状态更新
	if req.Status != "" {
		switch req.Status {
		case "online":
			updates["status"] = models.DeviceStatusOnline
		case "offline":
			updates["status"] = models.DeviceStatusOffline
		case "fault":
			updates["status"] = models.DeviceStatusFault
		}
	}

	deviceService := c.Container.GetService("device").(services.InterfaceDeviceService)

	device, err := deviceService.UpdateDevice(uint(deviceID), updates)
	if err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新设备失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    device,
	})
}

// 5. DeleteDevice 删除设备
// @Summary 删除设备
// @Description 根据ID删除设备
// @Tags device
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "设备ID"
// @Success 204 {object} nil
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /devices/{id} [delete]
func (c *DeviceController) DeleteDevice() {
	id := c.Ctx.Param("id")
	deviceID, err := strconv.Atoi(id)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
			"data":    nil,
		})
		return
	}

	deviceService := c.Container.GetService("device").(services.InterfaceDeviceService)

	if err := deviceService.DeleteDevice(uint(deviceID)); err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除设备失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    nil,
	})
}

// 6. GetDeviceStatus 获取设备状态
// @Summary      Get Device Status
// @Description  Get the current status of a device by ID
// @Tags         device
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Device ID" example:"1"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /devices/{id}/status [get]
func (c *DeviceController) GetDeviceStatus() {
	// 获取URL参数中的ID
	idStr := c.Ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID参数",
			"data":    nil,
		})
		return
	}

	// 查询数据库
	var device models.Device
	db := c.Container.GetService("db").(*gorm.DB)
	result := db.First(&device, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.Ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "设备未找到",
				"data":    nil,
			})
		} else {
			c.Ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询设备失败: " + result.Error.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 返回设备状态信息
	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"id":            device.ID,
			"name":          device.Name,
			"serial_number": device.SerialNumber,
			"status":        device.Status,
			"location":      device.Location,
			"last_online":   device.UpdatedAt,
		},
	})
}
