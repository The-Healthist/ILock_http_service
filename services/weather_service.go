package services

import (
	"encoding/json"
	"fmt"
	"ilock-http-service/config"
	"io"
	"net/http"
	"time"
)

// WeatherService handles weather data retrieval
type WeatherService struct {
	Config *config.Config
}

// WeatherData represents weather data for frontend/API responses
type WeatherData struct {
	Location    string    `json:"location"`
	Temperature float64   `json:"temperature"`
	Humidity    int       `json:"humidity"`
	Condition   string    `json:"condition"`
	Icon        string    `json:"icon"`
	Warning     string    `json:"warning,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MojiWeatherData 墨迹天气15日数据
type MojiWeatherData struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Place string `json:"place"`
	Data  []struct {
		Week1  string `json:"week1"`  // 星期，如"周二"
		Week2  string `json:"week2"`  // 日历，如"04/15"
		Wea1   string `json:"wea1"`   // 天气1，如"阴"
		Wea2   string `json:"wea2"`   // 天气2，如"阴"
		Wendu1 string `json:"wendu1"` // 最高温度，如"26°"
		Wendu2 string `json:"wendu2"` // 最低温度，如"16°"
		Img1   string `json:"img1"`   // 天气1图标，如"https://h5tq.moji.com/tianqi/assets/images/weather/w2.png"
		Img2   string `json:"img2"`   // 天气2图标，如"https://h5tq.moji.com/tianqi/assets/images/weather/w2.png"
	} `json:"data"`
}

// WeatherAPIResponse represents the response from the weather API
type WeatherAPIResponse struct {
	Location struct {
		Name    string  `json:"name"`
		Region  string  `json:"region"`
		Country string  `json:"country"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
		} `json:"condition"`
		Humidity int    `json:"humidity"`
		Wind     string `json:"wind_kph"`
	} `json:"current"`
	Alerts struct {
		Alert []struct {
			Headline string `json:"headline"`
		} `json:"alert"`
	} `json:"alerts"`
}

// NewWeatherService creates a new weather service
func NewWeatherService(cfg *config.Config) *WeatherService {
	return &WeatherService{
		Config: cfg,
	}
}

// GetWeatherByLocation fetches weather data for a specific location
func (s *WeatherService) GetWeatherByLocation(location string) (*WeatherData, error) {
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s&aqi=no&alerts=yes",
		s.Config.WeatherAPIURL,
		s.Config.WeatherAPIKey,
		location)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status code %d", resp.StatusCode)
	}

	var apiResp WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding weather response: %w", err)
	}

	// No alerts by default
	warning := ""
	if len(apiResp.Alerts.Alert) > 0 {
		warning = apiResp.Alerts.Alert[0].Headline
	}

	// Create weather data for API response
	weatherData := &WeatherData{
		Location:    fmt.Sprintf("%s, %s", apiResp.Location.Name, apiResp.Location.Country),
		Temperature: apiResp.Current.TempC,
		Humidity:    apiResp.Current.Humidity,
		Condition:   apiResp.Current.Condition.Text,
		Icon:        apiResp.Current.Condition.Icon,
		Warning:     warning,
		UpdatedAt:   time.Now(),
	}

	// Here's how you would save to database model if needed
	// weather := &models.Weather{
	//   Temperature: apiResp.Current.TempC,
	//   WeatherDesc: apiResp.Current.Condition.Text,
	//   Humidity:    float64(apiResp.Current.Humidity),
	//   Wind:        fmt.Sprintf("%s kph", apiResp.Current.Wind),
	//   Warning:     warning,
	//   UpdatedAt:   time.Now(),
	// }

	return weatherData, nil
}

// GetMojiWeatherForecasts 根据IP获取墨迹天气15日预报
func (s *WeatherService) GetMojiWeatherForecasts(ipAddress string) (*MojiWeatherData, error) {
	// 首先检查ID和KEY是否已经设置
	if s.Config.MojiWeatherID == "" || s.Config.MojiWeatherKey == "" {
		return nil, fmt.Errorf("墨迹天气API凭证未设置，请检查环境变量MOJI_WEATHER_ID和MOJI_WEATHER_KEY")
	}

	// 打印API凭证值用于调试
	fmt.Printf("使用墨迹天气API凭证: ID=%s, KEY=%s\n", s.Config.MojiWeatherID, s.Config.MojiWeatherKey)

	// 尝试直接使用GET请求，构建完整URL
	directURL := fmt.Sprintf("https://cn.apihz.cn/api/tianqi/tqybmoji15ip.php?id=%s&key=%s",
		s.Config.MojiWeatherID, s.Config.MojiWeatherKey)

	if ipAddress != "" {
		directURL += fmt.Sprintf("&ip=%s", ipAddress)
	}

	fmt.Printf("尝试直接请求URL: %s\n", directURL)

	// 使用简单的GET请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(directURL)
	if err != nil {
		return nil, fmt.Errorf("直接GET请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 打印响应内容
	respString := string(respBody)
	fmt.Printf("墨迹天气API响应: %s\n", respString)

	// 解析JSON响应
	var weatherData MojiWeatherData
	if err := json.Unmarshal(respBody, &weatherData); err != nil {
		return nil, fmt.Errorf("解析JSON响应失败: %w", err)
	}

	// 检查API返回的状态码
	if weatherData.Code != 0 && weatherData.Code != 200 {
		return nil, fmt.Errorf("API返回错误: %s (代码: %d)", weatherData.Msg, weatherData.Code)
	}

	return &weatherData, nil
}

// UpdateDeviceWeather updates weather data for a specific device
func (s *WeatherService) UpdateDeviceWeather(deviceID uint, location string) (*WeatherData, error) {
	weatherData, err := s.GetWeatherByLocation(location)
	if err != nil {
		return nil, err
	}

	// Here you would typically update the device's weather in database
	// but for simplicity we just return the data

	return weatherData, nil
}
