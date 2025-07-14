package benchmark

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

// 测试配置
type TestConfig struct {
	BaseURL     string `json:"base_url"`
	AdminUser   string `json:"admin_user"`
	AdminPass   string `json:"admin_pass"`
	Concurrency int    `json:"concurrency"`
	Requests    int    `json:"requests"`
}

// 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 登录响应
type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

var (
	config    TestConfig
	authToken string
)

// TestMain 测试主函数
func TestMain(m *testing.M) {
	// 加载测试配置
	if err := loadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 获取认证令牌
	if err := getAuthToken(); err != nil {
		fmt.Printf("获取认证令牌失败: %v\n", err)
		os.Exit(1)
	}

	// 运行测试
	os.Exit(m.Run())
}

// loadConfig 加载测试配置
func loadConfig() error {
	// 默认配置
	config = TestConfig{
		BaseURL:     "http://localhost:8080/api",
		AdminUser:   "admin",
		AdminPass:   "admin123",
		Concurrency: 10,
		Requests:    100,
	}

	// 尝试从文件加载配置
	data, err := ioutil.ReadFile("test_config.json")
	if err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("解析配置文件失败: %v", err)
		}
	}

	return nil
}

// getAuthToken 获取认证令牌
func getAuthToken() error {
	benchmark := NewAPIBenchmark(config.BaseURL, 1, 1, "")

	loginReq := LoginRequest{
		Username: config.AdminUser,
		Password: config.AdminPass,
	}

	result := benchmark.RunPOST("/auth/login", loginReq)
	if result.FailureCount > 0 {
		return fmt.Errorf("登录失败: %v", result.Errors[0])
	}

	// 解析响应获取令牌
	// 这里简化处理，实际应该从响应体中解析
	authToken = "your_auth_token_here" // 实际项目中应从响应中获取

	return nil
}

// TestDeviceList 测试设备列表接口
func TestDeviceList(t *testing.T) {
	benchmark := NewAPIBenchmark(config.BaseURL, config.Concurrency, config.Requests, authToken)
	result := benchmark.RunGET("/devices")
	result.PrintResult()

	// 验证结果
	if result.FailureCount > 0 {
		t.Errorf("设备列表接口测试失败: 成功率 %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	}
}

// TestDeviceDetail 测试设备详情接口
func TestDeviceDetail(t *testing.T) {
	benchmark := NewAPIBenchmark(config.BaseURL, config.Concurrency, config.Requests, authToken)
	result := benchmark.RunGET("/devices/1") // 假设ID为1的设备存在
	result.PrintResult()

	// 验证结果
	if result.FailureCount > 0 {
		t.Errorf("设备详情接口测试失败: 成功率 %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	}
}

// TestBuildingList 测试楼号列表接口
func TestBuildingList(t *testing.T) {
	benchmark := NewAPIBenchmark(config.BaseURL, config.Concurrency, config.Requests, authToken)
	result := benchmark.RunGET("/buildings")
	result.PrintResult()

	// 验证结果
	if result.FailureCount > 0 {
		t.Errorf("楼号列表接口测试失败: 成功率 %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	}
}

// TestResidentList 测试住户列表接口
func TestResidentList(t *testing.T) {
	benchmark := NewAPIBenchmark(config.BaseURL, config.Concurrency, config.Requests, authToken)
	result := benchmark.RunGET("/residents")
	result.PrintResult()

	// 验证结果
	if result.FailureCount > 0 {
		t.Errorf("住户列表接口测试失败: 成功率 %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	}
}

// TestCallRecordList 测试通话记录列表接口
func TestCallRecordList(t *testing.T) {
	benchmark := NewAPIBenchmark(config.BaseURL, config.Concurrency, config.Requests, authToken)
	result := benchmark.RunGET("/call-records")
	result.PrintResult()

	// 验证结果
	if result.FailureCount > 0 {
		t.Errorf("通话记录列表接口测试失败: 成功率 %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	}
}

// TestMQTTCallInitiate 测试MQTT通话发起接口
func TestMQTTCallInitiate(t *testing.T) {
	benchmark := NewAPIBenchmark(config.BaseURL, config.Concurrency, config.Requests, authToken)

	// 通话请求数据
	callRequest := map[string]interface{}{
		"device_id":    "SN12345678",
		"household_id": 1,
	}

	result := benchmark.RunPOST("/mqtt/call", callRequest)
	result.PrintResult()

	// 验证结果
	if result.FailureCount > 0 {
		t.Errorf("MQTT通话发起接口测试失败: 成功率 %.2f%%", float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	}
}
