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

	// MQTT通话和消息路由组 - 更新以匹配API文档
	api.POST("/mqtt/call", controllers.HandleMQTTCallFunc(container, "initiateCall"))                // 发起通话，支持可选的户号参数或住户电话
	api.POST("/mqtt/controller/device", controllers.HandleMQTTCallFunc(container, "callerAction"))   // 修改路径从caller-action到controller/device
	api.POST("/mqtt/controller/resident", controllers.HandleMQTTCallFunc(container, "calleeAction")) // 修改路径从callee-action到controller/resident
	api.GET("/mqtt/session", controllers.HandleMQTTCallFunc(container, "getCallSession"))            // 修改为GET请求
	api.POST("/mqtt/end-session", controllers.HandleMQTTCallFunc(container, "endCallSession"))
	api.POST("/mqtt/device/status", controllers.HandleMQTTCallFunc(container, "publishDeviceStatus"))
	api.POST("/mqtt/system/message", controllers.HandleMQTTCallFunc(container, "publishSystemMessage"))

	// 设备健康检测路由
	api.POST("/device/status", controllers.HandleDeviceFunc(container, "checkDeviceHealth"))
}

// registerAuthenticatedRoutes 注册需要认证的路由
func registerAuthenticatedRoutes(
	api *gin.RouterGroup,
	container *container.ServiceContainer,
) {
	// 添加认证中间件
	auth := api.Group("/")
	auth.Use(middleware.AuthenticateSystemAdmin())

	// 管理员路由
	auth.Group("/admin").GET("", controllers.HandleAdminFunc(container, "getAdmins"))
	auth.Group("/admin").GET("/:id", controllers.HandleAdminFunc(container, "getAdmin"))
	auth.Group("/admin").POST("", controllers.HandleAdminFunc(container, "createAdmin"))
	auth.Group("/admin").PUT("/:id", controllers.HandleAdminFunc(container, "updateAdmin"))
	auth.Group("/admin").DELETE("/:id", controllers.HandleAdminFunc(container, "deleteAdmin"))

	// 设备路由
	auth.Group("/devices").GET("", controllers.HandleDeviceFunc(container, "getDevices"))
	auth.Group("/devices").GET("/:id", controllers.HandleDeviceFunc(container, "getDevice"))
	auth.Group("/devices").POST("", controllers.HandleDeviceFunc(container, "createDevice"))
	auth.Group("/devices").PUT("/:id", controllers.HandleDeviceFunc(container, "updateDevice"))
	auth.Group("/devices").DELETE("/:id", controllers.HandleDeviceFunc(container, "deleteDevice"))
	auth.Group("/devices").GET("/:id/status", controllers.HandleDeviceFunc(container, "getDeviceStatus"))
	// 设备与楼号关联
	auth.Group("/devices").POST("/:id/building", controllers.HandleDeviceFunc(container, "associateDeviceWithBuilding"))
	// 设备与户号关联
	auth.Group("/devices").GET("/:id/households", controllers.HandleDeviceFunc(container, "getDeviceHouseholds"))
	auth.Group("/devices").POST("/:id/households", controllers.HandleDeviceFunc(container, "associateDeviceWithHousehold"))
	auth.Group("/devices").DELETE("/:id/households/:household_id", controllers.HandleDeviceFunc(container, "removeDeviceHouseholdAssociation"))

	// 居民路由
	auth.Group("/residents").GET("", controllers.HandleResidentFunc(container, "getResidents"))
	auth.Group("/residents").GET("/:id", controllers.HandleResidentFunc(container, "getResident"))
	auth.Group("/residents").POST("", controllers.HandleResidentFunc(container, "createResident"))
	auth.Group("/residents").PUT("/:id", controllers.HandleResidentFunc(container, "updateResident"))
	auth.Group("/residents").DELETE("/:id", controllers.HandleResidentFunc(container, "deleteResident"))

	// 物业员工路由
	auth.Group("/staffs").GET("", controllers.HandleStaffFunc(container, "getStaff"))
	auth.Group("/staffs").GET("/with-devices", controllers.HandleStaffFunc(container, "getStaffWithDevices"))
	auth.Group("/staffs").GET("/:id", controllers.HandleStaffFunc(container, "getStaffByID"))
	auth.Group("/staffs").POST("", controllers.HandleStaffFunc(container, "createStaff"))
	auth.Group("/staffs").PUT("/:id", controllers.HandleStaffFunc(container, "updateStaff"))
	auth.Group("/staffs").DELETE("/:id", controllers.HandleStaffFunc(container, "deleteStaff"))

	// 通话记录路由
	auth.Group("/call-records").GET("", controllers.HandleCallRecordFunc(container, "getCallRecords"))
	auth.Group("/call-records").GET("/statistics", controllers.HandleCallRecordFunc(container, "getCallStatistics"))
	auth.Group("/call-records").GET("/device/:deviceId", controllers.HandleCallRecordFunc(container, "getDeviceCallRecords"))
	auth.Group("/call-records").GET("/resident/:residentId", controllers.HandleCallRecordFunc(container, "getResidentCallRecords"))
	auth.Group("/call-records").GET("/session", controllers.HandleCallRecordFunc(container, "getCallSession"))
	auth.Group("/call-records").GET("/:id", controllers.HandleCallRecordFunc(container, "getCallRecordByID"))
	auth.Group("/call-records").POST("/:id/feedback", controllers.HandleCallRecordFunc(container, "submitCallFeedback"))

	// 紧急情况路由
	auth.Group("/emergency").GET("", controllers.HandleEmergencyFunc(container, "getEmergencyLogs"))
	auth.Group("/emergency").GET("/:id", controllers.HandleEmergencyFunc(container, "getEmergencyLogByID"))
	auth.Group("/emergency").PUT("/:id", controllers.HandleEmergencyFunc(container, "updateEmergencyLog"))
	auth.Group("/emergency").POST("/trigger", controllers.HandleEmergencyFunc(container, "triggerEmergency"))
	auth.Group("/emergency").POST("/alarm", controllers.HandleEmergencyFunc(container, "triggerAlarm"))
	auth.Group("/emergency").GET("/contacts", controllers.HandleEmergencyFunc(container, "getEmergencyContacts"))
	auth.Group("/emergency").POST("/notify-all", controllers.HandleEmergencyFunc(container, "notifyAllUsers"))
	auth.Group("/emergency").POST("/unlock-all", controllers.HandleEmergencyFunc(container, "emergencyUnlockAll"))

	// 楼号路由
	auth.Group("/buildings").GET("", controllers.HandleBuildingFunc(container, "getBuildings"))
	auth.Group("/buildings").GET("/:id", controllers.HandleBuildingFunc(container, "getBuilding"))
	auth.Group("/buildings").POST("", controllers.HandleBuildingFunc(container, "createBuilding"))
	auth.Group("/buildings").PUT("/:id", controllers.HandleBuildingFunc(container, "updateBuilding"))
	auth.Group("/buildings").DELETE("/:id", controllers.HandleBuildingFunc(container, "deleteBuilding"))
	auth.Group("/buildings").GET("/:id/devices", controllers.HandleBuildingFunc(container, "getBuildingDevices"))
	auth.Group("/buildings").GET("/:id/households", controllers.HandleBuildingFunc(container, "getBuildingHouseholds"))

	// 户号路由
	auth.Group("/households").GET("", controllers.HandleHouseholdFunc(container, "getHouseholds"))
	auth.Group("/households").GET("/:id", controllers.HandleHouseholdFunc(container, "getHousehold"))
	auth.Group("/households").POST("", controllers.HandleHouseholdFunc(container, "createHousehold"))
	auth.Group("/households").PUT("/:id", controllers.HandleHouseholdFunc(container, "updateHousehold"))
	auth.Group("/households").DELETE("/:id", controllers.HandleHouseholdFunc(container, "deleteHousehold"))
	auth.Group("/households").GET("/:id/devices", controllers.HandleHouseholdFunc(container, "getHouseholdDevices"))
	auth.Group("/households").GET("/:id/residents", controllers.HandleHouseholdFunc(container, "getHouseholdResidents"))
	auth.Group("/households").POST("/:id/devices", controllers.HandleHouseholdFunc(container, "associateHouseholdWithDevice"))
	auth.Group("/households").DELETE("/:id/devices/:device_id", controllers.HandleHouseholdFunc(container, "removeHouseholdDeviceAssociation"))
}
