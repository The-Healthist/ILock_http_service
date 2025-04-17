package services

import (
	"errors"
	"ilock-http-service/config"
	"ilock-http-service/models"

	"gorm.io/gorm"
)

// AdminService 提供管理员相关的服务
type AdminService struct {
	DB     *gorm.DB
	Config *config.Config
}

// NewAdminService 创建一个新的管理员服务
func NewAdminService(db *gorm.DB, cfg *config.Config) *AdminService {
	return &AdminService{
		DB:     db,
		Config: cfg,
	}
}

// GetAllAdmins 获取所有管理员，支持分页
func (s *AdminService) GetAllAdmins(page, pageSize int) ([]models.Admin, int64, error) {
	var admins []models.Admin
	var total int64

	// 获取总数
	if err := s.DB.Model(&models.Admin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.DB.Limit(pageSize).Offset(offset).Find(&admins).Error; err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

// GetAdminByID 根据ID获取管理员
func (s *AdminService) GetAdminByID(id uint) (*models.Admin, error) {
	var admin models.Admin
	if err := s.DB.First(&admin, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("管理员不存在")
		}
		return nil, err
	}
	return &admin, nil
}

// CreateAdmin 创建新管理员
func (s *AdminService) CreateAdmin(admin *models.Admin) error {
	// 验证用户名唯一性
	var count int64
	if err := s.DB.Model(&models.Admin{}).Where("username = ?", admin.Username).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("用户名已存在")
	}

	// 设置密码哈希
	hashedPassword, err := models.HashPassword(admin.Password)
	if err != nil {
		return errors.New("密码加密失败")
	}
	admin.Password = hashedPassword

	return s.DB.Create(admin).Error
}

// UpdateAdmin 更新管理员信息
func (s *AdminService) UpdateAdmin(id uint, updates map[string]interface{}) (*models.Admin, error) {
	admin, err := s.GetAdminByID(id)
	if err != nil {
		return nil, err
	}

	// 如果更新用户名，需要检查唯一性
	if username, ok := updates["username"].(string); ok && username != admin.Username {
		var count int64
		if err := s.DB.Model(&models.Admin{}).Where("username = ? AND id != ?", username, id).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("用户名已被其他管理员使用")
		}
	}

	// 如果更新密码，需要进行哈希处理
	if password, ok := updates["password"].(string); ok {
		hashedPassword, err := models.HashPassword(password)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		updates["password"] = hashedPassword
	}

	if err := s.DB.Model(admin).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的管理员信息
	return s.GetAdminByID(id)
}

// DeleteAdmin 删除管理员
func (s *AdminService) DeleteAdmin(id uint) error {
	// 确保系统中至少有一个管理员
	var count int64
	if err := s.DB.Model(&models.Admin{}).Count(&count).Error; err != nil {
		return err
	}
	if count <= 1 {
		return errors.New("系统必须至少有一个管理员，无法删除最后一个管理员")
	}

	admin, err := s.GetAdminByID(id)
	if err != nil {
		return err
	}
	return s.DB.Delete(admin).Error
}
