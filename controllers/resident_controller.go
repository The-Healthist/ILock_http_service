package controllers

import (
	"ilock-http-service/models"
	"ilock-http-service/services"
	"ilock-http-service/services/container"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InterfaceResidentController 定义居民控制器接口
type InterfaceResidentController interface {
	GetResidents()
	GetResident()
	CreateResident()
	UpdateResident()
	DeleteResident()
}

// ResidentController 处理居民相关的请求
type ResidentController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewResidentController 创建一个新的居民控制器
func NewResidentController(ctx *gin.Context, container *container.ServiceContainer) *ResidentController {
	return &ResidentController{
		Ctx:       ctx,
		Container: container,
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
	// 获取分页参数
	page, _ := strconv.Atoi(c.Ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Ctx.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 使用 ResidentService 获取居民列表
	residentService := c.Container.GetService("resident").(services.InterfaceResidentService)
	residents, total, err := residentService.GetAllResidents(page, pageSize)
	if err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取居民列表失败",
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
			"data":        residents,
		},
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
	id := c.Ctx.Param("id")
	if id == "" {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "居民ID不能为空",
			"data":    nil,
		})
		return
	}

	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	// 使用 ResidentService 获取居民详情
	residentService := c.Container.GetService("resident").(services.InterfaceResidentService)
	resident, err := residentService.GetResidentByID(uint(idUint))
	if err != nil {
		if err.Error() == "居民不存在" {
			c.Ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取居民信息失败",
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
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
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"data":    nil,
		})
		return
	}

	// 创建居民对象
	resident := &models.Resident{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		DeviceID: req.DeviceID,
		// 密码将在 ResidentService 中处理
	}

	// 使用 ResidentService 创建居民
	residentService := c.Container.GetService("resident").(services.InterfaceResidentService)
	if err := residentService.CreateResident(resident); err != nil {
		if err.Error() == "手机号已被使用" || err.Error() == "设备不存在" {
			c.Ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建居民失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusCreated, gin.H{
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
	id := c.Ctx.Param("id")
	if id == "" {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	var req UpdateResidentRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的请求参数",
			"data":    nil,
		})
		return
	}

	// 构建更新字段映射
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.DeviceID != 0 {
		updates["device_id"] = req.DeviceID
	}

	// 使用 ResidentService 更新居民
	residentService := c.Container.GetService("resident").(services.InterfaceResidentService)
	resident, err := residentService.UpdateResident(uint(idUint), updates)
	if err != nil {
		if err.Error() == "居民不存在" {
			c.Ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}
		if err.Error() == "手机号已被使用" || err.Error() == "设备不存在" {
			c.Ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新居民失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
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
	id := c.Ctx.Param("id")
	if id == "" {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的居民ID",
			"data":    nil,
		})
		return
	}

	// 使用 ResidentService 删除居民
	residentService := c.Container.GetService("resident").(services.InterfaceResidentService)
	if err := residentService.DeleteResident(uint(idUint)); err != nil {
		if err.Error() == "居民不存在" {
			c.Ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除居民失败: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "居民删除成功",
		"data":    nil,
	})
}

// HandleResidentFunc 返回一个处理居民请求的Gin处理函数
func HandleResidentFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewResidentController(ctx, container)

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
