package models

import (
	"time"

	"gorm.io/gorm"
)

// PropertyStaff 表示物业员工
type PropertyStaff struct {
	ID           uint      `gorm:"primaryKey;unique" json:"id"`
	Phone        string    `gorm:"type:varchar(20);unique;not null" json:"phone"`
	PropertyName string    `gorm:"type:varchar(100)" json:"property_name"`
	Position     string    `gorm:"type:varchar(50)" json:"position"`
	Role         string    `gorm:"type:varchar(20);not null" json:"role"` // manager, staff, etc.
	Status       string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	Remark       string    `gorm:"type:text" json:"remark"`
	Username     string    `gorm:"type:varchar(50);unique;not null" json:"username"`
	Password     string    `gorm:"type:varchar(100);not null" json:"-"` // Password not exposed in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系 - 使用多对多关系替代直接关联
	Devices []Device `gorm:"many2many:staff_device_relations;" json:"devices,omitempty"` // 通过关系表关联的设备列表
}

// BeforeCreate 是一个GORM钩子，在创建新记录前运行
func (s *PropertyStaff) BeforeCreate(tx *gorm.DB) error {
	// 如果提供了密码，对其进行哈希处理
	if s.Password != "" {
		hashedPassword, err := HashPassword(s.Password)
		if err != nil {
			return err
		}
		s.Password = hashedPassword
	}
	return nil
}

// BeforeSave 是一个GORM钩子，在保存记录前运行
func (s *PropertyStaff) BeforeSave(tx *gorm.DB) error {
	// 如果提供了密码且不是已哈希的，对其进行哈希处理
	if s.Password != "" && len(s.Password) < 60 {
		hashedPassword, err := HashPassword(s.Password)
		if err != nil {
			return err
		}
		s.Password = hashedPassword
	}
	return nil
}
