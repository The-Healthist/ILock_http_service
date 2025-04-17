package models

import (
	"time"
)

// Weather represents weather data for display on devices
type Weather struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DeviceID    uint      `gorm:"unique" json:"device_id"` // One-to-one with device
	Temperature float64   `json:"temperature"`
	WeatherDesc string    `gorm:"type:varchar(50)" json:"weather"` // Sunny, cloudy, rain, etc.
	Humidity    float64   `json:"humidity"`
	Wind        string    `gorm:"type:varchar(50)" json:"wind"`     // Wind description
	Warning     string    `gorm:"type:varchar(100)" json:"warning"` // Weather warnings
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Device *Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
}
