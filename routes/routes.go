package routes

import (
	"ilock-http-service/config"
	"ilock-http-service/controllers"
	_ "ilock-http-service/docs" // 导入 Swagger 文档，这行很重要！
	"ilock-http-service/middleware"
	"ilock-http-service/services/container"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// SetupRouter 初始化并返回配置好的路由
func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	// 初始化 Gin
	r := gin.Default()

	// 设置正确的Content-Type，确保UTF-8编码
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")

		// 自定义JSON响应编码
		c.Next()
	})

	// 创建服务容器
	serviceContainer := container.NewServiceContainer(db, cfg, nil)

	// 初始化中间件
	middleware.InitAuthMiddleware(cfg)

	// 添加 Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册路由
	registerRoutes(r, db, serviceContainer)
	return r
}

// registerRoutes 配置所有API路由
func registerRoutes(
	r *gin.Engine,
	db *gorm.DB,
	container *container.ServiceContainer,
) {
	// API 路由根路径
	api := r.Group("/api")
	// 注册公共路由
	registerPublicRoutes(api, db, container)
	// 注册需要认证的路由
	registerAuthenticatedRoutes(api, db, container)
}

// registerPublicRoutes 注册公共路由
func registerPublicRoutes(
	api *gin.RouterGroup,
	db *gorm.DB,
	container *container.ServiceContainer,
) {
	// 健康检查
	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 认证路由
	api.POST("/auth/login", func(c *gin.Context) {
		auth := controllers.NewAuthController(db, container.GetJWTService())
		auth.Login(c)
	})

	// RTC 路由
	api.POST("/rtc/token", controllers.HandleRTCFunc(container, "getToken"))
	api.POST("/rtc/call", controllers.HandleRTCFunc(container, "startCall"))

	// Weather 路由
	api.GET("/weather", controllers.HandleWeatherFunc(container, db, "getWeather"))
	api.GET("/weather/device/:deviceId", controllers.HandleWeatherFunc(container, db, "getDeviceWeather"))
	api.GET("/weather/forecast", controllers.HandleWeatherFunc(container, db, "get15DaysForecast"))
}

// registerAuthenticatedRoutes 注册需要认证的路由
func registerAuthenticatedRoutes(
	api *gin.RouterGroup,
	db *gorm.DB,
	container *container.ServiceContainer,
) {
	// 系统管理员路由组
	adminGroup := api.Group("")
	adminGroup.Use(middleware.AuthenticateSystemAdmin())

	// 管理员API
	adminRoutes := adminGroup.Group("/admins")
	{
		adminRoutes.GET("", controllers.HandleAdminFunc(container, "getAdmins"))
		adminRoutes.GET("/:id", controllers.HandleAdminFunc(container, "getAdmin"))
		adminRoutes.POST("", controllers.HandleAdminFunc(container, "createAdmin"))
		adminRoutes.PUT("/:id", controllers.HandleAdminFunc(container, "updateAdmin"))
		adminRoutes.DELETE("/:id", controllers.HandleAdminFunc(container, "deleteAdmin"))
	}

	// 物业人员API (管理员权限)
	adminStaffRoutes := adminGroup.Group("/staffs")
	{
		adminStaffRoutes.GET("", controllers.HandleStaffFunc(container, "getStaffs"))
		adminStaffRoutes.GET("/:id", controllers.HandleStaffFunc(container, "getStaff"))
		adminStaffRoutes.POST("", controllers.HandleStaffFunc(container, "createStaff"))
		adminStaffRoutes.PUT("/:id", controllers.HandleStaffFunc(container, "updateStaff"))
		adminStaffRoutes.DELETE("/:id", controllers.HandleStaffFunc(container, "deleteStaff"))
	}

	// 居民API (管理员权限)
	adminResidentRoutes := adminGroup.Group("/residents")
	{
		adminResidentRoutes.GET("", controllers.HandleResidentFunc(container, "getResidents"))
		adminResidentRoutes.GET("/:id", controllers.HandleResidentFunc(container, "getResident"))
		adminResidentRoutes.POST("", controllers.HandleResidentFunc(container, "createResident"))
		adminResidentRoutes.PUT("/:id", controllers.HandleResidentFunc(container, "updateResident"))
		adminResidentRoutes.DELETE("/:id", controllers.HandleResidentFunc(container, "deleteResident"))
	}

	// 设备API (管理员权限)
	adminDeviceRoutes := adminGroup.Group("/devices")
	{
		adminDeviceRoutes.GET("", controllers.HandleDeviceFunc(container, "getDevices"))
		adminDeviceRoutes.GET("/:id", controllers.HandleDeviceFunc(container, "getDevice"))
		adminDeviceRoutes.POST("", controllers.HandleDeviceFunc(container, "createDevice"))
		adminDeviceRoutes.PUT("/:id", controllers.HandleDeviceFunc(container, "updateDevice"))
		adminDeviceRoutes.DELETE("/:id", controllers.HandleDeviceFunc(container, "deleteDevice"))
		adminDeviceRoutes.GET("/:id/status", controllers.HandleDeviceFunc(container, "getDeviceStatus"))
	}

	// 物业人员路由
	// staffRoutes := api.Group("/staff")
	// staffRoutes.Use(middleware.AuthenticatePropertyStaff())
	{
		// 设备硬件相关接口
		// deviceRoutes.GET("/:id/status", controllers.HandleDeviceFunc(container, "getDeviceStatus"))
		// TODO: 以下接口需要硬件集成，后续实现
		// PUT /api/device/{id}/configuration - 更新设备配置
		// POST /api/device/{id}/reboot - 重启设备
		// POST /api/device/{id}/unlock - 远程开门
	}
	// 居民路由，可以被系统管理员和物业人员访问
	// residentRoutes := api.Group("/residents")
	// residentRoutes.Use(middleware.AuthenticatePropertyStaff()) // 物业人员及以上权限可以访问
	// {
	// }
	// 通话记录路由
	callRoutes := api.Group("/calls")
	callRoutes.Use(middleware.Authentication()) // 需要认证才能访问
	{
		callRoutes.GET("", controllers.HandleCallRecordFunc(container, "getCallRecords"))
		callRoutes.GET("/:id", controllers.HandleCallRecordFunc(container, "getCallRecord"))
		callRoutes.GET("/statistics", controllers.HandleCallRecordFunc(container, "getCallStatistics"))
		callRoutes.GET("/device/:deviceId", controllers.HandleCallRecordFunc(container, "getDeviceCallRecords"))
		callRoutes.GET("/resident/:residentId", controllers.HandleCallRecordFunc(container, "getResidentCallRecords"))
		callRoutes.POST("/:id/feedback", controllers.HandleCallRecordFunc(container, "submitCallFeedback"))
	}

	// 紧急情况路由
	emergencyRoutes := api.Group("/emergency")
	emergencyRoutes.Use(middleware.AuthenticatePropertyStaff()) // 物业人员及以上权限可以访问
	{
		// 紧急情况相关的API
		emergencyRoutes.POST("/alarm", controllers.HandleEmergencyFunc(container, "triggerAlarm"))
		emergencyRoutes.GET("/contacts", controllers.HandleEmergencyFunc(container, "getEmergencyContacts"))
		emergencyRoutes.POST("/unlock-all", controllers.HandleEmergencyFunc(container, "emergencyUnlockAll"))
		emergencyRoutes.POST("/notify-all", controllers.HandleEmergencyFunc(container, "notifyAllUsers"))
	}
}
