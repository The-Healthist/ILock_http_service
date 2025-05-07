package models

import (
	"time"
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
