package controllers

import (
	"fmt"
	"ilock-http-service/models"
	"ilock-http-service/services"
	"ilock-http-service/services/container"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WeatherController 处理天气相关的请求
type WeatherController struct {
	BaseControllerImpl
	DB *gorm.DB
}

// NewWeatherController 创建一个新的天气控制器
func (f *ControllerFactory) NewWeatherController(ctx *gin.Context, db *gorm.DB) *WeatherController {
	return &WeatherController{
		BaseControllerImpl: BaseControllerImpl{
			Container: f.Container,
			Context:   ctx,
		},
		DB: db,
	}
}

// WeatherResponse 表示天气API响应
type WeatherResponse struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Condition   string  `json:"condition"`
	Icon        string  `json:"icon"`
	UpdatedAt   string  `json:"updated_at"`
}

func (c *WeatherController) GetWeather() {
	// 从查询参数获取位置
	location := c.Context.Query("location")
	if location == "" {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "需要位置参数",
			"data":    nil,
		})
		return
	}

	weatherService := c.Container.GetWeatherService()
	redisService := c.Container.GetRedisService()

	// 首先尝试从缓存获取
	var cachedWeather WeatherResponse
	err := redisService.GetWeather(location, &cachedWeather)
	if err == nil {
		// 在缓存中找到
		c.Context.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "成功",
			"data":    cachedWeather,
		})
		return
	}

	// 从服务获取天气数据
	weatherData, err := weatherService.GetWeatherByLocation(location)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取天气数据失败",
			"data":    nil,
		})
		return
	}

	// 格式化响应
	response := WeatherResponse{
		Location:    weatherData.Location,
		Temperature: weatherData.Temperature,
		Humidity:    weatherData.Humidity,
		Condition:   weatherData.Condition,
		Icon:        weatherData.Icon,
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// 缓存1小时
	redisService.CacheWeather(location, response, 1*time.Hour)

	// 保存到数据库
	c.saveWeatherToDB(weatherData)

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    response,
	})
}

func (c *WeatherController) GetDeviceWeather() {
	// 从路径参数获取设备ID
	deviceID := c.Context.Param("deviceId")
	if deviceID == "" {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "需要设备ID",
			"data":    nil,
		})
		return
	}

	// 从数据库获取设备
	var device models.Device
	if err := c.DB.First(&device, "id = ?", deviceID).Error; err != nil {
		c.Context.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "设备未找到",
			"data":    nil,
		})
		return
	}

	// 使用设备位置
	location := device.Location
	if location == "" {
		c.Context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "设备没有位置信息",
			"data":    nil,
		})
		return
	}

	weatherService := c.Container.GetWeatherService()
	redisService := c.Container.GetRedisService()

	// 首先尝试从缓存获取
	var cachedWeather WeatherResponse
	err := redisService.GetWeather(location, &cachedWeather)
	if err == nil {
		// 在缓存中找到
		c.Context.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "成功",
			"data":    cachedWeather,
		})
		return
	}

	// 从服务获取天气数据
	weatherData, err := weatherService.GetWeatherByLocation(location)
	if err != nil {
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取天气数据失败",
			"data":    nil,
		})
		return
	}

	// 格式化响应
	response := WeatherResponse{
		Location:    weatherData.Location,
		Temperature: weatherData.Temperature,
		Humidity:    weatherData.Humidity,
		Condition:   weatherData.Condition,
		Icon:        weatherData.Icon,
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// 缓存1小时
	redisService.CacheWeather(location, response, 1*time.Hour)

	// 保存到数据库
	c.saveWeatherToDB(weatherData)

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    response,
	})
}

func (c *WeatherController) Get15DaysForecast() {
	// 获取IP参数，如果不提供则使用为空，API将自动使用访问者IP
	ip := c.Context.Query("ip")

	// 记录请求信息到日志
	fmt.Printf("开始处理15日天气预报请求，IP参数: %s\n", ip)

	// 获取天气服务
	weatherService := c.Container.GetWeatherService()
	redisService := c.Container.GetRedisService()

	// 尝试从缓存获取数据
	cacheKey := "forecast_15days"
	if ip != "" {
		cacheKey = "forecast_15days:" + ip
	}

	var cachedForecast services.MojiWeatherData
	err := redisService.GetWeather(cacheKey, &cachedForecast)
	if err == nil && len(cachedForecast.Data) > 0 {
		// 从缓存获取到数据
		fmt.Println("从缓存获取天气数据成功")
		c.Context.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "成功",
			"data":    cachedForecast,
		})
		return
	}

	fmt.Println("缓存中无数据，准备从API获取")

	// 从API获取15日天气预报
	forecastData, err := weatherService.GetMojiWeatherForecasts(ip)
	if err != nil {
		// 详细记录错误
		errMsg := fmt.Sprintf("获取15日天气预报失败: %s", err.Error())
		fmt.Println(errMsg) // 打印到控制台

		// 如果错误包含凭证未设置的信息，提供更有针对性的错误消息
		if strings.Contains(err.Error(), "凭证未设置") {
			c.Context.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "系统配置错误: 天气API凭证未正确设置",
				"error":   err.Error(),
			})
			return
		}

		// 返回更友好的错误信息
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取天气预报失败，请稍后再试",
			"error":   err.Error(),
		})
		return
	}

	// 如果数据为空，返回错误
	if forecastData == nil || len(forecastData.Data) == 0 {
		fmt.Println("API返回的数据为空")
		c.Context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "天气API返回的数据为空",
			"data":    nil,
		})
		return
	}

	fmt.Printf("成功获取天气数据，地点: %s, 天数: %d\n", forecastData.Place, len(forecastData.Data))

	// 缓存数据6小时
	redisService.CacheWeather(cacheKey, forecastData, 6*time.Hour)

	c.Context.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    forecastData,
	})
}

// saveWeatherToDB 保存天气数据到数据库
func (c *WeatherController) saveWeatherToDB(data *services.WeatherData) {
	weatherRecord := models.Weather{
		Temperature: data.Temperature,
		Humidity:    float64(data.Humidity),
		WeatherDesc: data.Condition,
		Warning:     "",
		UpdatedAt:   time.Now(),
	}

	c.DB.Create(&weatherRecord)
}

// HandleWeatherFunc 返回一个处理天气请求的Gin处理函数
func HandleWeatherFunc(container *container.ServiceContainer, db *gorm.DB, method string) gin.HandlerFunc {
	factory := NewControllerFactory(container)

	return func(ctx *gin.Context) {
		controller := factory.NewWeatherController(ctx, db)

		switch method {
		case "getWeather":
			controller.GetWeather()
		case "getDeviceWeather":
			controller.GetDeviceWeather()
		case "get15DaysForecast":
			controller.Get15DaysForecast()
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的方法",
				"data":    nil,
			})
		}
	}
}
