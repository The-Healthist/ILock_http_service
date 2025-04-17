package controllers

import (
	"ilock-http-service/models"
	"ilock-http-service/services/container"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StaffController 处理物业员工相关的请求
type StaffController struct {
	BaseControllerImpl
}

// NewStaffController 创建一个新的物业员工控制器
func (f *ControllerFactory) NewStaffController(ctx *gin.Context) *StaffController {
	return &StaffController{
		BaseControllerImpl: BaseControllerImpl{
			Container: f.Container,
			Context:   ctx,
		},
	}
}

// GetStaffs 获取物业员工列表
// @Summary      Get Property Staff List
// @Description  Get a list of all property staff members, with pagination and search support
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number, default is 1" example:"1"
// @Param        page_size query int false "Items per page, default is 10" example:"10"
// @Param        search query string false "Search keyword for name, phone, etc." example:"manager"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  ErrorResponse
// @Router       /staffs [get]
// @Security     BearerAuth
func (c *StaffController) GetStaffs() {
	// 获取分页参数
	page, _ := strconv.Atoi(c.Context.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Context.DefaultQuery("page_size", "10"))
	search := c.Context.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 查询数据库
	var staffs []models.PropertyStaff
	var total int64

	db := c.Container.GetDB()
	query := db.Model(&models.PropertyStaff{})

	// 如果有搜索关键词，添加搜索条件
	if search != "" {
		query = query.Where("phone LIKE ? OR property_name LIKE ?",
			"%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	result := query.Count(&total)
	if result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询物业员工数量失败",
			"data":    nil,
		})
		return
	}

	// 获取分页数据
	result = query.Limit(pageSize).Offset(offset).Find(&staffs)
	if result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询物业员工列表失败",
			"data":    nil,
		})
		return
	}

	// 构建响应
	var staffResponses []gin.H
	for _, staff := range staffs {
		staffResponses = append(staffResponses, gin.H{
			"id":            staff.ID,
			"phone":         staff.Phone,
			"property_name": staff.PropertyName,
			"position":      staff.Position,
			"role":          staff.Role,
			"status":        staff.Status,
			"created_at":    staff.CreatedAt,
			"updated_at":    staff.UpdatedAt,
		})
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"data":        staffResponses,
		},
	})
}

// GetStaff 获取单个物业员工详情
// @Summary      Get Property Staff By ID
// @Description  Get details of a specific property staff member by ID
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Param        id path int true "Property Staff ID" example:"1"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /staffs/{id} [get]
// @Security     BearerAuth
func (c *StaffController) GetStaff() {
	// 获取URL参数中的ID
	idStr := c.Context.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID参数",
			"data":    nil,
		})
		return
	}

	// 查询数据库
	var staff models.PropertyStaff
	db := c.Container.GetDB()
	result := db.First(&staff, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.Context.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "物业员工不存在",
				"data":    nil,
			})
		} else {
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询物业员工失败: " + result.Error.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 查询关联的设备
	var devices []models.Device
	if err := db.Model(&staff).Association("Devices").Find(&devices); err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询关联设备失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 提取设备ID
	var deviceIDs []uint
	for _, device := range devices {
		deviceIDs = append(deviceIDs, device.ID)
	}

	// 返回物业员工信息
	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"id":            staff.ID,
			"phone":         staff.Phone,
			"property_name": staff.PropertyName,
			"position":      staff.Position,
			"role":          staff.Role,
			"status":        staff.Status,
			"username":      staff.Username,
			"remark":        staff.Remark,
			"device_ids":    deviceIDs,
			"created_at":    staff.CreatedAt,
			"updated_at":    staff.UpdatedAt,
		},
	})
}

// CreateStaffRequest 表示创建物业员工的请求体
type CreateStaffRequest struct {
	Name         string `json:"name" example:"王物业"` // 注意: 已从模型中移除，但保留请求结构以兼容前端
	Phone        string `json:"phone" binding:"required" example:"13700001234"`
	PropertyName string `json:"property_name" example:"阳光花园小区"`
	Position     string `json:"position" example:"物业经理"`
	Role         string `json:"role" binding:"required" example:"manager"` // 可选值: manager, staff, security
	Status       string `json:"status" example:"active"`                   // 可选值: active, inactive, suspended
	Remark       string `json:"remark" example:"负责A区日常管理工作"`
	Username     string `json:"username" binding:"required" example:"wangwuye"`
	Password     string `json:"password" binding:"required" example:"Property@123"`
	DeviceIDs    []uint `json:"device_ids" example:"[1,2,3]"` // 关联的设备ID列表
}

// CreateStaff 添加新物业员工
// @Summary      Create Property Staff
// @Description  Create a new property staff account, with specified role, position and property
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Param        request body CreateStaffRequest true "Property staff information - including name, phone, username, password, property ID and role"
// @Success      201  {object}  map[string]interface{} "Success response with created staff details"
// @Failure      400  {object}  ErrorResponse "Bad request, phone or username already in use"
// @Failure      500  {object}  ErrorResponse "Server error"
// @Router       /staffs [post]
// @Security     BearerAuth
func (c *StaffController) CreateStaff() {
	var req CreateStaffRequest
	if err := c.Context.ShouldBindJSON(&req); err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 检查手机号是否已存在
	db := c.Container.GetDB()
	var count int64
	db.Model(&models.PropertyStaff{}).Where("phone = ?", req.Phone).Count(&count)
	if count > 0 {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "手机号已被注册",
			"data":    nil,
		})
		return
	}

	// 检查用户名是否已存在
	db.Model(&models.PropertyStaff{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户名已存在",
			"data":    nil,
		})
		return
	}

	// 创建新物业员工
	staff := models.PropertyStaff{
		Phone:        req.Phone,
		PropertyName: req.PropertyName,
		Position:     req.Position,
		Role:         req.Role,
		Status:       req.Status,
		Remark:       req.Remark,
		Username:     req.Username,
		Password:     req.Password,
	}

	// 开始事务
	tx := db.Begin()

	// 保存物业员工
	if err := tx.Create(&staff).Error; err != nil {
		tx.Rollback()
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建物业员工失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 关联设备
	if len(req.DeviceIDs) > 0 {
		// 查询所有关联的设备
		var devices []models.Device
		if err := tx.Where("id IN ?", req.DeviceIDs).Find(&devices).Error; err != nil {
			tx.Rollback()
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询关联设备失败: " + err.Error(),
				"data":    nil,
			})
			return
		}

		// 创建关联关系
		if err := tx.Model(&staff).Association("Devices").Append(&devices); err != nil {
			tx.Rollback()
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "关联设备失败: " + err.Error(),
				"data":    nil,
			})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "成功创建物业员工",
		"data": gin.H{
			"id":         staff.ID,
			"phone":      staff.Phone,
			"username":   staff.Username,
			"created_at": staff.CreatedAt,
			"device_ids": req.DeviceIDs,
		},
	})
}

// UpdateStaffRequest 表示更新物业员工的请求体
type UpdateStaffRequest struct {
	Name         string `json:"name" example:"李物业"`
	Phone        string `json:"phone" example:"13700005678"`
	PropertyName string `json:"property_name" example:"幸福家园小区"`
	Position     string `json:"position" example:"前台客服"`
	Role         string `json:"role" example:"staff"`    // 可选值: manager, staff, security
	Status       string `json:"status" example:"active"` // 可选值: active, inactive, suspended
	Remark       string `json:"remark" example:"负责接待访客和处理居民投诉"`
	Username     string `json:"username" example:"liwuye"`
	Password     string `json:"password" example:"NewProperty@456"`
	DeviceIDs    []uint `json:"device_ids" example:"[1,3,5]"` // 更新关联的设备ID列表
}

// UpdateStaff 更新物业员工信息
// @Summary      Update Property Staff
// @Description  Update details of a property staff member with the specified ID
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Param        id path int true "Property Staff ID" example:"1"
// @Param        request body UpdateStaffRequest true "Updated property staff information"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /staffs/{id} [put]
// @Security     BearerAuth
func (c *StaffController) UpdateStaff() {
	// 获取URL参数中的ID
	idStr := c.Context.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID参数",
			"data":    nil,
		})
		return
	}

	var req UpdateStaffRequest
	if err := c.Context.ShouldBindJSON(&req); err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 查询数据库
	db := c.Container.GetDB()
	var staff models.PropertyStaff
	result := db.First(&staff, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.Context.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "物业员工不存在",
				"data":    nil,
			})
		} else {
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询物业员工失败: " + result.Error.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 更新字段
	updateMap := make(map[string]interface{})

	if req.Name != "" {
		updateMap["name"] = req.Name
	}

	if req.Phone != "" && req.Phone != staff.Phone {
		// 检查手机号是否已被其他用户使用
		var count int64
		db.Model(&models.PropertyStaff{}).Where("phone = ? AND id != ?", req.Phone, id).Count(&count)
		if count > 0 {
			c.Context.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "手机号已被其他物业员工使用",
				"data":    nil,
			})
			return
		}
		updateMap["phone"] = req.Phone
	}

	if req.PropertyName != "" {
		updateMap["property_name"] = req.PropertyName
	}

	if req.Position != "" {
		updateMap["position"] = req.Position
	}

	if req.Role != "" {
		updateMap["role"] = req.Role
	}

	if req.Status != "" {
		updateMap["status"] = req.Status
	}

	// Remark可以为空字符串，所以需要特殊处理
	if req.Remark != staff.Remark {
		updateMap["remark"] = req.Remark
	}

	if req.Username != "" && req.Username != staff.Username {
		// 检查用户名是否已被其他用户使用
		var count int64
		db.Model(&models.PropertyStaff{}).Where("username = ? AND id != ?", req.Username, id).Count(&count)
		if count > 0 {
			c.Context.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "用户名已被其他物业员工使用",
				"data":    nil,
			})
			return
		}
		updateMap["username"] = req.Username
	}

	if req.Password != "" {
		updateMap["password"] = req.Password
	}

	// 开始事务
	tx := db.Begin()

	// 更新数据库
	if len(updateMap) > 0 {
		result = tx.Model(&staff).Updates(updateMap)
		if result.Error != nil {
			tx.Rollback()
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新物业员工失败: " + result.Error.Error(),
				"data":    nil,
			})
			return
		}
	}

	// 如果提供了设备ID列表，更新关联
	var deviceIDs []uint
	if req.DeviceIDs != nil {
		// 先删除所有现有关联
		if err := tx.Model(&staff).Association("Devices").Clear(); err != nil {
			tx.Rollback()
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "清除设备关联失败: " + err.Error(),
				"data":    nil,
			})
			return
		}

		// 如果有新的设备ID，添加新关联
		if len(req.DeviceIDs) > 0 {
			var devices []models.Device
			if err := tx.Where("id IN ?", req.DeviceIDs).Find(&devices).Error; err != nil {
				tx.Rollback()
				c.Context.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "查询关联设备失败: " + err.Error(),
					"data":    nil,
				})
				return
			}

			if err := tx.Model(&staff).Association("Devices").Append(&devices); err != nil {
				tx.Rollback()
				c.Context.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "关联设备失败: " + err.Error(),
					"data":    nil,
				})
				return
			}

			deviceIDs = req.DeviceIDs
		}
	} else {
		// 查询当前关联的设备ID
		var devices []models.Device
		if err := tx.Model(&staff).Association("Devices").Find(&devices); err != nil {
			tx.Rollback()
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询关联设备失败: " + err.Error(),
				"data":    nil,
			})
			return
		}

		for _, device := range devices {
			deviceIDs = append(deviceIDs, device.ID)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 重新获取更新后的记录
	db.First(&staff, id)

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功更新物业员工",
		"data": gin.H{
			"id":         staff.ID,
			"phone":      staff.Phone,
			"username":   staff.Username,
			"updated_at": staff.UpdatedAt,
			"device_ids": deviceIDs,
		},
	})
}

// DeleteStaff 删除物业员工
// @Summary      Delete Property Staff
// @Description  Delete a property staff member with the specified ID
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Param        id path int true "Property Staff ID" example:"2"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /staffs/{id} [delete]
// @Security     BearerAuth
func (c *StaffController) DeleteStaff() {
	// 获取URL参数中的ID
	idStr := c.Context.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID参数",
			"data":    nil,
		})
		return
	}

	// 查询数据库
	db := c.Container.GetDB()
	var staff models.PropertyStaff
	result := db.First(&staff, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.Context.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "物业员工不存在",
				"data":    nil,
			})
		} else {
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询物业员工失败: " + result.Error.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 删除物业员工
	result = db.Delete(&staff)
	if result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除物业员工失败: " + result.Error.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功删除物业员工",
		"data":    nil,
	})
}

// HandleStaffFunc 返回一个处理物业员工请求的Gin处理函数
func HandleStaffFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	factory := NewControllerFactory(container)

	return func(ctx *gin.Context) {
		controller := factory.NewStaffController(ctx)

		switch method {
		case "getStaffs":
			controller.GetStaffs()
		case "getStaff":
			controller.GetStaff()
		case "createStaff":
			controller.CreateStaff()
		case "updateStaff":
			controller.UpdateStaff()
		case "deleteStaff":
			controller.DeleteStaff()
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的方法",
				"data":    nil,
			})
		}
	}
}
