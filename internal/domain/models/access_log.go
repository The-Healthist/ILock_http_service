package models

import (
	"time"
)

// AccessResult represents the result of an access attempt
type AccessResult string

const (
	AccessResultSuccess AccessResult = "success"
	AccessResultFailure AccessResult = "failure"
)

// AccessMethod represents the method used for access
type AccessMethod string

const (
	AccessMethodRemote      AccessMethod = "remote"
	AccessMethodCode        AccessMethod = "code"
	AccessMethodFace        AccessMethod = "face"
	AccessMethodFingerprint AccessMethod = "fingerprint"
)

// AccessLog represents door access logs
type AccessLog struct {
	ID         uint         `gorm:"primaryKey" json:"id"`
	DeviceID   uint         `json:"device_id"`
	ResidentID uint         `json:"resident_id"`
	Result     AccessResult `gorm:"type:varchar(20)" json:"result"`
	Timestamp  time.Time    `json:"timestamp"`
	Method     AccessMethod `gorm:"type:varchar(20)" json:"method"`

	// Relations
	Device   *Device   `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	Resident *Resident `gorm:"foreignKey:ResidentID" json:"resident,omitempty"`
}
