package config

import (
	"os"
	"strconv"
	"sync"
)

var (
	config     *Config
	configOnce sync.Once
)

// Config stores all configuration of the application
type Config struct {
	// Database
	DBHost          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBPort          string
	DBMigrationMode string // 数据库迁移模式: "auto"(默认), "alter"(修改), "drop"(删除重建)

	// Server
	ServerPort string

	// Redis
	RedisHost string
	RedisPort string
	RedisDB   int

	// Aliyun RTC
	AliyunAccessKey string
	AliyunRTCAppID  string
	AliyunRTCRegion string

	// JWT Authentication
	JWTSecretKey string

	// Admin
	DefaultAdminPassword string
}

// LoadConfig loads config from environment variables
func LoadConfig() *Config {
	return &Config{
		// Database config
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBUser:          getEnv("DB_USER", "root"),
		DBPassword:      getEnv("DB_PASSWORD", "1090119your"),
		DBName:          getEnv("DB_NAME", "ilock_db"),
		DBPort:          getEnv("DB_PORT", "3308"),
		DBMigrationMode: getEnv("DB_MIGRATION_MODE", "alter"), // 默认为alter模式, 更安全地修改表结构

		// Server config
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// Redis config
		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6380"),
		RedisDB:   getEnvAsInt("REDIS_DB", 0),

		// Aliyun RTC config
		AliyunAccessKey: getEnv("ALIYUN_ACCESS_KEY", "67613a6a74064cad9859c8f794980cae"),
		AliyunRTCAppID:  getEnv("ALIYUN_RTC_APP_ID", "md3fh5x4"),
		AliyunRTCRegion: getEnv("ALIYUN_RTC_REGION", "cn-hangzhou"),

		// JWT Config
		JWTSecretKey: getEnv("JWT_SECRET_KEY", "ilock-secret-key-change-in-production"),

		// Admin Config
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "admin123"),
	}
}

// GetConfig returns the application configuration as a singleton
func GetConfig() *Config {
	configOnce.Do(func() {
		config = LoadConfig()
	})
	return config
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName + "?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true&multiStatements=true"
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to get environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// Helper function to get environment variable as boolean with default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
