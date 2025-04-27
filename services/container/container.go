package container

import (
	"context"
	"log"
	"sync"
	"time"

	"ilock-http-service/config"
	"ilock-http-service/services"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// ServiceContainer 管理所有服务的依赖注入
type ServiceContainer struct {
	db     *gorm.DB
	config *config.Config
	redis  *redis.Client

	// 服务
	jwtService   *services.JWTService
	rtcService   *services.RTCService
	redisService *services.RedisService

	// 新增服务
	deviceService     *services.DeviceService
	adminService      *services.AdminService
	residentService   *services.ResidentService
	staffService      *services.StaffService
	callRecordService *services.CallRecordService
	emergencyService  *services.EmergencyService

	mu sync.RWMutex
}

// NewServiceContainer 创建新的服务容器
func NewServiceContainer(db *gorm.DB, cfg *config.Config, redisClient *redis.Client) *ServiceContainer {
	if db == nil {
		panic("数据库连接为空")
	}

	if cfg == nil {
		panic("配置为空")
	}

	// 测试Redis连接
	if redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Printf("Redis连接测试失败: %v，将不使用Redis缓存", err)
		}
	}

	container := &ServiceContainer{
		db:     db,
		config: cfg,
		redis:  redisClient,
	}
	container.initializeServices()
	return container
}

// initializeServices 初始化所有服务
func (c *ServiceContainer) initializeServices() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 初始化服务
	c.jwtService = services.NewJWTService(c.config)
	c.rtcService = services.NewRTCService(c.config)
	c.redisService = services.NewRedisService(c.config)

	// 初始化新增服务
	c.deviceService = services.NewDeviceService(c.db, c.config)
	c.adminService = services.NewAdminService(c.db, c.config)
	c.residentService = services.NewResidentService(c.db, c.config)
	c.staffService = services.NewStaffService(c.db, c.config)
	c.callRecordService = services.NewCallRecordService(c.db, c.config)
	c.emergencyService = services.NewEmergencyService(c.db, c.config)
}

// GetJWTService 获取JWT服务
func (c *ServiceContainer) GetJWTService() *services.JWTService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.jwtService
}

// GetRTCService 获取RTC服务
func (c *ServiceContainer) GetRTCService() *services.RTCService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rtcService
}

// GetRedisService 获取Redis服务
func (c *ServiceContainer) GetRedisService() *services.RedisService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.redisService
}

// GetDeviceService 获取设备服务
func (c *ServiceContainer) GetDeviceService() *services.DeviceService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.deviceService
}

// GetAdminService 获取管理员服务
func (c *ServiceContainer) GetAdminService() *services.AdminService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.adminService
}

// GetResidentService 获取居民服务
func (c *ServiceContainer) GetResidentService() *services.ResidentService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.residentService
}

// GetStaffService 获取物业人员服务
func (c *ServiceContainer) GetStaffService() *services.StaffService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.staffService
}

// GetCallRecordService 获取通话记录服务
func (c *ServiceContainer) GetCallRecordService() *services.CallRecordService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.callRecordService
}

// GetEmergencyService 获取紧急事件服务
func (c *ServiceContainer) GetEmergencyService() *services.EmergencyService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.emergencyService
}

// GetDB 获取数据库连接
func (c *ServiceContainer) GetDB() *gorm.DB {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.db
}

// GetService 基于服务名称获取服务
func (c *ServiceContainer) GetService(name string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	switch name {
	case "jwt":
		return c.jwtService
	case "rtc":
		return c.rtcService
	case "redis":
		return c.redisService
	case "device":
		return c.deviceService
	case "admin":
		return c.adminService
	case "resident":
		return c.residentService
	case "staff":
		return c.staffService
	case "call_record":
		return c.callRecordService
	case "emergency":
		return c.emergencyService
	default:
		return nil
	}
}
