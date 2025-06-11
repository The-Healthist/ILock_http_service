package controllers

import (
	"ilock-http-service/internal/error/code"
	"ilock-http-service/internal/error/response"
	"ilock-http-service/internal/domain/models"
	"ilock-http-service/internal/domain/services"
	"ilock-http-service/internal/domain/services/container"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InterfaceBuildingController 定义楼号控制器接口
type InterfaceBuildingController interface {
	GetBuildings()
	GetBuilding()
	CreateBuilding()
	UpdateBuilding()
	DeleteBuilding()
	GetBuildingDevices()
	GetBuildingHouseholds()
}

// BuildingController 处理楼号相关的请求
type BuildingController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewBuildingController 创建一个新的楼号控制器
func NewBuildingController(ctx *gin.Context, container *container.ServiceContainer) *BuildingController {
	return &BuildingController{
		Ctx:       ctx,
		Container: container,
	}
}

// BuildingRequest 表示楼号请求
type BuildingRequest struct {
	BuildingName string `json:"building_name" binding:"required" example:"1号楼"`
	BuildingCode string `json:"building_code" binding:"required" example:"B001"`
	Address      string `json:"address" example:"小区东南角"`
	Status       string `json:"status" example:"active"` // active, inactive
}

// HandleBuildingFunc 返回一个处理楼号请求的Gin处理函数
func HandleBuildingFunc(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewBuildingController(ctx, container)

		switch method {
		case "getBuildings":
			controller.GetBuildings()
		case "getBuilding":
			controller.GetBuilding()
		case "createBuilding":
			controller.CreateBuilding()
		case "updateBuilding":
			controller.UpdateBuilding()
		case "deleteBuilding":
			controller.DeleteBuilding()
		case "getBuildingDevices":
			controller.GetBuildingDevices()
		case "getBuildingHouseholds":
			controller.GetBuildingHouseholds()
		default:
			response.FailWithMessage(ctx, code.ErrBind, "无效的方法", nil)
		}
	}
}

// 1. GetBuildings 获取所有楼号列表
// @Summary 获取所有楼号
// @Description 获取系统中所有楼号的列表
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页条数，默认为10"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /buildings [get]
func (c *BuildingController) GetBuildings() {
	// 获取分页参数
	page, _ := strconv.Atoi(c.Ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Ctx.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取楼号服务
	buildingService := c.Container.GetService("building").(services.InterfaceBuildingService)
	buildings, total, err := buildingService.GetAllBuildings(page, pageSize)
	if err != nil {
		response.FailWithMessage(c.Ctx, code.ErrDatabase, "获取楼号列表失败: "+err.Error(), nil)
		return
	}

	response.Success(c.Ctx, gin.H{
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		"data":        buildings,
	})
}

// 2. GetBuilding 获取单个楼号详情
// @Summary 获取楼号详情
// @Description 根据ID获取楼号详细信息
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "楼号ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id} [get]
func (c *BuildingController) GetBuilding() {
	id := c.Ctx.Param("id")
	buildingID, err := strconv.Atoi(id)
	if err != nil {
		response.ParamError(c.Ctx, "无效的楼号ID")
		return
	}

	// 获取楼号服务
	buildingService := c.Container.GetService("building").(services.InterfaceBuildingService)
	building, err := buildingService.GetBuildingByID(uint(buildingID))
	if err != nil {
		response.NotFound(c.Ctx, "楼号不存在: "+err.Error())
		return
	}

	response.Success(c.Ctx, building)
}

// 3. CreateBuilding 创建新楼号
// @Summary 创建楼号
// @Description 创建一个新的楼号
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param building body BuildingRequest true "楼号信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings [post]
func (c *BuildingController) CreateBuilding() {
	var req BuildingRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c.Ctx, code.ErrBind, "无效的请求参数: "+err.Error(), nil)
		return
	}

	// 创建楼号对象
	building := &models.Building{
		BuildingName: req.BuildingName,
		BuildingCode: req.BuildingCode,
		Address:      req.Address,
	}

	// 如果提供了状态，则设置状态
	if req.Status != "" {
		building.Status = req.Status
	} else {
		building.Status = "active"
	}

	// 获取楼号服务
	buildingService := c.Container.GetService("building").(services.InterfaceBuildingService)
	if err := buildingService.CreateBuilding(building); err != nil {
		response.FailWithMessage(c.Ctx, code.ErrDatabase, "创建楼号失败: "+err.Error(), nil)
		return
	}

	c.Ctx.Status(http.StatusCreated)
	response.Success(c.Ctx, building)
}

// 4. UpdateBuilding 更新楼号信息
// @Summary 更新楼号
// @Description 更新楼号信息
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "楼号ID"
// @Param building body BuildingRequest true "楼号信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id} [put]
func (c *BuildingController) UpdateBuilding() {
	id := c.Ctx.Param("id")
	buildingID, err := strconv.Atoi(id)
	if err != nil {
		response.ParamError(c.Ctx, "无效的楼号ID")
		return
	}

	var req BuildingRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c.Ctx, code.ErrBind, "无效的请求参数: "+err.Error(), nil)
		return
	}

	// 创建更新映射
	updates := make(map[string]interface{})
	if req.BuildingName != "" {
		updates["building_name"] = req.BuildingName
	}
	if req.BuildingCode != "" {
		updates["building_code"] = req.BuildingCode
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	// 获取楼号服务
	buildingService := c.Container.GetService("building").(services.InterfaceBuildingService)
	building, err := buildingService.UpdateBuilding(uint(buildingID), updates)
	if err != nil {
		response.FailWithMessage(c.Ctx, code.ErrDatabase, "更新楼号失败: "+err.Error(), nil)
		return
	}

	response.Success(c.Ctx, building)
}

// 5. DeleteBuilding 删除楼号
// @Summary 删除楼号
// @Description 删除指定的楼号
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "楼号ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id} [delete]
func (c *BuildingController) DeleteBuilding() {
	id := c.Ctx.Param("id")
	buildingID, err := strconv.Atoi(id)
	if err != nil {
		response.ParamError(c.Ctx, "无效的楼号ID")
		return
	}

	// 获取楼号服务
	buildingService := c.Container.GetService("building").(services.InterfaceBuildingService)
	if err := buildingService.DeleteBuilding(uint(buildingID)); err != nil {
		response.FailWithMessage(c.Ctx, code.ErrDatabase, "删除楼号失败: "+err.Error(), nil)
		return
	}

	response.Success(c.Ctx, nil)
}

// 6. GetBuildingDevices 获取楼号关联的设备
// @Summary 获取楼号关联的设备
// @Description 获取指定楼号关联的所有设备
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "楼号ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id}/devices [get]
func (c *BuildingController) GetBuildingDevices() {
	id := c.Ctx.Param("id")
	buildingID, err := strconv.Atoi(id)
	if err != nil {
		response.ParamError(c.Ctx, "无效的楼号ID")
		return
	}

	// 获取楼号服务
	buildingService := c.Container.GetService("building").(services.InterfaceBuildingService)
	devices, err := buildingService.GetBuildingDevices(uint(buildingID))
	if err != nil {
		response.FailWithMessage(c.Ctx, code.ErrDatabase, "获取楼号关联设备失败: "+err.Error(), nil)
		return
	}

	response.Success(c.Ctx, devices)
}

// 7. GetBuildingHouseholds 获取楼号下的户号
// @Summary 获取楼号下的户号
// @Description 获取指定楼号下的所有户号
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "楼号ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /buildings/{id}/households [get]
func (c *BuildingController) GetBuildingHouseholds() {
	id := c.Ctx.Param("id")
	buildingID, err := strconv.Atoi(id)
	if err != nil {
		response.ParamError(c.Ctx, "无效的楼号ID")
		return
	}

	// 获取楼号服务
	buildingService := c.Container.GetService("building").(services.InterfaceBuildingService)
	households, err := buildingService.GetBuildingHouseholds(uint(buildingID))
	if err != nil {
		response.FailWithMessage(c.Ctx, code.ErrDatabase, "获取楼号下户号失败: "+err.Error(), nil)
		return
	}

	response.Success(c.Ctx, households)
}
