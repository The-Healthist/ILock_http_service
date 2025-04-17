package controllers

import (
	"ilock-http-service/models"
	"ilock-http-service/services/container"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ResidentController 处理居民相关的请求
type ResidentController struct {
	BaseControllerImpl
}

// NewResidentController 创建一个新的居民控制器
func (f *ControllerFactory) NewResidentController(ctx *gin.Context) *ResidentController {
	return &ResidentController{
		BaseControllerImpl: BaseControllerImpl{
			Container: f.Container,
			Context:   ctx,
		},
	}
}

// ResidentRequest 表示居民请求
type ResidentRequest struct {
	Name     string `json:"name" binding:"required" example:"张三"`
	Email    string `json:"email" binding:"omitempty,email" example:"zhangsan@resident.com"`
	Phone    string `json:"phone" binding:"required" example:"13812345678"`
	DeviceID uint   `json:"device_id" binding:"required" example:"101"`
}

// UpdateResidentRequest 表示更新居民请求
type UpdateResidentRequest struct {
	Name     string `json:"name" example:"李四"`
	Email    string `json:"email" binding:"omitempty,email" example:"lisi@resident.com"`
	Phone    string `json:"phone" example:"13987654321"`
	DeviceID uint   `json:"device_id" example:"102"`
}

// GetResidents 获取所有居民
// @Summary      Get Resident List
// @Description  Get a list of all residents in the system
// @Tags         Resident
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /residents [get]
func (c *ResidentController) GetResidents() {
	var residents []models.Resident

	// 使用 Container 获取数据库连接
	db := c.Container.GetDB()

	result := db.Find(&residents)
	if result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取居民列表失败",
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    residents,
	})
}

// GetResident 获取单个居民
// @Summary      Get Resident By ID
// @Description  Get details of a specific resident by ID
// @Tags         Resident
// @Accept       json
// @Produce      json
// @Param        id path int true "Resident ID" example:"1"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /residents/{id} [get]
func (c *ResidentController) GetResident() {
	id := c.Context.Param("id")
	if id == "" {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "居民ID不能为空",
			"data":    nil,
		})
		return
	}

	var resident models.Resident

	// 使用 Container 获取数据库连接
	db := c.Container.GetDB()

	result := db.First(&resident, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.Context.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "居民不存在",
				"data":    nil,
			})
			return
		}
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取居民信息失败",
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    resident,
	})
}

// CreateResident 创建新居民
// @Summary      Create Resident
// @Description  Create a new resident account, requiring association with a specific device
// @Tags         Resident
// @Accept       json
// @Produce      json
// @Param        request body ResidentRequest true "Resident information - name, phone and device ID are required, email is optional"
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{} "Success response with created resident details"
// @Failure      400  {object}  ErrorResponse "Bad request, device not found or phone number already in use"
// @Failure      500  {object}  ErrorResponse "Server error"
// @Router       /residents [post]
func (c *ResidentController) CreateResident() {
	var req ResidentRequest
	if err := c.Context.ShouldBindJSON(&req); err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"data":    nil,
		})
		return
	}

	// 使用 Container 获取数据库连接
	db := c.Container.GetDB()

	// 检查手机号是否已存在
	var existingResident models.Resident
	if err := db.Where("phone = ?", req.Phone).First(&existingResident).Error; err == nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "手机号已被使用",
			"data":    nil,
		})
		return
	}

	// 检查设备是否存在
	var device models.Device
	if err := db.First(&device, req.DeviceID).Error; err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "设备不存在",
			"data":    nil,
		})
		return
	}

	// 创建新居民
	resident := models.Resident{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		DeviceID: req.DeviceID,
	}

	if result := db.Create(&resident); result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建居民失败",
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "居民创建成功",
		"data":    resident,
	})
}

// UpdateResident 更新居民信息
// @Summary      Update Resident
// @Description  Update details of a resident with the specified ID
// @Tags         Resident
// @Accept       json
// @Produce      json
// @Param        id path int true "Resident ID" example:"1"
// @Param        request body UpdateResidentRequest true "Updated resident information"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /residents/{id} [put]
func (c *ResidentController) UpdateResident() {
	id := c.Context.Param("id")
	if id == "" {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	_, err := strconv.Atoi(id)
	if err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	var req UpdateResidentRequest
	if err := c.Context.ShouldBindJSON(&req); err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"data":    nil,
		})
		return
	}

	// 使用 Container 获取数据库连接
	db := c.Container.GetDB()

	// 查找居民
	var resident models.Resident
	if err := db.First(&resident, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Context.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "居民不存在",
				"data":    nil,
			})
			return
		}
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取居民信息失败",
			"data":    nil,
		})
		return
	}

	// 验证更新字段

	// 检查手机号是否已被其他居民使用
	if req.Phone != "" && req.Phone != resident.Phone {
		var count int64
		db.Model(&models.Resident{}).Where("phone = ? AND id != ?", req.Phone, id).Count(&count)
		if count > 0 {
			c.Context.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "手机号已被使用",
				"data":    nil,
			})
			return
		}
	}

	// 更新可选字段
	if req.Name != "" {
		resident.Name = req.Name
	}
	if req.Email != "" {
		resident.Email = req.Email
	}
	if req.Phone != "" {
		resident.Phone = req.Phone
	}
	if req.DeviceID != 0 {
		// 检查设备是否存在
		var device models.Device
		if err := db.First(&device, req.DeviceID).Error; err != nil {
			c.Context.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "设备不存在",
				"data":    nil,
			})
			return
		}
		resident.DeviceID = req.DeviceID
	}

	// 保存更新
	if result := db.Save(&resident); result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新居民失败",
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "居民更新成功",
		"data":    resident,
	})
}

// DeleteResident 删除居民
// @Summary      Delete Resident
// @Description  Delete a resident with the specified ID
// @Tags         Resident
// @Accept       json
// @Produce      json
// @Param        id path int true "Resident ID" example:"2"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /residents/{id} [delete]
func (c *ResidentController) DeleteResident() {
	id := c.Context.Param("id")
	if id == "" {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	_, err := strconv.Atoi(id)
	if err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	// 使用 Container 获取数据库连接
	db := c.Container.GetDB()

	// 查找居民
	var resident models.Resident
	if err := db.First(&resident, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Context.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "居民不存在",
				"data":    nil,
			})
			return
		}
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取居民信息失败",
			"data":    nil,
		})
		return
	}

	// 删除居民
	if result := db.Delete(&resident); result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除居民失败",
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "居民删除成功",
		"data":    nil,
	})
}

// HandleResidentFunc 返回一个处理居民请求的Gin处理函数
func HandleResidentFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	factory := NewControllerFactory(container)

	return func(ctx *gin.Context) {
		controller := factory.NewResidentController(ctx)

		switch method {
		case "getResidents":
			controller.GetResidents()
		case "getResident":
			controller.GetResident()
		case "createResident":
			controller.CreateResident()
		case "updateResident":
			controller.UpdateResident()
		case "deleteResident":
			controller.DeleteResident()
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的方法",
				"data":    nil,
			})
		}
	}
}
