package controllers

import (
	"ilock-http-service/models"
	"ilock-http-service/services/container"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminController 处理管理员相关的请求
type AdminController struct {
	BaseControllerImpl
}

// NewAdminController 创建一个新的管理员控制器
func (f *ControllerFactory) NewAdminController(ctx *gin.Context) *AdminController {
	return &AdminController{
		BaseControllerImpl: BaseControllerImpl{
			Container: f.Container,
			Context:   ctx,
		},
	}
}

// GetAdmins 获取所有管理员
// @Summary      Get Admin List
// @Description  Get a list of all administrators in the system, with pagination support
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number, default is 1" example:"1"
// @Param        page_size query int false "Items per page, default is 10" example:"10"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /admins [get]
func (c *AdminController) GetAdmins() {
	// 获取分页参数
	page, _ := strconv.Atoi(c.Context.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Context.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 查询数据库
	var admins []models.Admin
	var total int64

	db := c.Container.GetDB()
	result := db.Model(&models.Admin{}).Count(&total)
	if result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询管理员数量失败",
			"data":    nil,
		})
		return
	}

	result = db.Limit(pageSize).Offset(offset).Find(&admins)
	if result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询管理员列表失败",
			"data":    nil,
		})
		return
	}

	// 处理敏感信息
	var adminResponses []gin.H
	for _, admin := range admins {
		adminResponses = append(adminResponses, gin.H{
			"id":         admin.ID,
			"username":   admin.Username,
			"email":      admin.Email,
			"created_at": admin.CreatedAt,
			"updated_at": admin.UpdatedAt,
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
			"data":        adminResponses,
		},
	})
}

// GetAdmin 获取单个管理员
// @Summary      Get Admin By ID
// @Description  Get details of a specific administrator by ID
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        id path int true "Administrator ID" example:"1"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /admins/{id} [get]
func (c *AdminController) GetAdmin() {
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
	var admin models.Admin
	db := c.Container.GetDB()
	result := db.First(&admin, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.Context.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "管理员不存在",
				"data":    nil,
			})
		} else {
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询管理员失败: " + result.Error.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 处理敏感信息
	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"id":         admin.ID,
			"username":   admin.Username,
			"email":      admin.Email,
			"created_at": admin.CreatedAt,
			"updated_at": admin.UpdatedAt,
		},
	})
}

// CreateAdminRequest 表示创建管理员的请求体
type CreateAdminRequest struct {
	Username string `json:"username" binding:"required" example:"admin1"`
	Password string `json:"password" binding:"required" example:"admin123"`
	Email    string `json:"email" binding:"required,email" example:"admin@ilock.com"`
}

// CreateAdmin 创建新管理员
// @Summary      Create Administrator
// @Description  Create a new system administrator account and return the created administrator info
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        request body CreateAdminRequest true "Administrator information - username, password and email are required"
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{} "Success response with created admin details"
// @Failure      400  {object}  ErrorResponse "Bad request or username already exists"
// @Failure      500  {object}  ErrorResponse "Server error"
// @Router       /admins [post]
func (c *AdminController) CreateAdmin() {
	var req CreateAdminRequest
	if err := c.Context.ShouldBindJSON(&req); err != nil {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// 检查用户名是否已存在
	db := c.Container.GetDB()
	var count int64
	db.Model(&models.Admin{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户名已存在",
			"data":    nil,
		})
		return
	}

	// 创建新管理员
	admin := models.Admin{
		Username: req.Username,
		Email:    req.Email,
	}

	// 设置密码（使用哈希）
	hashedPassword, err := models.HashPassword(req.Password)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
			"data":    nil,
		})
		return
	}
	admin.Password = hashedPassword

	// 保存到数据库
	result := db.Create(&admin)
	if result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建管理员失败: " + result.Error.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "成功创建管理员",
		"data": gin.H{
			"id":         admin.ID,
			"username":   admin.Username,
			"email":      admin.Email,
			"created_at": admin.CreatedAt,
		},
	})
}

// UpdateAdminRequest 表示更新管理员的请求体
type UpdateAdminRequest struct {
	Username string `json:"username" example:"admin_updated"`
	Password string `json:"password" example:"NewPassword@123"`
	Email    string `json:"email" binding:"omitempty,email" example:"admin@ilock.com"`
}

// UpdateAdmin 更新管理员信息
// @Summary      Update Administrator
// @Description  Update details of an administrator with the specified ID
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        id path int true "Administrator ID" example:"1"
// @Param        request body UpdateAdminRequest true "Updated administrator information"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /admins/{id} [put]
func (c *AdminController) UpdateAdmin() {
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

	var req UpdateAdminRequest
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
	var admin models.Admin
	result := db.First(&admin, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.Context.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "管理员不存在",
				"data":    nil,
			})
		} else {
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询管理员失败: " + result.Error.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 更新字段
	updateMap := make(map[string]interface{})

	if req.Username != "" {
		// 检查用户名是否已被其他用户使用
		var count int64
		db.Model(&models.Admin{}).Where("username = ? AND id != ?", req.Username, id).Count(&count)
		if count > 0 {
			c.Context.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "用户名已被其他用户使用",
				"data":    nil,
			})
			return
		}
		updateMap["username"] = req.Username
	}

	if req.Email != "" {
		updateMap["email"] = req.Email
	}

	if req.Password != "" {
		hashedPassword, err := models.HashPassword(req.Password)
		if err != nil {
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "密码加密失败",
				"data":    nil,
			})
			return
		}
		updateMap["password"] = hashedPassword
	}

	// 更新数据库
	if len(updateMap) > 0 {
		result = db.Model(&admin).Updates(updateMap)
		if result.Error != nil {
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "更新管理员失败: " + result.Error.Error(),
				"data":    nil,
			})
			return
		}
	}

	// 重新获取更新后的记录
	db.First(&admin, id)

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功更新管理员",
		"data": gin.H{
			"id":         admin.ID,
			"username":   admin.Username,
			"email":      admin.Email,
			"updated_at": admin.UpdatedAt,
		},
	})
}

// DeleteAdmin 删除管理员
// @Summary      Delete Administrator
// @Description  Delete an administrator with the specified ID
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        id path int true "Administrator ID" example:"2"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /admins/{id} [delete]
func (c *AdminController) DeleteAdmin() {
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
	var admin models.Admin
	result := db.First(&admin, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.Context.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "管理员不存在",
				"data":    nil,
			})
		} else {
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询管理员失败: " + result.Error.Error(),
				"data":    nil,
			})
		}
		return
	}

	// 确保系统中至少有一个管理员
	var count int64
	db.Model(&models.Admin{}).Count(&count)
	if count <= 1 {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "系统必须至少有一个管理员，无法删除最后一个管理员",
			"data":    nil,
		})
		return
	}

	// 删除管理员
	result = db.Delete(&admin)
	if result.Error != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除管理员失败: " + result.Error.Error(),
			"data":    nil,
		})
		return
	}

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功删除管理员",
		"data":    nil,
	})
}

// HandleAdminFunc 返回一个处理管理员请求的Gin处理函数
func HandleAdminFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	factory := NewControllerFactory(container)

	return func(ctx *gin.Context) {
		controller := factory.NewAdminController(ctx)

		switch method {
		case "getAdmins":
			controller.GetAdmins()
		case "getAdmin":
			controller.GetAdmin()
		case "createAdmin":
			controller.CreateAdmin()
		case "updateAdmin":
			controller.UpdateAdmin()
		case "deleteAdmin":
			controller.DeleteAdmin()
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的方法",
				"data":    nil,
			})
		}
	}
}
