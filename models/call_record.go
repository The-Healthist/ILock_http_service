package models

import (
	"time"
)

// CallStatus represents the status of a call
type CallStatus string

const (
	CallStatusAnswered CallStatus = "answered"
	CallStatusMissed   CallStatus = "missed"
	CallStatusTimeout  CallStatus = "timeout"
)

// CallRecord represents call records between devices and residents
type CallRecord struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	DeviceID   uint       `json:"device_id"`
	ResidentID uint       `json:"resident_id"`
	CallStatus CallStatus `gorm:"type:varchar(20)" json:"call_status"`
	Timestamp  time.Time  `json:"timestamp"`
	Duration   int        `json:"duration"` // in seconds

	// Relations
	Device   *Device   `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	Resident *Resident `gorm:"foreignKey:ResidentID" json:"resident,omitempty"`
}
