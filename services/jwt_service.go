package services

import (
	"errors"
	"fmt"
	"ilock-http-service/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTService 提供JWT相关服务
type JWTService struct {
	secretKey string
	issuer    string
}

// JWTClaims 定义JWT令牌的声明结构
type JWTClaims struct {
	UserID     uint   `json:"user_id"`
	Role       string `json:"role"`
	PropertyID *uint  `json:"property_id,omitempty"` // 物业ID，用于标识用户所属物业
	DeviceID   *uint  `json:"device_id,omitempty"`
	jwt.RegisteredClaims
}

// NewJWTService 创建一个新的JWT服务
func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		secretKey: cfg.JWTSecretKey,
		issuer:    "ilock-http-service",
	}
}

// GenerateToken 生成JWT令牌
func (s *JWTService) GenerateToken(userID uint, role string, propertyID, deviceID *uint) (string, error) {
	// 令牌有效期为24小时
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &JWTClaims{
		UserID:     userID,
		Role:       role,
		PropertyID: propertyID,
		DeviceID:   deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// ValidateToken 验证JWT令牌
func (s *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})
}

// ExtractClaims 从令牌中提取声明
func (s *JWTService) ExtractClaims(tokenString string) (*JWTClaims, error) {
	token, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 将map claims转换为JWTClaims结构
		jwtClaims := &JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer: claims["iss"].(string),
			},
		}

		// 提取用户ID
		if userID, ok := claims["user_id"].(float64); ok {
			jwtClaims.UserID = uint(userID)
		}

		// 提取角色
		if role, ok := claims["role"].(string); ok {
			jwtClaims.Role = role
		}

		// 提取物业ID（如果存在）
		if propertyID, ok := claims["property_id"].(float64); ok {
			propID := uint(propertyID)
			jwtClaims.PropertyID = &propID
		}

		// 提取设备ID（如果存在）
		if deviceID, ok := claims["device_id"].(float64); ok {
			devID := uint(deviceID)
			jwtClaims.DeviceID = &devID
		}

		return jwtClaims, nil
	}

	return nil, errors.New("invalid token claims")
}
