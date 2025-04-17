package models

import (
	"time"
)

// DeviceStatus represents the status of a door access device
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusFault   DeviceStatus = "fault"
)

// Device represents door access devices
type Device struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	Name         string       `gorm:"type:varchar(50);not null" json:"name"`
	SerialNumber string       `gorm:"type:varchar(50);unique;not null" json:"serial_number"`
	Location     string       `gorm:"type:varchar(100)" json:"location"`
	Status       DeviceStatus `gorm:"type:varchar(20);default:'offline'" json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`

	// Relations - 多对多关系
	Staff         []PropertyStaff `gorm:"many2many:staff_device_relations;" json:"staff,omitempty"` // 通过关系表关联的物业人员列表
	Resident      *Resident       `gorm:"foreignKey:DeviceID" json:"resident,omitempty"`
	CallRecords   []CallRecord    `gorm:"foreignKey:DeviceID" json:"call_records,omitempty"`
	AccessLogs    []AccessLog     `gorm:"foreignKey:DeviceID" json:"access_logs,omitempty"`
	EmergencyLogs []EmergencyLog  `gorm:"foreignKey:DeviceID" json:"emergency_logs,omitempty"`
	Weather       *Weather        `gorm:"foreignKey:DeviceID" json:"weather,omitempty"`
}
