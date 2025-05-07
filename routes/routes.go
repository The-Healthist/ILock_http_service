package routes

import (
	"ilock-http-service/config"
	"ilock-http-service/controllers"
	_ "ilock-http-service/docs"
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

	// 添加 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:20033")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	// 设置正确的Content-Type，确保UTF-8编码
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})
	// 创建服务容器
	serviceContainer := container.NewServiceContainer(db, cfg, nil)
	// 初始化中间件
	middleware.InitAuthMiddleware(cfg)
	// 添加 Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册路由
	registerRoutes(r, serviceContainer)
	return r
}

// registerRoutes 配置所有API路由
func registerRoutes(
	r *gin.Engine,
	container *container.ServiceContainer,
) {
	// API 路由根路径
	api := r.Group("/api")
	// 注册公共路由
	registerPublicRoutes(api, container)
	// 注册需要认证的路由
	registerAuthenticatedRoutes(api, container)
}

// registerPublicRoutes 注册公共路由
func registerPublicRoutes(
	api *gin.RouterGroup,
	container *container.ServiceContainer,
) {
	// 健康检查
	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 认证路由
	api.POST("/auth/login", controllers.HandleJWTFunc(container, "login"))
	// 阿里云RTC路由
	api.POST("/rtc/token", controllers.HandleRTCFunc(container, "getToken"))
	api.POST("/rtc/call", controllers.HandleRTCFunc(container, "startCall"))
	// 腾讯云RTC路由
	api.POST("/trtc/usersig", controllers.HandleTencentRTCFunc(container, "getUserSig"))
	api.POST("/trtc/call", controllers.HandleTencentRTCFunc(container, "startCall"))
	// MQTT通话和消息路由组
	api.POST("/mqtt/initiate", controllers.HandleMQTTCallFunc(container, "initiateCall"))
	api.POST("/mqtt/caller-action", controllers.HandleMQTTCallFunc(container, "callerAction"))
	api.POST("/mqtt/callee-action", controllers.HandleMQTTCallFunc(container, "calleeAction"))
	api.POST("/mqtt/device/status", controllers.HandleMQTTCallFunc(container, "publishDeviceStatus"))
	api.POST("/mqtt/system/message", controllers.HandleMQTTCallFunc(container, "publishSystemMessage"))

	api.POST("/mqtt/initiate", controllers.HandleMQTTCallFunc(container, "initiateCall"))
	api.POST("/mqtt/caller-action", controllers.HandleMQTTCallFunc(container, "callerAction"))
	api.POST("/mqtt/callee-action", controllers.HandleMQTTCallFunc(container, "calleeAction"))
	api.POST("/mqtt/session", controllers.HandleMQTTCallFunc(container, "getCallSession"))
	api.POST("/mqtt/end-session", controllers.HandleMQTTCallFunc(container, "endCallSession"))
	api.POST("/mqtt/device/status", controllers.HandleMQTTCallFunc(container, "publishDeviceStatus"))
	api.POST("/mqtt/system/message", controllers.HandleMQTTCallFunc(container, "publishSystemMessage"))

}

// registerAuthenticatedRoutes 注册需要认证的路由
func registerAuthenticatedRoutes(
	api *gin.RouterGroup,
	container *container.ServiceContainer,
) {
	// 系统管理员路由组
	adminGroup := api.Group("")
	adminGroup.Use(middleware.AuthenticateSystemAdmin())

	// 管理员管理
	adminGroup.GET("/admins", controllers.HandleAdminFunc(container, "getAdmins"))
	adminGroup.GET("/admins/:id", controllers.HandleAdminFunc(container, "getAdmin"))
	adminGroup.POST("/admins", controllers.HandleAdminFunc(container, "createAdmin"))
	adminGroup.PUT("/admins/:id", controllers.HandleAdminFunc(container, "updateAdmin"))
	adminGroup.DELETE("/admins/:id", controllers.HandleAdminFunc(container, "deleteAdmin"))

	// 物业人员管理
	adminGroup.GET("/staffs", controllers.HandleStaffFunc(container, "getStaffs"))
	adminGroup.GET("/staffs/:id", controllers.HandleStaffFunc(container, "getStaff"))
	adminGroup.POST("/staffs", controllers.HandleStaffFunc(container, "createStaff"))
	adminGroup.PUT("/staffs/:id", controllers.HandleStaffFunc(container, "updateStaff"))
	adminGroup.DELETE("/staffs/:id", controllers.HandleStaffFunc(container, "deleteStaff"))

	// 居民管理
	adminGroup.GET("/residents", controllers.HandleResidentFunc(container, "getResidents"))
	adminGroup.GET("/residents/:id", controllers.HandleResidentFunc(container, "getResident"))
	adminGroup.POST("/residents", controllers.HandleResidentFunc(container, "createResident"))
	adminGroup.PUT("/residents/:id", controllers.HandleResidentFunc(container, "updateResident"))
	adminGroup.DELETE("/residents/:id", controllers.HandleResidentFunc(container, "deleteResident"))

	// 设备管理
	adminGroup.GET("/devices", controllers.HandleDeviceFunc(container, "getDevices"))
	adminGroup.GET("/devices/:id", controllers.HandleDeviceFunc(container, "getDevice"))
	adminGroup.POST("/devices", controllers.HandleDeviceFunc(container, "createDevice"))
	adminGroup.PUT("/devices/:id", controllers.HandleDeviceFunc(container, "updateDevice"))
	adminGroup.DELETE("/devices/:id", controllers.HandleDeviceFunc(container, "deleteDevice"))
	adminGroup.GET("/devices/:id/status", controllers.HandleDeviceFunc(container, "getDeviceStatus"))

	// 通话记录管理
	adminGroup.GET("/call_records", controllers.HandleCallRecordFunc(container, "getCallRecords"))
	adminGroup.GET("/call_records/:id", controllers.HandleCallRecordFunc(container, "getCallRecord"))
	adminGroup.GET("/call_records/statistics", controllers.HandleCallRecordFunc(container, "getCallStatistics"))
	adminGroup.GET("/call_records/device/:deviceId", controllers.HandleCallRecordFunc(container, "getDeviceCallRecords"))
	adminGroup.GET("/call_records/resident/:residentId", controllers.HandleCallRecordFunc(container, "getResidentCallRecords"))
	adminGroup.POST("/call_records/:id/feedback", controllers.HandleCallRecordFunc(container, "submitCallFeedback"))

	// 紧急情况路由
	emergencyRoutes := api.Group("/emergency")
	emergencyRoutes.Use(middleware.AuthenticatePropertyStaff()) // 物业人员及以上权限可以访问
	{
		emergencyRoutes.POST("/alarm", controllers.HandleEmergencyFunc(container, "triggerAlarm"))
		emergencyRoutes.GET("/contacts", controllers.HandleEmergencyFunc(container, "getEmergencyContacts"))
		emergencyRoutes.POST("/unlock-all", controllers.HandleEmergencyFunc(container, "emergencyUnlockAll"))
		emergencyRoutes.POST("/notify-all", controllers.HandleEmergencyFunc(container, "notifyAllUsers"))
	}
}
