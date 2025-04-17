package controllers

import (
	"ilock-http-service/models"
	"ilock-http-service/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthController 处理身份验证请求
type AuthController struct {
	DB         *gorm.DB
	JWTService *services.JWTService
}

// LoginRequest 表示登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

// LoginResponse 表示登录响应
type LoginResponse struct {
	Code    int         `json:"code" example:"0"`
	Message string      `json:"message" example:"Login successful"`
	Data    interface{} `json:"data"`
}

// LoginData 表示登录成功后返回的数据
type LoginData struct {
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID    uint   `json:"user_id" example:"1"`
	Role      string `json:"role" example:"admin"`
	Username  string `json:"username" example:"admin"`
	CreatedAt string `json:"created_at" example:"2023-01-01T00:00:00Z"`
}

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Code    int         `json:"code" example:"401"`
	Message string      `json:"message" example:"Invalid username or password"`
	Data    interface{} `json:"data"`
}

// NewAuthController 创建一个新的认证控制器
func NewAuthController(db *gorm.DB, jwtService *services.JWTService) *AuthController {
	return &AuthController{
		DB:         db,
		JWTService: jwtService,
	}
}

// Login 处理用户登录
// @Summary      User Login
// @Description  Process user login and return JWT token with different permissions based on user role
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login request parameters"
// @Success      200  {object}  LoginResponse{data=LoginData}  "Success response with token"
// @Failure      400  {object}  ErrorResponse  "Bad request"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"data":    nil,
		})
		return
	}

	// 尝试查找管理员用户
	var admin models.Admin
	if err := c.DB.Where("username = ?", req.Username).First(&admin).Error; err == nil {
		// 比较密码
		if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err == nil {
			// 生成管理员令牌
			token, err := c.JWTService.GenerateToken(admin.ID, "admin", nil, nil)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Failed to generate token",
					"data":    nil,
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "Login successful",
				"data": gin.H{
					"token":      token,
					"user_id":    admin.ID,
					"role":       "admin",
					"username":   admin.Username,
					"created_at": admin.CreatedAt,
				},
			})
			return
		}
	}

	// 尝试查找物业人员
	var staff models.PropertyStaff
	if err := c.DB.Where("username = ?", req.Username).First(&staff).Error; err == nil {
		// 获取密码字段
		var password string

		// 使用原始查询获取所需字段，移除对不存在的property_id的引用
		row := c.DB.Table("property_staffs").
			Select("password").
			Where("id = ?", staff.ID).
			Row()

		if err := row.Scan(&password); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Database error: " + err.Error(),
				"data":    nil,
			})
			return
		}

		// 比较密码
		if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err == nil {
			// 生成物业人员令牌，不再传递propertyID
			token, err := c.JWTService.GenerateToken(staff.ID, "staff", nil, nil)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Failed to generate token",
					"data":    nil,
				})
				return
			}

			// 获取用户名
			var username string
			c.DB.Table("property_staffs").
				Select("username").
				Where("id = ?", staff.ID).
				Row().
				Scan(&username)

			ctx.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "Login successful",
				"data": gin.H{
					"token":      token,
					"user_id":    staff.ID,
					"role":       "staff",
					"username":   username,
					"created_at": staff.CreatedAt,
				},
			})
			return
		}
	}

	// 尝试查找普通居民
	var resident models.Resident
	if err := c.DB.Where("phone = ?", req.Username).First(&resident).Error; err == nil {
		// 获取密码字段
		var password string
		var name string
		var phone string

		// 使用原始查询获取所需字段
		row := c.DB.Table("residents").
			Select("password, name, phone").
			Where("id = ?", resident.ID).
			Row()

		if err := row.Scan(&password, &name, &phone); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Database error: " + err.Error(),
				"data":    nil,
			})
			return
		}

		// 比较密码
		if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err == nil {
			// 生成居民令牌
			token, err := c.JWTService.GenerateToken(resident.ID, "user", nil, nil)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Failed to generate token",
					"data":    nil,
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "Login successful",
				"data": gin.H{
					"token":      token,
					"user_id":    resident.ID,
					"role":       "user",
					"username":   name,
					"phone":      phone,
					"created_at": resident.CreatedAt,
				},
			})
			return
		}
	}

	// 用户名或密码无效
	ctx.JSON(http.StatusUnauthorized, gin.H{
		"code":    401,
		"message": "Invalid username or password",
		"data":    nil,
	})
}
