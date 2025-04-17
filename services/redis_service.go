package services

import (
	"context"
	"encoding/json"
	"ilock-http-service/config"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisService handles Redis operations
type RedisService struct {
	Client *redis.Client
	Ctx    context.Context
}

// NewRedisService creates a new Redis service
func NewRedisService(cfg *config.Config) *RedisService {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: "", // No password set
		DB:       cfg.RedisDB,
	})

	ctx := context.Background()

	return &RedisService{
		Client: client,
		Ctx:    ctx,
	}
}

// Set sets a key-value pair in Redis with expiration
func (s *RedisService) Set(key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.Client.Set(s.Ctx, key, jsonValue, expiration).Err()
}

// Get gets a value from Redis by key
func (s *RedisService) Get(key string, dest interface{}) error {
	val, err := s.Client.Get(s.Ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete deletes a key from Redis
func (s *RedisService) Delete(key string) error {
	return s.Client.Del(s.Ctx, key).Err()
}

// CacheRTCToken caches an RTC token with expiration
func (s *RedisService) CacheRTCToken(userID, channelID, token string, expiration time.Duration) error {
	key := "rtc_token:" + userID + ":" + channelID
	return s.Client.Set(s.Ctx, key, token, expiration).Err()
}

// GetRTCToken gets an RTC token from cache
func (s *RedisService) GetRTCToken(userID, channelID string) (string, error) {
	key := "rtc_token:" + userID + ":" + channelID
	return s.Client.Get(s.Ctx, key).Result()
}

// CacheWeather caches weather data with expiration
func (s *RedisService) CacheWeather(location string, weatherData interface{}, expiration time.Duration) error {
	key := "weather:" + location
	jsonValue, err := json.Marshal(weatherData)
	if err != nil {
		return err
	}

	return s.Client.Set(s.Ctx, key, jsonValue, expiration).Err()
}

// GetWeather gets weather data from cache
func (s *RedisService) GetWeather(location string, dest interface{}) error {
	key := "weather:" + location
	val, err := s.Client.Get(s.Ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}
