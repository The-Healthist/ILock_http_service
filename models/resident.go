package models

import (
	"time"

	"gorm.io/gorm"
)

// Resident represents home residents
type Resident struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(50);not null" json:"name"`
	Email     string    `gorm:"type:varchar(100)" json:"email"`
	Phone     string    `gorm:"type:varchar(20);not null" json:"phone"`
	Password  string    `gorm:"type:varchar(100);not null" json:"-"` // 不在JSON中暴露密码
	DeviceID  uint      `gorm:"unique" json:"device_id"`             // One-to-one relationship with Device
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Device        *Device        `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	CallRecords   []CallRecord   `gorm:"foreignKey:ResidentID" json:"call_records,omitempty"`
	AccessLogs    []AccessLog    `gorm:"foreignKey:ResidentID" json:"access_logs,omitempty"`
	EmergencyLogs []EmergencyLog `gorm:"foreignKey:ResidentID" json:"emergency_logs,omitempty"`
}

// BeforeCreate 是一个GORM钩子，在创建新记录前运行
func (r *Resident) BeforeCreate(tx *gorm.DB) error {
	// 如果提供了密码，对其进行哈希处理
	if r.Password != "" {
		hashedPassword, err := HashPassword(r.Password)
		if err != nil {
			return err
		}
		r.Password = hashedPassword
	}
	return nil
}

// BeforeSave 是一个GORM钩子，在保存记录前运行
func (r *Resident) BeforeSave(tx *gorm.DB) error {
	// 如果提供了密码且不是已哈希的，对其进行哈希处理
	if r.Password != "" && len(r.Password) < 60 {
		hashedPassword, err := HashPassword(r.Password)
		if err != nil {
			return err
		}
		r.Password = hashedPassword
	}
	return nil
}
